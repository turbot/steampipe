package db_local

// Startup-wiring test for the data-tank cross-major migration.
//
// exec-6c wires the shared engine into the cross-major startup path: after the
// public-schema migration succeeds (and while the old cluster is still live),
// restoreDBBackup calls migrateDataTankSchemasOnStartup, which detects data-tank
// schemas on the old cluster and migrates them old -> new. The single deletion
// gate (removeOldDataDirOnMigrationSuccess) then removes the old data directory
// only when BOTH the public-schema and the data-tank migrations confirm full
// success.
//
// restoreDBBackup itself pulls in heavy production plumbing (service state file,
// install-dir filepaths, viper-configured connection timeouts), so this suite
// drives the production entry point the cross-major branch calls -
// migrateDataTankSchemasOnStartup - against two real clusters booted by the
// data-tank harness, plus the deletion gate, asserting the wired contract:
//
//   - data-tank schemas present  -> the migration runs and commits
//   - no data-tank schemas       -> a clean no-op (committed=true, no work)
//   - the deletion gate removes the old dir ONLY on combined full success; a
//     data-tank failure preserves the old dir
//
// It reuses the TestDataTankMigration harness (dtInitCluster / dtStartCluster /
// dtClusterRef) and the same pre-placed PG14 + PG18 binaries.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// startupClusters boots a PG14 source + PG18 target over Unix sockets, applies
// srcSQL to the source, and returns the two engine cluster handles plus the raw
// harness cluster handles (so a test can stop one mid-flight to induce a real
// failure), the source data dir, and a writable backup dir. It mirrors
// runPublicShapeEngine's short-socket-path discipline.
func startupClusters(t *testing.T, srcSQL string) (old, new *pgClusterRef, srcCluster, targetCluster *dtCluster, oldDataDir, backupDir string) {
	t.Helper()
	ctx := context.Background()

	base := filepath.Join(dtTestRoot(), "wire", shortTestKey(t.Name()))
	if err := os.RemoveAll(base); err != nil {
		t.Fatalf("clean base: %v", err)
	}
	pg14Data := filepath.Join(base, "pg14-data")
	pg18Data := filepath.Join(base, "pg18-data")
	pg14Sock := filepath.Join(base, "s14")
	pg18Sock := filepath.Join(base, "s18")
	backupDir = filepath.Join(base, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("mkdir backups: %v", err)
	}

	if err := dtInitCluster(ctx, dtPG14Version, pg14Data); err != nil {
		t.Fatalf("init pg14: %v", err)
	}
	src, err := dtStartCluster(ctx, dtPG14Version, pg14Data, pg14Sock)
	if err != nil {
		t.Fatalf("start pg14: %v", err)
	}
	t.Cleanup(src.stop)
	if err := src.ensureFixtureDB(ctx); err != nil {
		t.Fatalf("ensure pg14 db: %v", err)
	}
	if srcSQL != "" {
		if err := src.applyFixtureSQL(ctx, srcSQL); err != nil {
			t.Fatalf("apply fixture: %v", err)
		}
	}

	if err := dtInitCluster(ctx, dtPG18Version, pg18Data); err != nil {
		t.Fatalf("init pg18: %v", err)
	}
	target, err := dtStartCluster(ctx, dtPG18Version, pg18Data, pg18Sock)
	if err != nil {
		t.Fatalf("start pg18: %v", err)
	}
	t.Cleanup(target.stop)
	if err := target.ensureFixtureDB(ctx); err != nil {
		t.Fatalf("ensure pg18 db: %v", err)
	}

	return dtClusterRef(src), dtClusterRef(target), src, target, pg14Data, backupDir
}

// TestDataTankStartup_MigratesWhenSchemasPresent: a cross-major start whose old
// cluster carries a data tank runs the data-tank migration through the shared
// engine and commits. This is the wired path the production cross-major success
// branch takes.
func TestDataTankStartup_MigratesWhenSchemasPresent(t *testing.T) {
	skipIfNoBinaries(t)

	srcSQL, err := dtReadFixture("DT-A2_list_partition_4.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	old, newRef, _, _, oldDataDir, backupDir := startupClusters(t, srcSQL)

	res, err := migrateDataTankSchemasOnStartup(context.Background(), old, newRef, backupDir)
	if err != nil {
		t.Fatalf("data-tank startup migration returned error: %v", err)
	}
	if !res.committed {
		t.Errorf("expected committed=true for a clean data-tank migration, got %+v", res)
	}
	if res.tierReached != dtRestoreTier1Parallel {
		t.Errorf("expected a clean tank to restore at tier 1, got tier %v", res.tierReached)
	}

	// On full success the deletion gate is unlocked. Confirm the gate removes the
	// old dir only when committed; here it is, so it removes.
	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)
	if _, statErr := os.Stat(oldDataDir); !os.IsNotExist(statErr) {
		t.Errorf("expected old data dir removed after combined full success, stat err=%v", statErr)
	}
}

// TestDataTankStartup_NoOpWhenNoSchemas: the normal CLI workspace has no
// data-tank schemas. The wired call must be a clean no-op - committed=true, no
// error, no dump written - so it never blocks the deletion gate or slows
// startup. (Only the public schema is present here, which the data-tank
// enumeration excludes.)
func TestDataTankStartup_NoOpWhenNoSchemas(t *testing.T) {
	skipIfNoBinaries(t)

	const publicOnlySQL = `
create table public.things (id int primary key, name text);
insert into public.things values (1, 'alpha'), (2, 'bravo');`

	old, newRef, _, _, oldDataDir, backupDir := startupClusters(t, publicOnlySQL)

	res, err := migrateDataTankSchemasOnStartup(context.Background(), old, newRef, backupDir)
	if err != nil {
		t.Fatalf("no-op data-tank startup returned error: %v", err)
	}
	if !res.committed {
		t.Errorf("expected committed=true no-op when no data-tank schemas exist, got %+v", res)
	}
	if res.tierReached != dtRestoreTierNone {
		t.Errorf("expected no restore tier reached for a no-op, got %v", res.tierReached)
	}
	// No data-tank dump should have been written.
	if _, statErr := os.Stat(filepath.Join(backupDir, "data-tank")); !os.IsNotExist(statErr) {
		t.Errorf("expected no data-tank dump dir for a no-op, but one exists (stat err=%v)", statErr)
	}

	// The combined gate still removes the old dir on a public-success +
	// data-tank-no-op start.
	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)
	if _, statErr := os.Stat(oldDataDir); !os.IsNotExist(statErr) {
		t.Errorf("expected old data dir removed after public success + data-tank no-op, stat err=%v", statErr)
	}
}

// TestDataTankStartup_FailurePreservesOldDir: when the data-tank migration does
// NOT commit, the deletion gate must keep the old data directory (the public
// success is not reverted, but the old dir is the preserved recovery copy).
//
// This drives a REAL failure through the production entry point rather than
// hand-setting committed=false: a data-tank schema is present on the old cluster,
// but the target cluster is stopped before the migration runs, so every restore
// tier fails and migrateDataTankSchemasOnStartup returns committed=false. The
// test then feeds that actual result into the gate, exactly as restoreDBBackup's
// cross-major success branch does (dtCommitted = res.committed; gate(dtCommitted,
// location)). It proves the wired contract end-to-end: a real data-tank failure
// produces committed=false, and committed=false preserves the old dir.
func TestDataTankStartup_FailurePreservesOldDir(t *testing.T) {
	skipIfNoBinaries(t)

	srcSQL, err := dtReadFixture("DT-A2_list_partition_4.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	old, newRef, _, target, oldDataDir, backupDir := startupClusters(t, srcSQL)

	// Make the target unreachable so the restore cannot land. Every tier fails;
	// the engine reports committed=false and preserves the source (old dir + the
	// retained data-tank dump) - the 2026-06-08 data-preservation invariant.
	target.stop()

	res, mErr := migrateDataTankSchemasOnStartup(context.Background(), old, newRef, backupDir)
	if res.committed {
		t.Fatalf("expected committed=false when the target is unreachable, got committed=true (err=%v, res=%+v)", mErr, res)
	}

	// Feed the actual migration result into the gate exactly as restoreDBBackup
	// does on the cross-major success branch.
	dtCommitted := res.committed
	removeOldDataDirOnMigrationSuccess(dtCommitted, oldDataDir)

	if err := dtAssertOldDataDirPreserved(oldDataDir); err != nil {
		t.Errorf("data-preservation invariant violated: a real data-tank failure must keep the old dir: %v", err)
	}
}
