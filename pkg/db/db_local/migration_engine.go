package db_local

// Shared cross-major (PG14 -> PG18) migration engine (runMigrationEngine): dump, fallback restore ladder, validation,
// commit.
//
// This is the single engine both data shapes run on (the 2026-06-08 governing decision: "public-schema and data
// migrations converge onto one engine; their differences are parameters"). The public-schema migration and the
// data-tank migration differ only in a small parameter set held by migrationShape; every other step - the insurance
// dump, the reserved-word scan, the fallback restore ladder (parallel pg_restore -> serial pg_restore -> per-table COPY
// -> per-partition COPY), the post-restore sanity check, and the terminal data-preservation rule - is shared, not
// duplicated.
//
// Terminal rule (the hard invariant): the old data directory is removed ONLY on confirmed full success, in exactly one
// place (removeOldDataDirOnMigrationSuccess). Any failure OR partial result preserves the original on disk in the
// untouched old data directory - plus, once the dump step has run, that attempt's retained safety dump as a second
// independent copy (a retry replaces the dump from the still-intact old directory) - and surfaces the failure to the
// caller. Whether the service then runs is the CALLER's decision; the production caller (restoreDBBackup) fail-stops
// startup on any cross-major failure. No version-revert.
//
// THE FULL FLOW (orchestrated by prepareBackup/restoreDBBackup in backup.go; the engine below is steps 4-7):
//
//  1. Detection: startup scans ~/.steampipe/db/ for an old version dir holding BOTH postgres/bin/postgres and
//     data/PG_VERSION. PG_VERSION is postgres's own file (initdb writes it), which is why detection can key on it:
//     it predates this code on every existing install. Same-major bump -> light dump/restore, warn-and-continue.
//     Cross-major -> everything below, and every failure fail-stops startup.
//  2. Start the old server and dump its public schema (the insurance copy). The old server stays up for steps 6-7.
//  3. Write db/migration-incomplete.flag - OUR file, on the new side's tree. The new data dir is a valid, bootable
//     cluster from the moment initdb runs, so a 40% restore is indistinguishable from a finished one by anything
//     postgres writes; the flag is the only witness that a restore started and was never confirmed complete.
//  4. Pre-flight: scan the old data for non-ASCII text under collation-ordered indexes (public shape only).
//  5. Restore into the new version, escalating: parallel pg_restore -> serial -> per-table COPY -> per-partition COPY.
//  6. Validate against the still-live old server: row counts, sample checksums, index validity (public shape only).
//  7. Data-tank schemas (when present) run through this same engine with the pre-flight/validation gates off.
//  8. Status files: the CLI writes ~/.steampipe/db/public-migration-status.json (and
//     data-tank-migration-status.json when tanks exist) on EVERY outcome; "committed" is the verdict field a
//     consuming orchestrator reads to decide whether the old volume can be unmounted/released.
//
// Two endings:
//   - Success: clear the flag -> stop the old server -> retain the dump -> clear the old dir's contents, unlinking
//     PG_VERSION FIRST (the old side atomically stops looking migratable; new side was declared trustworthy first).
//   - Failure: startup fails with retry/opt-out/recovery instructions. Old dir untouched (PG_VERSION still present,
//     so the next start detects and re-runs; a retry always wipes the new side and redoes it from the old data).
//     Deleting the flag is the explicit operator opt-out.
//
// Why BOTH PG_VERSION and the flag: they answer different questions from different disks. Old PG_VERSION = "is there
// old data to migrate from?" - but startup can only ask "do I SEE it now?", and in a hosted deployment the old data
// can live on a separate volume that is simply not mounted (in which case absence proves nothing). The flag lives on
// the NEW side's tree, so it travels with the database it vouches for: present + no pending migration detected ->
// REFUSE to start rather than boot a half-restored database that looks healthy. State table:
//
//   | old PG_VERSION              | flag    | meaning                                  | startup behaviour            |
//   |-----------------------------|---------|------------------------------------------|------------------------------|
//   | absent (no old install)     | absent  | fresh machine / fresh pod                | fresh initdb                 |
//   | present                     | absent  | migration pending, dump never completed  | run the migration            |
//   | present                     | present | crash mid-restore / post-dump failure    | re-run (wipe new side first) |
//   | absent (unlinked at commit) | absent  | post-migration steady state              | normal start                 |
//   | leftovers, no PG_VERSION    | absent  | died during post-commit bulk cleanup     | normal start (leftovers inert)|
//   | NOT VISIBLE (vol unmounted) | present | half-restored new db, detection blind    | REFUSE to start              |
//   | not visible                 | absent  | complete db on its own volume            | normal start                 |
//   | present                     | deleted by operator | explicit opt-out             | start without migrating      |

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	putils "github.com/turbot/pipe-fittings/v2/utils"
)

// migrationShape holds the per-shape parameters that distinguish the public-schema migration from the data-tank
// migration. Everything the engine does beyond reading these fields is identical for both shapes.
type migrationShape struct {
	// name identifies the shape in logs ("public" | "data-tank").
	name string

	// collationPreCheck runs the non-ASCII / collation pre-flight scan (runPreflightCollationScan) against the source
	// before restoring. ON for public (arbitrary user content can carry non-ASCII text under collation-ordered
	// indexes/views); OFF for data tank (its key columns were shown ASCII-only, so the scan would never fire).
	collationPreCheck bool

	// rowChecksumValidation runs the post-restore row-count + sample-row checksum + index-validity pass
	// (runValidateRestore) comparing old vs new. ON for public; OFF for data tank (light-migration: a structural sanity
	// check only).
	rowChecksumValidation bool

	// refreshPauseHook gates the orchestrator refresh-pause coordination step. ON for data tank (an upstream refresh
	// must be paused before the dump); OFF for public.
	refreshPauseHook bool

	// copyTierCeiling is the highest restore tier the ladder may escalate to for this shape. The per-table /
	// per-partition COPY tiers reconstruct PARTITION BY LIST topology from the live catalog, which only the data-tank
	// shape has; the public shape carries arbitrary objects (views, functions, extensions) that COPY cannot rebuild, so
	// it stops after the pg_restore tiers.
	//   - data tank: dtRestoreTier4PerPartition (full ladder)
	//   - public:    dtRestoreTier2Serial (pg_restore tiers only)
	copyTierCeiling dataTankRestoreTier

	// dumpSubdir is the dump artefact location under backupDir (a directory for the directory-format data-tank dump; a
	// file for the custom-format public dump).
	dumpSubdir string

	// listSchemas enumerates the schemas this shape migrates from a live source connection (public => ["public"]; data
	// tank => the <handle>/<handle>-parts pairs).
	listSchemas func(ctx context.Context, conn *pgx.Conn) ([]string, error)

	// matviewRefresh splits the restore into two separate pg_restore invocations - the objects+static-data list first,
	// then the REFRESH MATERIALIZED VIEW list - each in its own --single-transaction. The first commits the data; a
	// failure of the second is WARNED and recorded (matviewRefreshFailed), never rolled back - a failed matview refresh
	// is a warning, not a migration failure. ON for public (matviews with transitive unqualified table references fail
	// to refresh under pg_dump's blank search_path); OFF for data tank (no matviews).
	matviewRefresh bool

	// dumpFn, restoreTier1Fn, restoreTier2Fn supply the shape-specific insurance dump and the two pg_restore tiers of
	// the shared ladder. They differ only in the dump FORMAT and the schema-selection flags: the data-tank shape uses a
	// directory-format dump restored with parallel / serial pg_restore over fresh <handle> schemas; the public shape
	// uses a custom-format --schema=public dump restored with --single-transaction (which tolerates the pre-existing
	// public schema, where a directory restore's CREATE SCHEMA public would abort). The reset-between-tiers,
	// escalation, ceiling and (data-tank) COPY tiers are shared and live in runTieredRestore.
	dumpFn         func(ctx context.Context, src *pgClusterRef, schemas []string, dumpPath string, jobs int) error
	restoreTier1Fn func(ctx context.Context, target *pgClusterRef, dumpPath string, jobs int) error
	restoreTier2Fn func(ctx context.Context, target *pgClusterRef, dumpPath string) error
}

// dataTankMigrationShape is the data-tank parameterisation of the shared engine.
func dataTankMigrationShape() migrationShape {
	return migrationShape{
		name:                  "data-tank",
		collationPreCheck:     false,
		rowChecksumValidation: false,
		refreshPauseHook:      true,
		matviewRefresh:        false,
		copyTierCeiling:       dtRestoreTier4PerPartition,
		dumpSubdir:            "data-tank",
		listSchemas:           listDataTankSchemas,
		dumpFn:                dumpDataTankSchemas,
		restoreTier1Fn:        restoreTier1ParallelDump,
		restoreTier2Fn:        restoreTier2SerialDump,
	}
}

// publicMigrationShape is the public-schema parameterisation of the shared engine. The COPY tiers are unreachable (a
// public schema carries arbitrary objects that COPY cannot rebuild), so the ladder stops after the pg_restore tiers;
// those restore the objects+static-data list of a custom-format --schema=public dump with --single-transaction, leaving
// the matview-refresh list to the engine's separate, tolerated refresh step (matviewRefresh).
func publicMigrationShape() migrationShape {
	return migrationShape{
		name:                  "public",
		collationPreCheck:     true,
		rowChecksumValidation: true,
		refreshPauseHook:      false,
		matviewRefresh:        true,
		copyTierCeiling:       dtRestoreTier2Serial,
		dumpSubdir:            "public.dump",
		listSchemas: func(ctx context.Context, _ *pgx.Conn) ([]string, error) {
			return []string{"public"}, nil
		},
		dumpFn:         dumpPublicSchemaCustom,
		restoreTier1Fn: restorePublicObjectsOnly,
		restoreTier2Fn: func(ctx context.Context, target *pgClusterRef, dumpPath string) error {
			return restorePublicObjectsOnly(ctx, target, dumpPath, 1)
		},
	}
}

// dumpPublicSchemaCustom runs a custom-format pg_dump of the public schema to a single file, mirroring production's
// takeBackup. The shared engine reuses this dump for both pg_restore tiers.
func dumpPublicSchemaCustom(ctx context.Context, src *pgClusterRef, _ []string, dumpPath string, _ int) error {
	if err := os.MkdirAll(filepath.Dir(dumpPath), 0755); err != nil {
		return err
	}
	args := []string{
		"--file=" + dumpPath,
		"--format=custom",
		"--schema=public",
		"--dbname=" + src.dbName,
		"--username=" + src.user,
	}
	args = append(args, src.toolConnArgs()...)
	cmd := exec.CommandContext(ctx, src.tool("pg_dump"), args...)
	cmd.Env = src.env
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pg_dump (custom, public) failed: %w\n%s", err, out)
	}
	return nil
}

// publicDumpListFiles extracts the dump's table of contents and partitions it into the objects+static-data list and the
// matview-refresh list, written alongside the dump. The caller removes both files.
func publicDumpListFiles(ctx context.Context, target *pgClusterRef, dumpPath string) (string, string, error) {
	toc, err := getTableOfContentsFromBackup(ctx, target, dumpPath)
	if err != nil {
		return "", "", err
	}
	return partitionTableOfContents(toc, filepath.Dir(dumpPath))
}

// restorePublicObjectsOnly restores the custom-format public dump WITHOUT its REFRESH MATERIALIZED VIEW entries, in one
// pg_restore --single-transaction invocation (which tolerates the pre-existing public schema, unlike a directory-format
// restore). The refresh entries run later as their own transaction (refreshPublicMatviews) so a failed refresh cannot
// roll back the committed objects+data. jobs is ignored (a single transaction cannot run parallel jobs); both ladder
// tiers map onto it because the public shape has no parallel-vs-serial distinction.
func restorePublicObjectsOnly(ctx context.Context, target *pgClusterRef, dumpPath string, _ int) error {
	objectListFile, matviewListFile, err := publicDumpListFiles(ctx, target, dumpPath)
	if err != nil {
		return err
	}
	defer os.Remove(objectListFile)
	defer os.Remove(matviewListFile)
	return runRestoreUsingList(ctx, target, dumpPath, objectListFile)
}

// refreshPublicMatviews runs ONLY the REFRESH MATERIALIZED VIEW entries of the public dump, as a second, separate
// pg_restore --single-transaction invocation - the tolerated half of the matview split. The objects+data restore has
// already committed; the engine treats a failure here as a warning, never a rollback.
func refreshPublicMatviews(ctx context.Context, target *pgClusterRef, dumpPath string) error {
	objectListFile, matviewListFile, err := publicDumpListFiles(ctx, target, dumpPath)
	if err != nil {
		return err
	}
	defer os.Remove(objectListFile)
	defer os.Remove(matviewListFile)
	return runRestoreUsingList(ctx, target, dumpPath, matviewListFile)
}

// runMigrationEngine is the shared engine. Given a shape and the two cluster handles it runs the same flow for both
// data shapes:
//
//  1. enumerate the shape's schemas and (for the COPY tiers) the partitioned parent tables;
//  2. disk pre-flight;
//  3. refresh-pause coordination (shape-gated);
//  4. optional collation pre-check (shape-gated);
//  5. reserved-word scan -> route decision;
//  6. insurance dump (format per shape: directory for data-tank, custom single-file for public; reused by every restore
//     tier);
//  7. the shared fallback restore ladder (capped at the shape's tier ceiling);
//  8. post-restore sanity check + optional row-checksum validation (shape-gated);
//  9. tolerant matview refresh (shape-gated): a second, separate pg_restore transaction over the REFRESH MATERIALIZED
//     VIEW list; a failure here is warned and recorded, never rolled back;
//  10. commit.
//
// On any failure or partial result the original is preserved on disk and the failure is surfaced
// (oldClusterRetained=true + a non-nil error). The deletion gate (removeOldDataDirOnMigrationSuccess) is NOT called
// here - it is the caller's single, success-only side effect - so the engine stays unit-testable against two real
// clusters without touching the install-dir filesystem.
func runMigrationEngine(ctx context.Context, shape migrationShape, src, target *pgClusterRef, backupDir, statusPath string, jobs int, faults migrationFaults) (migrationResult, error) {
	res := migrationResult{}

	srcConn, err := src.connect(ctx)
	if err != nil {
		return res, err
	}

	// Step 1: enumerate the schemas this shape migrates and the partitioned parents the COPY tiers iterate.
	schemas, err := shape.listSchemas(ctx, srcConn)
	if err != nil {
		srcConn.Close(ctx)
		return res, err
	}
	parents, mgrTables, err := listDataTankParentTables(ctx, srcConn, schemas)
	if err != nil {
		srcConn.Close(ctx)
		return res, err
	}
	// Record the _mgr_ intermediates left to the user-driven cleanup workflow.
	for _, m := range mgrTables {
		res.skippedMgrTables = append(res.skippedMgrTables, qualName(m.schema, m.table))
	}

	dumpPath := filepath.Join(backupDir, shape.dumpSubdir)
	res.dumpPath = dumpPath

	// Step 2: disk pre-flight.
	if faults.forceDiskPreflightFail {
		srcConn.Close(ctx)
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, "disk pre-flight: insufficient space for migration window")
		return res, errDataTankDiskPreflight
	}
	if shortfall, derr := dataTankDiskPreflight(ctx, srcConn, schemas, target.dataDir); derr != nil {
		// An error in the scan itself is a pre-flight failure, not a license to proceed: silently skipping the gate
		// would convert the exact disk-full scenario it guards into a mid-dump failure on a possibly shared volume.
		srcConn.Close(ctx)
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, "disk pre-flight could not run: "+derr.Error())
		return res, fmt.Errorf("%w: scan failed: %v", errDataTankDiskPreflight, derr)
	} else if shortfall > 0 {
		srcConn.Close(ctx)
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, fmt.Sprintf("disk pre-flight: need %d more bytes", shortfall))
		return res, errDataTankDiskPreflight
	}

	// Step 3: refresh-pause coordination (shape-gated). If the orchestrator did not honour the pause before the dump,
	// surface it rather than dumping over a live refresh.
	if shape.refreshPauseHook && faults.refreshPauseNotHonoured {
		srcConn.Close(ctx)
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, "refresh-pause hook not honoured by orchestrator")
		return res, errDataTankRefreshPause
	}

	// Step 4: optional collation pre-check (shape-gated). A flagged risk OR a scan error skips the restore
	// conservatively and preserves the original. This is the public shape's pre-flight gate; the data-tank shape skips
	// it.
	if shape.collationPreCheck {
		risks, perr := runPreflightCollationScan(ctx, srcConn)
		if perr != nil {
			log.Printf("[WARN] %s migration: pre-flight collation scan failed: %v", shape.name, perr)
			srcConn.Close(ctx)
			res.oldClusterRetained = true
			res.preflightSkipped = true
			writeDataTankStatus(statusPath, res, "pre-flight collation scan failed: "+perr.Error())
			return res, errMigrationPreflightSkipped
		}
		if len(risks) > 0 {
			log.Printf("[TRACE] %s migration: pre-flight flagged %d collation risk(s); skipping restore", shape.name, len(risks))
			srcConn.Close(ctx)
			res.oldClusterRetained = true
			res.preflightSkipped = true
			writeDataTankStatus(statusPath, res, fmt.Sprintf("pre-flight flagged %d collation risk(s); restore skipped", len(risks)))
			return res, errMigrationPreflightSkipped
		}
	}

	// Step 5: reserved-word scan -> route decision (not a stop signal).
	hits, err := reservedWordColumnScan(ctx, srcConn, schemas)
	if err != nil {
		srcConn.Close(ctx)
		return res, err
	}
	res.reservedWordRouted = len(hits) > 0
	srcConn.Close(ctx)

	// Step 6: insurance dump (always), unless the harness forces a dump failure.
	if faults.forceDumpFailure {
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, "pg_dump failed (disk full during dump)")
		return res, errDataTankDumpFailed
	}
	if derr := shape.dumpFn(ctx, src, schemas, dumpPath, jobs); derr != nil {
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, derr.Error())
		return res, fmt.Errorf("%w: %v", errDataTankDumpFailed, derr)
	}

	// Step 7: the shared fallback restore ladder. An interrupted migration or an unusable target preserves the original
	// and surfaces the failure.
	if faults.interruptMidRestore || faults.targetUnusable {
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, "restore interrupted / target cluster unusable; original preserved on disk")
		return res, errDataTankAllTiersFailed
	}

	tier, partFailures, restoreErr := runTieredRestore(ctx, src, target, parents, dumpPath, jobs, res.reservedWordRouted, shape.copyTierCeiling, shape.restoreTier1Fn, shape.restoreTier2Fn, faults)
	res.tierReached = tier
	res.partitionFailures = partFailures
	if restoreErr != nil {
		// Every tier (up to this shape's ceiling) failed. Preserve the original; alert the orchestrator.
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, restoreErr.Error())
		return res, errDataTankAllTiersFailed
	}

	// Step 8: post-restore sanity check (every expected parent + partition present), then the optional row-checksum
	// validation pass (shape-gated).
	if serr := sanityCheckRestore(ctx, src, target, parents); serr != nil {
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, "post-restore sanity check failed: "+serr.Error())
		return res, errDataTankAllTiersFailed
	}
	if shape.rowChecksumValidation {
		newConn, nerr := target.connect(ctx)
		if nerr != nil {
			res.oldClusterRetained = true
			res.validationDiverged = true
			writeDataTankStatus(statusPath, res, "post-restore validation: could not connect to new cluster: "+nerr.Error())
			return res, errMigrationValidationDiverged
		}
		oldConn, oerr := src.connect(ctx)
		if oerr != nil {
			newConn.Close(ctx)
			return res, oerr
		}
		divergences, verr := runValidateRestore(ctx, oldConn, newConn)
		oldConn.Close(ctx)
		newConn.Close(ctx)
		if verr != nil {
			res.oldClusterRetained = true
			res.validationDiverged = true
			writeDataTankStatus(statusPath, res, "post-restore validation query failed: "+verr.Error())
			return res, errMigrationValidationDiverged
		}
		if len(divergences) > 0 {
			for _, d := range divergences {
				log.Printf("[WARN] post-restore validation divergence: kind=%s target=%s detail=%s", d.kind, d.target, d.detail)
			}
			res.oldClusterRetained = true
			res.validationDiverged = true
			writeDataTankStatus(statusPath, res, fmt.Sprintf("post-restore validation found %d divergence(s)", len(divergences)))
			return res, errMigrationValidationDiverged
		}
	}

	// Step 9: tolerant matview refresh (shape-gated). The objects+static-data restore above already committed its own
	// transaction; the REFRESH MATERIALIZED VIEW entries now run as a SECOND, separate pg_restore --single-transaction
	// invocation. A failure is warned and recorded - never a rollback - because matviews with transitive unqualified
	// table references cannot refresh under pg_dump's blank search_path and the data is already safely restored (a
	// failed matview refresh is a warning, not a migration failure).
	if shape.matviewRefresh {
		if rerr := refreshPublicMatviews(ctx, target, dumpPath); rerr != nil {
			log.Printf("[WARN] %s migration: could not refresh materialized views (refresh manually): %v", shape.name, rerr)
			res.matviewRefreshFailed = true
		}
	}

	// Step 10: commit - but only if the restore was FULLY complete. A non-empty partitionFailures list means the
	// per-partition COPY tier isolated and skipped at least one partition: most rows landed, but that partition's rows
	// never reached the new cluster. The sanityCheckRestore above only confirms the partition TABLE exists, not that
	// its rows arrived, so it does not catch this. This is a PARTIAL migration, not a success: leave committed=false so
	// the deletion gate preserves the old PG14 data dir alongside the safety dump, and surface a distinct partial
	// sentinel (the governing decision's hard invariant: a partition that did not migrate must survive in the preserved
	// original, not only in the dump).
	if len(res.partitionFailures) > 0 {
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, fmt.Sprintf("%d partition(s) could not be migrated; original preserved on disk", len(res.partitionFailures)))
		return res, errDataTankPartialRestore
	}

	res.committed = true
	writeDataTankStatus(statusPath, res, "")
	return res, nil
}

// removeOldDataDirOnMigrationSuccess is THE single deletion gate. It is the only code that removes the old (source)
// data directory, and it does so only after a migration has been confirmed fully committed. Any caller that has not
// proven full success must not call it; on a failed or partial migration the old directory is the user's preserved copy
// of the original data.
//
// committed MUST be the engine's confirmed-success signal (migrationResult.committed). A best-effort removal failure is
// logged and swallowed: a left-behind old directory is safe (data is preserved), only wasteful.
func removeOldDataDirOnMigrationSuccess(committed bool, oldDataDir string) {
	if !committed {
		// Not a confirmed full success - preserve the original. This is the invariant the governing decision protects.
		return
	}
	if oldDataDir == "" {
		return
	}

	// oldDataDir is the old cluster's data directory. In Pipes it is a mounted volume - delete its CONTENTS, not the
	// directory itself (you cannot remove a mount point). Single flow: identical behaviour on a laptop. The old
	// binaries (the sibling postgres/ dir) are left untouched - harmless, and on a baked image they are an ephemeral
	// layer anyway; the dir-selection scan ignores a binaries-only old install because it requires data/PG_VERSION.

	// Commit (point of no return): unlink PG_VERSION FIRST. This single atomic operation makes the directory stop
	// reading as a migratable cluster (the startup trigger keys on PG_VERSION), so a crash during the bulk cleanup
	// below can never leave a corrupt-but-migratable source behind.
	pgVersion := filepath.Join(oldDataDir, "PG_VERSION")
	if err := os.Remove(pgVersion); err != nil && !os.IsNotExist(err) {
		log.Printf("[WARN] migration commit: could not remove %s (%v); old data left in place", pgVersion, err)
		return
	}

	// Cleanup: remove the remaining contents. Crash-tolerant - any leftover files are harmless now that PG_VERSION is
	// gone. The directory itself is preserved.
	if err := putils.RemoveDirectoryContents(oldDataDir); err != nil {
		log.Printf("[WARN] migration commit: could not clear old data dir contents at %s: %v", oldDataDir, err)
	}
}
