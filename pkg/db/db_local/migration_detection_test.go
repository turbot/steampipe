package db_local

// Migration scenario matrix
// -------------------------
// The cross-major migration is driven by exactly three facts on disk: whether an old version directory holds a real
// cluster (db/<old>/data/PG_VERSION present, plus its binaries), the state of the target version's data directory, and
// the migration-incomplete marker (db/migration-incomplete.flag), consulted by prepareDb's stale-marker guard: when the
// marker exists but no pending migration was detected, startup refuses rather than booting a half-written draft. The
// status JSON files are write-only reports - they are never consulted for decisions. The full permutation table, with
// where each scenario's behaviour is pinned by an executing test:
//
// | #  | Scenario                                   | Old data dir                   | New data dir              | Decision                                              | Covered by                                                                  |
// |----|--------------------------------------------|--------------------------------|---------------------------|-------------------------------------------------------|-----------------------------------------------------------------------------|
// | 1  | Brand-new workspace / fresh laptop         | no PG_VERSION                  | empty                     | fresh initdb, no migration                            | TestMigrationDetection (this file); acceptance suite (fresh install)         |
// | 2  | Normal 14->18 upgrade                      | full cluster, PG_VERSION       | empty                     | migrate: dump -> restore -> validate -> tanks -> commit | TestMigrationDetection (detect); TestSharedEnginePublicShape_*,             |
// |    |                                            |                                |                           |                                                       | TestCrossMajorMigration, TestDataTankMigration, deletion-gate suite          |
// | 3  | Died during dump                           | full cluster, PG_VERSION       | empty                     | re-run, stale dump discarded                          | TestDataTankStartup_RetryAfterFailureOverwritesStaleDump                     |
// | 4  | Died mid-restore                           | full cluster, PG_VERSION       | cleared and reimported    | re-run                                                | TestDeletionGate_DataTank_Interrupted_ThenReRun; DT-F2/DT-F3 (tier resets);  |
// |    |                                            |                                |                           |                                                       | public shape restores in one transaction (rollback is a PostgreSQL guarantee)|
// | 5  | Died just before commit                    | full cluster, PG_VERSION       | cleared and reimported    | re-run (old data is source of truth until commit)     | TestDeletionGate_DataTank_Interrupted_ThenReRun                              |
// | 6  | Migration failed (pre-flight/restore/validation) | full cluster, PG_VERSION  | cleared on retry          | no commit; STARTUP FAILS with instructions; next start retries | TestCrossMajorMigration outcome cases; TestDeletionGate_Public_* /    |
// |    |                                            |                                |                           |                                                       | _DataTank_* preserve tests; TestDeletionGate_WiredStartup_FailurePreservesOldDir |
// | 7  | Data-tank restore partial (>=1 partition failed) | full cluster, PG_VERSION | cleared on retry          | treated as failure, no commit; STARTUP FAILS          | TestDeletionGate_DataTank_PartialSuccess_PreservesEverything; DT-F4, LE-10   |
// | 8  | Died after commit, during bulk cleanup     | leftovers, NO PG_VERSION       | kept - it is the database | normal PG18 start, old dir no longer a source         | TestMigrationDetection (this file); TestDeletionGate_*_FullSuccess_* (unlink-first contract) |
// | 9  | Steady state after migration               | empty mount point              | normal cluster            | normal start                                          | trivial; acceptance suite                                                    |
// | 10 | Several old version dirs (fossils)         | multiple with PG_VERSION       | empty                     | migrate from highest version older than target with real data | TestMigrationDetection (this file)                                    |
// | 11 | Leftover dir newer than this build         | PG_VERSION, version > target   | any                       | ignored - never migrate down                          | TestMigrationDetection (this file); TestClassifyPgMigration                  |
// | 12 | Minor bump (14.17 -> 14.19)                | full cluster, PG_VERSION       | empty                     | same-major lightweight dump/restore                   | TestClassifyPgMigration; migration.bats (acceptance)                         |
// | 13 | Opt-out: old dir parked, marker present    | no PG_VERSION anywhere         | possibly half-written draft | startup REFUSES with instructions                    | TestPrepareDb_StaleMarkerNoPendingMigration_RefusesStartup (this file)       |
//
// This file unit-tests the detector itself - findDifferentPgInstallation - which needs no live clusters: it only
// inspects the directory layout. The engine-behaviour rows are pinned by the heavyweight suites named above.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	"github.com/turbot/steampipe/v2/pkg/ociinstaller/versionfile"
)

// detectionTestInstallDir points app_specific.InstallDir (the root all filepaths helpers resolve under) at a fresh
// temp dir for the duration of one test, restoring the prior value afterwards.
func detectionTestInstallDir(t *testing.T) string {
	t.Helper()
	prev := app_specific.InstallDir
	dir := t.TempDir()
	app_specific.InstallDir = dir
	t.Cleanup(func() { app_specific.InstallDir = prev })
	return dir
}

// mkVersionDir lays down db/<version>/ with optionally the postgres binary and an initialised-looking data dir
// (data/PG_VERSION). extraDataFiles simulates leftover content without PG_VERSION (the post-commit state).
func mkVersionDir(t *testing.T, installDir, version string, withBinary, withPGVersion bool, extraDataFiles ...string) {
	t.Helper()
	base := filepath.Join(installDir, "db", version)
	if withBinary {
		binDir := filepath.Join(base, "postgres", "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(binDir, "postgres"), []byte("stub"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	dataDir := filepath.Join(base, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}
	if withPGVersion {
		if err := os.WriteFile(filepath.Join(dataDir, "PG_VERSION"), []byte("14\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range extraDataFiles {
		if err := os.WriteFile(filepath.Join(dataDir, f), []byte("leftover"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMigrationDetection(t *testing.T) {
	const target = "18.4.0"
	ctx := context.Background()

	tests := []struct {
		name      string
		setup     func(t *testing.T, installDir string)
		wantFound bool
		wantDir   string // version-dir name expected when found
	}{
		{
			// Scenario 1: nothing on disk but the target itself.
			name:      "fresh install - no old dirs",
			setup:     func(t *testing.T, d string) {},
			wantFound: false,
		},
		{
			// Scenario 2: the normal upgrade source.
			name: "full old cluster detected",
			setup: func(t *testing.T, d string) {
				mkVersionDir(t, d, "14.19.0", true, true)
			},
			wantFound: true,
			wantDir:   "14.19.0",
		},
		{
			// Fresh Pipes pod: PG14 binaries are baked into every image; without data/PG_VERSION the dir is inert.
			name: "binaries-only dir ignored (fresh Pipes pod)",
			setup: func(t *testing.T, d string) {
				mkVersionDir(t, d, "14.19.0", true, false)
			},
			wantFound: false,
		},
		{
			// A data dir with no binaries has nothing to dump with.
			name: "data-without-binaries dir ignored",
			setup: func(t *testing.T, d string) {
				mkVersionDir(t, d, "14.19.0", false, true)
			},
			wantFound: false,
		},
		{
			// Scenario 8: post-commit, PG_VERSION unlinked, bulk cleanup interrupted - leftovers are inert.
			name: "leftovers without PG_VERSION ignored (post-commit crash)",
			setup: func(t *testing.T, d string) {
				mkVersionDir(t, d, "14.19.0", true, false, "base_1234", "pg_wal_stub")
			},
			wantFound: false,
		},
		{
			// Scenario 10: pick the most-recent prior install, not the first alphabetically.
			name: "picks highest version older than target",
			setup: func(t *testing.T, d string) {
				mkVersionDir(t, d, "14.17.0", true, true)
				mkVersionDir(t, d, "14.19.0", true, true)
			},
			wantFound: true,
			wantDir:   "14.19.0",
		},
		{
			// Scenario 10 variant: a fossil with data outranked by version still loses to the live higher one;
			// a higher dir WITHOUT data must not shadow a lower one with data.
			name: "binaries-only higher dir does not shadow real lower cluster",
			setup: func(t *testing.T, d string) {
				mkVersionDir(t, d, "14.17.0", true, true)
				mkVersionDir(t, d, "14.19.0", true, false)
			},
			wantFound: true,
			wantDir:   "14.17.0",
		},
		{
			// Scenario 11: never migrate down from a leftover newer than this build.
			name: "newer-than-target dir ignored",
			setup: func(t *testing.T, d string) {
				mkVersionDir(t, d, "19.0.0", true, true)
			},
			wantFound: false,
		},
		{
			// The target's own dir is never a source, whatever it contains.
			name: "target dir itself skipped",
			setup: func(t *testing.T, d string) {
				mkVersionDir(t, d, target, true, true)
			},
			wantFound: false,
		},
		{
			// Non-semver entries (stray files, scratch dirs) are skipped, not fatal.
			name: "non-semver entries skipped",
			setup: func(t *testing.T, d string) {
				if err := os.MkdirAll(filepath.Join(d, "db", "scratch"), 0755); err != nil {
					t.Fatal(err)
				}
				mkVersionDir(t, d, "14.19.0", true, true)
			},
			wantFound: true,
			wantDir:   "14.19.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			installDir := detectionTestInstallDir(t)
			tc.setup(t, installDir)

			found, location, err := findDifferentPgInstallation(ctx, target)
			if err != nil {
				t.Fatalf("findDifferentPgInstallation: %v", err)
			}
			if found != tc.wantFound {
				t.Fatalf("found = %v, want %v (location %q)", found, tc.wantFound, location)
			}
			if tc.wantFound && filepath.Base(location) != tc.wantDir {
				t.Fatalf("location = %q, want version dir %q", location, tc.wantDir)
			}
		})
	}
}

// corruptOldClusterInstallDir lays down a db/14.19.0 install whose binaries are real (symlinked from the test binary
// root) but whose data dir holds only PG_VERSION - a cluster the detector accepts but the postmaster cannot start.
func corruptOldClusterInstallDir(t *testing.T) {
	t.Helper()
	installDir := detectionTestInstallDir(t)
	base := filepath.Join(installDir, "db", dtPG14Version)
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dtTestRoot(), "db", dtPG14Version, "postgres"), filepath.Join(base, "postgres")); err != nil {
		t.Fatal(err)
	}
	dataDir := filepath.Join(base, "data")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "PG_VERSION"), []byte("14\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Keep the connect-wait short: the postmaster exits immediately on the corrupt data dir, so there is nothing to
	// wait for.
	prevTimeout := viper.Get(pconstants.ArgDatabaseStartTimeout)
	viper.Set(pconstants.ArgDatabaseStartTimeout, 2)
	t.Cleanup(func() { viper.Set(pconstants.ArgDatabaseStartTimeout, prevTimeout) })
}

// TestPrepareBackup_CrossMajor_OldClusterWontStart_FailStops: scenario 6 before the dump even runs. When the old
// cluster exists but will not start (corrupt data dir), a CROSS-major prepareBackup must return the fail-stop sentinel
// - otherwise both install paths would warn-and-continue and the service would start empty on the new major with the
// real data unmigrated.
func TestPrepareBackup_CrossMajor_OldClusterWontStart_FailStops(t *testing.T) {
	skipIfNoBinaries(t)
	corruptOldClusterInstallDir(t)

	_, err := prepareBackup(context.Background(), "18.4.0")
	if err == nil {
		t.Fatal("prepareBackup succeeded against a cluster that cannot start")
	}
	if !errors.Is(err, errCrossMajorDumpFailed) {
		t.Fatalf("cross-major prepare failure missing the fail-stop sentinel: %v", err)
	}
}

// Same failure on a SAME-major (minor) bump keeps the historical warn-and-continue contract: the error comes back
// plain, without the fail-stop sentinel.
func TestPrepareBackup_SameMajor_OldClusterWontStart_NoSentinel(t *testing.T) {
	skipIfNoBinaries(t)
	corruptOldClusterInstallDir(t)

	_, err := prepareBackup(context.Background(), "14.20.0")
	if err == nil {
		t.Fatal("prepareBackup succeeded against a cluster that cannot start")
	}
	if errors.Is(err, errCrossMajorDumpFailed) {
		t.Fatalf("same-major prepare failure must not carry the cross-major fail-stop sentinel: %v", err)
	}
}

// The working dumps never outlive a cross-major attempt: deleteMigrationDumps runs at commit AND before
// fail-stopping. It must remove all three artifacts (backup.bk, public.dump, the data-tank dump dir) and be a clean
// no-op when none exist (early failures never wrote them).
func TestDeleteMigrationDumps(t *testing.T) {
	detectionTestInstallDir(t)

	dbDir := filepaths.EnsureDatabaseDir()
	if err := os.WriteFile(filepath.Join(dbDir, "backup.bk"), []byte("dump"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "public.dump"), []byte("dump"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dbDir, "data-tank"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "data-tank", "toc.dat"), []byte("dump"), 0644); err != nil {
		t.Fatal(err)
	}

	deleteMigrationDumps()

	for _, p := range []string{
		filepath.Join(dbDir, "backup.bk"),
		filepath.Join(dbDir, "public.dump"),
		filepath.Join(dbDir, "data-tank"),
	} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("dump artifact survived deleteMigrationDumps: %s (stat err=%v)", p, err)
		}
	}

	// Idempotent / no-op safe: nothing exists now, must not panic or error-spam.
	deleteMigrationDumps()
}

// Scenario 13: a migration-incomplete marker with NO pending migration (the old directory was parked / never
// mounted) must refuse startup - the current data dir may hold a half-written draft from the unfinished attempt.
// This drives the production entry point (prepareDb) end to end, with the install faked as complete and up to date
// so no download path runs: the version file claims the shipped digests, and the FDW files exist as stubs.
func TestPrepareDb_StaleMarkerNoPendingMigration_RefusesStartup(t *testing.T) {
	installDir := detectionTestInstallDir(t)
	_ = installDir

	// Version file matching the shipped constants -> dbNeedsUpdate and fdwNeedsUpdate are both false.
	vf := versionfile.NewDBVersionFile()
	vf.EmbeddedDB.ImageDigest = constants.PostgresImageDigest
	vf.FdwExtension.Version = constants.FdwVersion
	if err := vf.Save(); err != nil {
		t.Fatalf("save version file: %v", err)
	}

	// FDW present as stubs -> IsFDWInstalled is true, installFDW never runs.
	sqlLoc, controlLoc := filepaths.GetFDWSQLAndControlLocation()
	for _, f := range []string{sqlLoc, controlLoc, filepaths.GetFDWBinaryLocation()} {
		if err := os.MkdirAll(filepath.Dir(f), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(f, []byte("stub"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// No old version dir anywhere (prepareBackup is a no-op) - but the marker from an unfinished attempt remains.
	writeMigrationIncompleteMarker("14.19.0", constants.DatabaseVersion)

	err := prepareDb(context.Background())
	if err == nil {
		t.Fatal("prepareDb started up over a stale migration-incomplete marker with no migration source")
	}
	if !strings.Contains(err.Error(), "never completed") {
		t.Fatalf("expected the stale-marker refusal, got: %v", err)
	}

	// The opt-out instruction must point at the real marker path.
	if !strings.Contains(err.Error(), migrationIncompleteMarkerPath()) {
		t.Fatalf("refusal message does not name the marker path: %v", err)
	}
}
