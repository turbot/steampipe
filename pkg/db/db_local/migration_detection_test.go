package db_local

// Migration scenario matrix
// -------------------------
// The cross-major migration is driven by exactly two facts on disk: whether an old version directory holds a real
// cluster (db/<old>/data/PG_VERSION present, plus its binaries), and the state of the target version's data directory.
// The status JSON files are write-only reports - they are never consulted for decisions. The full permutation table,
// with where each scenario's behaviour is pinned by an executing test:
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
//
// This file unit-tests the detector itself - findDifferentPgInstallation - which needs no live clusters: it only
// inspects the directory layout. The engine-behaviour rows are pinned by the heavyweight suites named above.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
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
