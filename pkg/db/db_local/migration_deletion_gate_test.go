package db_local

// Deletion-gate guarantee test (exec-6d).
//
// The whole data-safety design reduces to one rule (the 2026-06-08 governing
// decision): the old data directory is NEVER deleted until the migration is
// confirmed 100% complete; on any failure or partial result the original is
// preserved on disk in two independent forms - the untouched old PG14 data
// directory AND the retained safety dump - and the new Postgres version still
// runs (no version-revert).
//
// exec-6a asserts per-breakage OUTCOMES. This suite asserts the INVARIANT itself,
// exhaustively over outcome categories, for both data shapes:
//
//   | Outcome category            | Old data dir            | Safety dump |
//   | full success                | removed (gate fires)    | retained    |
//   | restore failed (all tiers)  | present + populated     | retained    |
//   | partial success (some parts)| present + populated     | retained    |
//   | pre-check aborted           | present + populated     | retained    |
//   | interrupted mid-restore     | present + populated     | retained    |
//   |   then re-run               | survives the re-run                   |
//
// For every NON-full-success outcome the test does not merely stat the path: it
// re-boots the preserved old data directory as a live PostgreSQL cluster on a
// fresh socket (after the source used during the migration has been stopped) and
// QUERIES THE SOURCE ROWS BACK. That proves the preserved directory is a fully
// usable cluster still holding the original data, which is the property the
// governing decision actually guarantees - not just that some files remain on
// disk.
//
// It reuses the TestDataTankMigration harness (dtInitCluster / dtStartCluster /
// dtClusterRef) and the same pre-placed PG14 + PG18 binaries.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// gateClusters boots a PG14 source + PG18 target over Unix sockets, applies
// srcSQL to the source, and returns the engine cluster handles, the raw harness
// handles (so a test can stop the source and re-open its data dir, or stop the
// target to force a real restore failure), the source data dir, and the backup
// dir. It mirrors runPublicShapeEngine's short-socket-path discipline.
func gateClusters(t *testing.T, srcSQL string) (src, target *pgClusterRef, srcCluster, targetCluster *dtCluster, oldDataDir, backupDir string) {
	t.Helper()
	ctx := context.Background()

	base := filepath.Join(dtTestRoot(), "gate", shortTestKey(t.Name()))
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
	srcC, err := dtStartCluster(ctx, dtPG14Version, pg14Data, pg14Sock)
	if err != nil {
		t.Fatalf("start pg14: %v", err)
	}
	t.Cleanup(srcC.stop)
	if err := srcC.ensureFixtureDB(ctx); err != nil {
		t.Fatalf("ensure pg14 db: %v", err)
	}
	if srcSQL != "" {
		if err := srcC.applyFixtureSQL(ctx, srcSQL); err != nil {
			t.Fatalf("apply fixture: %v", err)
		}
	}

	if err := dtInitCluster(ctx, dtPG18Version, pg18Data); err != nil {
		t.Fatalf("init pg18: %v", err)
	}
	targetC, err := dtStartCluster(ctx, dtPG18Version, pg18Data, pg18Sock)
	if err != nil {
		t.Fatalf("start pg18: %v", err)
	}
	t.Cleanup(targetC.stop)
	if err := targetC.ensureFixtureDB(ctx); err != nil {
		t.Fatalf("ensure pg18 db: %v", err)
	}

	return dtClusterRef(srcC), dtClusterRef(targetC), srcC, targetC, pg14Data, backupDir
}

// reopenAndCountRows re-boots a (preserved) data directory as a live cluster on a
// fresh socket and runs countSQL, returning the scalar bigint it yields. This is
// the strong form of the data-preservation assertion: it proves the preserved
// old data directory is not an empty husk but a working cluster that still
// returns the source rows. The caller must have stopped any other server that
// was attached to dataDir first (two postmasters cannot share one data dir).
func reopenAndCountRows(t *testing.T, dataDir, countSQL string) int64 {
	t.Helper()
	ctx := context.Background()

	sock := filepath.Join(filepath.Dir(dataDir), "reopen-"+shortTestKey(dataDir))
	c, err := dtStartCluster(ctx, dtPG14Version, dataDir, sock)
	if err != nil {
		t.Fatalf("re-open preserved old data dir %s: %v", dataDir, err)
	}
	defer c.stop()

	conn, err := c.connect(ctx, dtFixtureDBName)
	if err != nil {
		t.Fatalf("connect to re-opened preserved cluster: %v", err)
	}
	defer conn.Close(ctx)

	var n int64
	if err := conn.QueryRow(ctx, countSQL).Scan(&n); err != nil {
		t.Fatalf("query preserved source data (%s): %v", countSQL, err)
	}
	return n
}

// dataTankRowCountSQL counts the rows in the DT-F base fixture's partitioned
// parent table. The fixture loads 4 partitions x 25 rows = 100 rows.
const dataTankRowCountSQL = `SELECT count(*) FROM "fast_aws"."aws_resource"`

// publicRowCountSQL counts the rows in the public-shape fixture's table.
const publicRowCountSQL = `SELECT count(*) FROM public.things`

// assertOldDataDirCleared asserts the migration commit fired on success: the old
// data dir's PG_VERSION is gone (the atomic commit the startup trigger keys on)
// and the directory's contents are cleared, but the directory itself is preserved
// (in Pipes it is a mounted volume - we delete the contents, not the dir; Victor:
// "delete the contents"). After this the old dir no longer reads as a migratable
// cluster and cannot be booted.
func assertOldDataDirCleared(t *testing.T, oldDataDir string) {
	t.Helper()
	// The directory itself must still exist - we delete contents, not the dir.
	if _, err := os.Stat(oldDataDir); err != nil {
		t.Fatalf("commit removed the data directory itself (must keep the dir, clear its contents): %v", err)
	}
	// PG_VERSION gone = the commit signal; without it the dir is not a cluster.
	if _, err := os.Stat(filepath.Join(oldDataDir, "PG_VERSION")); !os.IsNotExist(err) {
		t.Errorf("commit did not remove PG_VERSION from %s (stat err=%v); old dir still reads as a migratable cluster", oldDataDir, err)
	}
	// Contents must be cleared.
	entries, err := os.ReadDir(oldDataDir)
	if err != nil {
		t.Fatalf("read old data dir after commit: %v", err)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("commit left %d entries in the old data dir (must be cleared): %v", len(entries), names)
	}
}

// assertPreservedOldDirHasRows stops the migration's source server, re-opens the
// preserved old data directory, and asserts it still holds wantRows source rows.
// It is the per-category data-preservation guarantee: present + populated +
// queryable, not just a path that exists.
func assertPreservedOldDirHasRows(t *testing.T, srcCluster *dtCluster, oldDataDir, countSQL string, wantRows int64) {
	t.Helper()
	// Path + PG_VERSION first (cheap structural check), then the live query.
	if err := dtAssertOldDataDirPreserved(oldDataDir); err != nil {
		t.Fatalf("data-preservation invariant: %v", err)
	}
	// Free the data dir so it can be re-opened by a fresh postmaster.
	srcCluster.stop()
	got := reopenAndCountRows(t, oldDataDir, countSQL)
	if got != wantRows {
		t.Errorf("preserved old data dir lost rows: queried %d, want %d source rows still present", got, wantRows)
	}
}

// -----------------------------------------------------------------------------
// Data-tank shape.
// -----------------------------------------------------------------------------

// TestDeletionGate_DataTank_RestoreFailed_PreservesEverything: every restore tier
// fails (failAllTiers). The engine reports DataPreservedOnDisk; the gate must NOT
// remove the old dir, and the source rows must still be queryable from the
// preserved directory, with the safety dump retained.
func TestDeletionGate_DataTank_RestoreFailed_PreservesEverything(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	srcSQL, err := dtReadFixture("DT-F_base.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	_, _, srcCluster, targetCluster, oldDataDir, backupDir := gateClusters(t, srcSQL)

	got, _, mErr := runDataTankMigration(ctx, srcCluster, targetCluster, backupDir, dtCaseSetup{failAllTiers: true})
	if mErr != nil {
		t.Fatalf("migration returned unexpected error: %v", mErr)
	}
	if got != dtOutcomeDataPreservedOnDisk {
		t.Fatalf("expected DataPreservedOnDisk when every tier fails, got %s", got)
	}

	res := migrationResult{committed: false}
	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)

	// Safety dump must be retained.
	if err := dtAssertDumpDirRetained(filepath.Join(backupDir, "data-tank")); err != nil {
		t.Errorf("safety dump not retained after all-tiers-failed: %v", err)
	}
	// Old dir present + populated + queryable (100 source rows survive).
	assertPreservedOldDirHasRows(t, srcCluster, oldDataDir, dataTankRowCountSQL, 100)
}

// TestDeletionGate_DataTank_PreCheckAborted_PreservesEverything: the disk
// pre-flight aborts before any restore. Old dir preserved + queryable; the gate
// does not fire. (Pre-check aborts before the dump, so only the old dir is the
// recovery copy here - matching the engine, which returns before shape.dumpFn.)
func TestDeletionGate_DataTank_PreCheckAborted_PreservesEverything(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	srcSQL, err := dtReadFixture("DT-F_base.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	_, _, srcCluster, targetCluster, oldDataDir, backupDir := gateClusters(t, srcSQL)

	got, _, mErr := runDataTankMigration(ctx, srcCluster, targetCluster, backupDir, dtCaseSetup{forceDiskPreflightFail: true})
	if mErr != nil {
		t.Fatalf("migration returned unexpected error: %v", mErr)
	}
	if got != dtOutcomeDiskPreflightFailed {
		t.Fatalf("expected DiskPreflightFailed, got %s", got)
	}

	removeOldDataDirOnMigrationSuccess(false, oldDataDir)
	assertPreservedOldDirHasRows(t, srcCluster, oldDataDir, dataTankRowCountSQL, 100)
}

// TestDeletionGate_DataTank_DumpFailed_PreservesOldDir: the insurance dump fails.
// Old dir preserved + queryable. (No completed dump exists by definition, so the
// old dir is the surviving recovery copy.)
func TestDeletionGate_DataTank_DumpFailed_PreservesOldDir(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	srcSQL, err := dtReadFixture("DT-F_base.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	_, _, srcCluster, targetCluster, oldDataDir, backupDir := gateClusters(t, srcSQL)

	got, _, mErr := runDataTankMigration(ctx, srcCluster, targetCluster, backupDir, dtCaseSetup{forceDumpFailure: true})
	if mErr != nil {
		t.Fatalf("migration returned unexpected error: %v", mErr)
	}
	if got != dtOutcomeDumpFailed {
		t.Fatalf("expected DumpFailed, got %s", got)
	}

	removeOldDataDirOnMigrationSuccess(false, oldDataDir)
	assertPreservedOldDirHasRows(t, srcCluster, oldDataDir, dataTankRowCountSQL, 100)
}

// TestDeletionGate_DataTank_PartialSuccess_PreservesEverything: a partial restore
// (one partition unrestorable; the rest land at tier 4). The governing decision:
// "Partial success is safe too: any table/partition that didn't migrate is still
// in the preserved original." A degraded restore is NOT 100% complete, so the
// old data directory MUST be preserved (and the safety dump retained) so the
// failed partition's rows remain recoverable.
//
// This asserts the DESIRED behaviour. If the shipped engine commits a degraded
// tier-4 result (res.committed=true) and the gate then removes the old dir, this
// test fails - and that failure is the spec, not a reason to weaken the
// assertion.
func TestDeletionGate_DataTank_PartialSuccess_PreservesEverything(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	srcSQL, err := dtReadFixture("DT-F4_partition_corrupt_base.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	_, _, srcCluster, targetCluster, oldDataDir, backupDir := gateClusters(t, srcSQL)

	// Drive the engine directly so we can read the raw result (committed +
	// partitionFailures) the production gate keys on, not just the suite enum.
	// Force tiers 1-3 to fail and corrupt exactly one partition so tier 4 lands a
	// degraded (partial) result.
	res, mErr := migrateDataTank(ctx, dtClusterRef(srcCluster), dtClusterRef(targetCluster),
		backupDir, filepath.Join(backupDir, "data-tank-migration-status.json"), dtParallelism(), migrationFaults{
			failTier1: true, failTier2: true, failTier3: true, corruptOnePartition: true,
		})
	// A partial restore is signalled by the distinct errDataTankPartialRestore
	// sentinel (not nil, and not errDataTankAllTiersFailed which means nothing
	// moved). Any OTHER error means the case did not reach the partial path.
	if !errors.Is(mErr, errDataTankPartialRestore) {
		t.Fatalf("expected errDataTankPartialRestore (the partial-result signal), got: %v", mErr)
	}
	if len(res.partitionFailures) == 0 {
		t.Fatalf("expected a partial (degraded) result with >=1 failed partition, got %+v", res)
	}

	// Production gate: removeOldDataDirOnMigrationSuccess(res.committed, location).
	// A partial result is NOT a confirmed-100%-complete migration, so committed
	// must be false here for the gate to preserve the old dir.
	if res.committed {
		t.Errorf("partial success (%d partition(s) failed) must NOT report committed=true: "+
			"the deletion gate keys on committed, and removing the old dir would lose the "+
			"failed partition's only recoverable copy", len(res.partitionFailures))
	}
	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)

	// Safety dump retained.
	if err := dtAssertDumpDirRetained(filepath.Join(backupDir, "data-tank")); err != nil {
		t.Errorf("safety dump not retained after partial success: %v", err)
	}
	// The DT-F4 fixture loads 20 partitions x 5 rows = 100 source rows; all of
	// them must still be present in the preserved original.
	assertPreservedOldDirHasRows(t, srcCluster, oldDataDir, dataTankRowCountSQL, 100)
}

// TestDeletionGate_DataTank_Interrupted_ThenReRun: a migration interrupted
// mid-restore (SIGKILL / OOM / pod restart) preserves the old dir, and a re-run
// on the next start does not destroy it. The old data directory survives BOTH the
// interruption and the subsequent re-run.
func TestDeletionGate_DataTank_Interrupted_ThenReRun(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	srcSQL, err := dtReadFixture("DT-F_base.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	_, _, srcCluster, targetCluster, oldDataDir, backupDir := gateClusters(t, srcSQL)

	// First attempt: interrupted mid-restore.
	got, _, mErr := runDataTankMigration(ctx, srcCluster, targetCluster, backupDir, dtCaseSetup{interruptMidRestore: true})
	if mErr != nil {
		t.Fatalf("first attempt returned unexpected error: %v", mErr)
	}
	if got != dtOutcomeDataPreservedOnDisk {
		t.Fatalf("expected DataPreservedOnDisk after interruption, got %s", got)
	}
	removeOldDataDirOnMigrationSuccess(false, oldDataDir)

	// Old dir intact across the interruption (path + PG_VERSION still here).
	if err := dtAssertOldDataDirPreserved(oldDataDir); err != nil {
		t.Fatalf("old dir not intact across the interruption: %v", err)
	}

	// Re-run on next start (no interruption this time). The re-run reads the same
	// still-live source and restores into the same target; it must not destroy the
	// old dir before it confirms success.
	got2, _, mErr2 := runDataTankMigration(ctx, srcCluster, targetCluster, backupDir, dtCaseSetup{})
	if mErr2 != nil {
		t.Fatalf("re-run returned unexpected error: %v", mErr2)
	}
	// The re-run completes (clean tier-1 restore) OR stays preserved; either way
	// the old dir must survive until the gate confirms success. Confirm the gate
	// only fires when the re-run committed.
	committed := isDataTankSuccess(got2)
	removeOldDataDirOnMigrationSuccess(committed, oldDataDir)

	if committed {
		// Gate fired: old dir removed only after the re-run confirmed success.
		assertOldDataDirCleared(t, oldDataDir)
	} else {
		// Re-run did not commit: old dir must still be present + queryable.
		assertPreservedOldDirHasRows(t, srcCluster, oldDataDir, dataTankRowCountSQL, 100)
	}
}

// TestDeletionGate_DataTank_FullSuccess_GateRemovesOnlyAfterCommit: a clean
// migration commits; the gate is the ONLY remover and removes the old dir ONLY
// when committed=true. The two-step assertion proves the gate is a no-op before
// success is confirmed and removes after.
func TestDeletionGate_DataTank_FullSuccess_GateRemovesOnlyAfterCommit(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	srcSQL, err := dtReadFixture("DT-F_base.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	_, _, srcCluster, targetCluster, oldDataDir, backupDir := gateClusters(t, srcSQL)

	res, mErr := migrateDataTank(ctx, dtClusterRef(srcCluster), dtClusterRef(targetCluster),
		backupDir, filepath.Join(backupDir, "data-tank-migration-status.json"), dtParallelism(), migrationFaults{})
	if mErr != nil {
		t.Fatalf("clean migration returned unexpected error: %v", mErr)
	}
	if !res.committed {
		t.Fatalf("expected committed=true for a clean data-tank migration, got %+v", res)
	}

	// Before confirming success the gate must be a no-op: call it with committed
	// hard-false and prove the old dir is untouched.
	removeOldDataDirOnMigrationSuccess(false, oldDataDir)
	if _, statErr := os.Stat(oldDataDir); statErr != nil {
		t.Fatalf("gate removed the old dir while committed=false (it must only remove on success): %v", statErr)
	}

	// Now fire the gate with the real confirmed-success signal.
	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)
	assertOldDataDirCleared(t, oldDataDir)
}

// -----------------------------------------------------------------------------
// Public shape.
// -----------------------------------------------------------------------------

// TestDeletionGate_Public_PreCheckAborted_PreservesEverything: the public shape's
// collation pre-check fires (non-ASCII data under a btree index) and skips the
// restore. Old dir preserved + queryable; the gate does not fire.
func TestDeletionGate_Public_PreCheckAborted_PreservesEverything(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	const publicSQL = `
create table public.things (id int primary key, label text);
insert into public.things values (1, 'Zürich'), (2, 'Köln'), (3, 'Genève');
create index things_label_idx on public.things (label);`

	src, target, srcCluster, targetCluster, oldDataDir, backupDir := gateClusters(t, publicSQL)

	res, mErr := runMigrationEngine(ctx, publicMigrationShape(), src, target,
		backupDir, filepath.Join(backupDir, "public-migration-status.json"), dtParallelism(), migrationFaults{})
	if !errors.Is(mErr, errMigrationPreflightSkipped) {
		t.Fatalf("expected errMigrationPreflightSkipped from the public collation pre-check, got err=%v res=%+v", mErr, res)
	}
	if res.committed {
		t.Fatalf("a pre-flight-skipped public migration must not be committed")
	}
	_ = targetCluster

	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)
	// 3 source rows survive in the preserved original.
	assertPreservedOldDirHasRows(t, srcCluster, oldDataDir, publicRowCountSQL, 3)
}

// TestDeletionGate_Public_RestoreFailed_PreservesEverything: the public shape's
// restore fails for every tier (target unreachable). Old dir preserved +
// queryable; safety dump retained; the gate does not fire.
func TestDeletionGate_Public_RestoreFailed_PreservesEverything(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	const publicSQL = `
create table public.things (id int primary key, name text);
insert into public.things values (1, 'alpha'), (2, 'bravo'), (3, 'charlie'), (4, 'delta');`

	src, target, srcCluster, targetCluster, oldDataDir, backupDir := gateClusters(t, publicSQL)

	// Make the restore fail at every tier: stop the target so pg_restore cannot
	// connect. The dump (read from the live source) still completes, so the safety
	// dump is retained.
	targetCluster.stop()

	res, mErr := runMigrationEngine(ctx, publicMigrationShape(), src, target,
		backupDir, filepath.Join(backupDir, "public-migration-status.json"), dtParallelism(), migrationFaults{})
	if mErr == nil {
		t.Fatalf("expected a restore failure when the target is unreachable, got committed res=%+v", res)
	}
	if res.committed {
		t.Fatalf("a failed public restore must not be committed (res=%+v)", res)
	}

	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)

	// The public dump is a custom-format FILE under backups/public.dump.
	if err := assertDumpRetained(filepath.Join(backupDir, "public.dump")); err != nil {
		t.Errorf("public safety dump not retained after restore failure: %v", err)
	}
	assertPreservedOldDirHasRows(t, srcCluster, oldDataDir, publicRowCountSQL, 4)
}

// -----------------------------------------------------------------------------
// Wired startup path (exec-6c).
// -----------------------------------------------------------------------------

// TestDeletionGate_WiredStartup_FailurePreservesOldDir: drive a real failure
// through the production entry point migrateDataTankSchemasOnStartup (target
// stopped -> every tier fails -> committed=false), feed the actual result into
// the gate exactly as restoreDBBackup's cross-major success branch does, and
// prove the old dir is preserved + queryable. This is the wired-path half of the
// guarantee.
func TestDeletionGate_WiredStartup_FailurePreservesOldDir(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	srcSQL, err := dtReadFixture("DT-F_base.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	old, newRef, srcCluster, targetCluster, oldDataDir, backupDir := gateClusters(t, srcSQL)

	// Stop the target so the wired migration's restore cannot land.
	targetCluster.stop()

	res, mErr := migrateDataTankSchemasOnStartup(ctx, old, newRef, backupDir)
	if res.committed {
		t.Fatalf("expected committed=false when the target is unreachable (err=%v res=%+v)", mErr, res)
	}

	// restoreDBBackup: dtCommitted = res.committed; gate(dtCommitted, location).
	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)

	assertPreservedOldDirHasRows(t, srcCluster, oldDataDir, dataTankRowCountSQL, 100)
}

// TestDeletionGate_WiredStartup_FullSuccessRemovesOldDir: the combined-success
// wired path. A clean data-tank migration through migrateDataTankSchemasOnStartup
// commits, and (public success already assumed) the single gate removes the old
// dir only on the combined committed=true.
func TestDeletionGate_WiredStartup_FullSuccessRemovesOldDir(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	srcSQL, err := dtReadFixture("DT-F_base.sql")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	old, newRef, _, _, oldDataDir, backupDir := gateClusters(t, srcSQL)

	res, mErr := migrateDataTankSchemasOnStartup(ctx, old, newRef, backupDir)
	if mErr != nil {
		t.Fatalf("wired data-tank migration returned error: %v", mErr)
	}
	if !res.committed {
		t.Fatalf("expected committed=true for a clean wired migration, got %+v", res)
	}

	// committed=false would leave the dir; committed=true (combined success)
	// removes it. Prove the gate fires only on the confirmed combined success.
	removeOldDataDirOnMigrationSuccess(res.committed, oldDataDir)
	assertOldDataDirCleared(t, oldDataDir)
}
