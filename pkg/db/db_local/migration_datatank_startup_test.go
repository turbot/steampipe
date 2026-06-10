package db_local

// Startup-wiring test for the data-tank cross-major migration.
//
// exec-6c wires the shared engine into the cross-major startup path: after the public-schema migration succeeds (and
// while the old cluster is still live), restoreDBBackup calls migrateDataTankSchemasOnStartup, which detects data-tank
// schemas on the old cluster and migrates them old -> new. The single deletion gate
// (removeOldDataDirOnMigrationSuccess) then removes the old data directory only when BOTH the public-schema and the
// data-tank migrations confirm full success.
//
// restoreDBBackup itself pulls in heavy production plumbing (service state file, install-dir filepaths,
// viper-configured connection timeouts), so this suite drives the production entry point the cross-major branch calls -
// migrateDataTankSchemasOnStartup - against two real clusters booted by the data-tank harness, plus the deletion gate,
// asserting the wired contract:
//
//   - data-tank schemas present  -> the migration runs and commits
//   - no data-tank schemas       -> a clean no-op (committed=true, no work)
//   - the deletion gate removes the old dir ONLY on combined full success; a data-tank failure preserves the old dir
//
// It reuses the TestDataTankMigration harness (dtInitCluster / dtStartCluster / dtClusterRef) and the same pre-placed
// PG14 + PG18 binaries.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// startupClusters boots a PG14 source + PG18 target over Unix sockets, applies srcSQL to the source, and returns the
// two engine cluster handles plus the raw harness cluster handles (so a test can stop one mid-flight to induce a real
// failure), the source data dir, and a writable backup dir. It mirrors runPublicShapeEngine's short-socket-path
// discipline.
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

// TestDataTankStartup_MigratesWhenSchemasPresent: a cross-major start whose old cluster carries a data tank runs the
// data-tank migration through the shared engine and commits. This is the wired path the production cross-major success
// branch takes.
func TestDataTankStartup_MigratesWhenSchemasPresent(t *testing.T) {
	skipIfNoBinaries(t)

	srcSQL, err := dtReadFixture("DT-A2_list_partition_4.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	old, newRef, srcCluster, _, oldDataDir, backupDir := startupClusters(t, srcSQL)

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

	// On full success the deletion gate is unlocked. Production order: the retained old server stops BEFORE the gate
	// removes its data dir (deleting under a live postmaster forces an abnormal shutdown that leaks its SysV shared
	// memory segment).
	srcCluster.stop()
	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)
	assertOldDataDirCleared(t, oldDataDir)
}

// TestDataTankStartup_NoOpWhenNoSchemas: the normal CLI workspace has no data-tank schemas. The wired call must be a
// clean no-op - committed=true, no error, no dump written - so it never blocks the deletion gate or slows startup.
// (Only the public schema is present here, which the data-tank enumeration excludes.)
// The no-op must still write the status file with committed:true, overwriting any stale committed:false a prior
// attempt's early-error write left behind - the status JSON is the orchestrator's only view of the outcome.
func TestDataTankStartup_NoOpWhenNoSchemas(t *testing.T) {
	skipIfNoBinaries(t)

	const publicOnlySQL = `
create table public.things (id int primary key, name text);
insert into public.things values (1, 'alpha'), (2, 'bravo');`

	old, newRef, srcCluster, _, oldDataDir, backupDir := startupClusters(t, publicOnlySQL)

	// Seed a stale failure status from a hypothetical earlier attempt; the no-op must overwrite it.
	statusPath := filepath.Join(backupDir, "data-tank-migration-status.json")
	if err := os.WriteFile(statusPath, []byte(`{"committed":false,"detail":"stale prior failure"}`), 0644); err != nil {
		t.Fatalf("seed stale status: %v", err)
	}

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
	// The stale committed:false must have been replaced by the no-op's committed:true.
	statusBytes, readErr := os.ReadFile(statusPath)
	if readErr != nil {
		t.Fatalf("read status file after no-op: %v", readErr)
	}
	var status struct {
		Committed bool `json:"committed"`
	}
	if jsonErr := json.Unmarshal(statusBytes, &status); jsonErr != nil {
		t.Fatalf("unmarshal status file: %v", jsonErr)
	}
	if !status.Committed {
		t.Errorf("no-op left status file committed=false; stale prior status not overwritten: %s", statusBytes)
	}

	// The combined gate still removes the old dir on a public-success + data-tank-no-op start. Production order: stop
	// the retained old server before the gate removes its data dir.
	srcCluster.stop()
	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)
	assertOldDataDirCleared(t, oldDataDir)
}

// TestDataTankStartup_RetryAfterFailureOverwritesStaleDump: a FAILED data-tank migration leaves its retained dump dir
// behind (it is a recovery copy), and the old data dir is preserved, so the next startup retries the migration.
// pg_dump's directory format refuses an existing directory, so without clearing the stale dump first every retry would
// fail at the dump step with "File exists" - a permanent fail-loop on the failure path (observed live 2026-06-09). The
// retry must replace the stale dump and commit.
func TestDataTankStartup_RetryAfterFailureOverwritesStaleDump(t *testing.T) {
	skipIfNoBinaries(t)

	srcSQL, err := dtReadFixture("DT-A2_list_partition_4.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	old, newRef, srcCluster, _, oldDataDir, backupDir := startupClusters(t, srcSQL)

	// Simulate the prior failed attempt's leftover: a non-empty retained dump dir at the exact path the engine dumps
	// to.
	staleDump := filepath.Join(backupDir, "data-tank")
	if err := os.MkdirAll(staleDump, 0755); err != nil {
		t.Fatalf("plant stale dump dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staleDump, "toc.dat"), []byte("stale prior attempt"), 0644); err != nil {
		t.Fatalf("plant stale toc: %v", err)
	}

	res, err := migrateDataTankSchemasOnStartup(context.Background(), old, newRef, backupDir)
	if err != nil {
		t.Fatalf("retry over a stale dump dir must succeed, got error: %v (res=%+v)", err, res)
	}
	if !res.committed {
		t.Errorf("retry over a stale dump dir must commit, got %+v", res)
	}

	// The stale marker is gone - the dump was replaced, not appended to.
	if b, rerr := os.ReadFile(filepath.Join(staleDump, "toc.dat")); rerr == nil && string(b) == "stale prior attempt" {
		t.Errorf("stale dump content survived the retry; the dump dir was not replaced")
	}

	// Production order: stop the retained old server before the gate removes its data dir.
	srcCluster.stop()
	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)
	assertOldDataDirCleared(t, oldDataDir)
}

// TestDataTankStartup_FailurePreservesOldDir: when the data-tank migration does NOT commit, the deletion gate must keep
// the old data directory (the public success is not reverted, but the old dir is the preserved recovery copy).
//
// This drives a REAL failure through the production entry point rather than hand-setting committed=false: a data-tank
// schema is present on the old cluster, but the target cluster is stopped before the migration runs, so every restore
// tier fails and migrateDataTankSchemasOnStartup returns committed=false. The test then feeds that actual result into
// the gate, exactly as restoreDBBackup's cross-major success branch does (dtCommitted = res.committed;
// gate(dtCommitted, location)). It proves the wired contract end-to-end: a real data-tank failure produces
// committed=false, and committed=false preserves the old dir.
func TestDataTankStartup_FailurePreservesOldDir(t *testing.T) {
	skipIfNoBinaries(t)

	srcSQL, err := dtReadFixture("DT-A2_list_partition_4.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	old, newRef, _, target, oldDataDir, backupDir := startupClusters(t, srcSQL)

	// Make the target unreachable so the restore cannot land. Every tier fails; the engine reports committed=false and
	// preserves the source (old dir + the retained data-tank dump) - the 2026-06-08 data-preservation invariant.
	target.stop()

	res, mErr := migrateDataTankSchemasOnStartup(context.Background(), old, newRef, backupDir)
	if res.committed {
		t.Fatalf("expected committed=false when the target is unreachable, got committed=true (err=%v, res=%+v)", mErr, res)
	}

	// Feed the actual migration result into the gate exactly as restoreDBBackup does on the cross-major success branch.
	dtCommitted := res.committed
	removeOldDataDirOnMigrationSuccess(dtCommitted, oldDataDir)

	if err := dtAssertOldDataDirPreserved(oldDataDir); err != nil {
		t.Errorf("data-preservation invariant violated: a real data-tank failure must keep the old dir: %v", err)
	}
}
