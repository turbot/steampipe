package db_local

// Shared cross-major (PG14 -> PG18) copy-and-fallback migration engine.
//
// This is the single engine both data shapes run on (the 2026-06-08 governing
// decision: "public-schema and data migrations converge onto one engine; their
// differences are parameters"). The public-schema migration and the data-tank
// migration differ only in a small parameter set held by migrationShape; every
// other step - the insurance dump, the reserved-word scan, the fallback restore
// ladder (parallel pg_restore -> serial pg_restore -> per-table COPY ->
// per-partition COPY), the post-restore sanity check, and the terminal
// data-preservation rule - is shared, not duplicated.
//
// Terminal rule (the hard invariant): the old data directory is removed ONLY on
// confirmed full success, in exactly one place (removeOldDataDirOnMigrationSuccess).
// Any failure OR partial result leaves the new version running (possibly
// empty/partial) with the original preserved on disk in two independent forms -
// the untouched old data directory plus the retained safety dump. No
// version-revert.

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

// migrationShape holds the per-shape parameters that distinguish the
// public-schema migration from the data-tank migration. Everything the engine
// does beyond reading these fields is identical for both shapes.
type migrationShape struct {
	// name identifies the shape in logs ("public" | "data-tank").
	name string

	// collationPreCheck runs the non-ASCII / collation pre-flight scan
	// (runPreflightCollationScan) against the source before restoring. ON for
	// public (arbitrary user content can carry non-ASCII text under
	// collation-ordered indexes/views); OFF for data tank (its key columns were
	// shown ASCII-only, so the scan would never fire - exec-5-followup Q2).
	collationPreCheck bool

	// rowChecksumValidation runs the post-restore row-count + sample-row
	// checksum + index-validity pass (runValidateRestore) comparing old vs new.
	// ON for public; OFF for data tank (light-migration: a structural sanity
	// check only).
	rowChecksumValidation bool

	// refreshPauseHook gates the orchestrator refresh-pause coordination step.
	// ON for data tank (an upstream refresh must be paused before the dump);
	// OFF for public.
	refreshPauseHook bool

	// copyTierCeiling is the highest restore tier the ladder may escalate to for
	// this shape. The per-table / per-partition COPY tiers reconstruct
	// PARTITION BY LIST topology from the live catalog, which only the data-tank
	// shape has; the public shape carries arbitrary objects (views, functions,
	// extensions) that COPY cannot rebuild, so it stops after the pg_restore
	// tiers.
	//   - data tank: dtRestoreTier4PerPartition (full ladder)
	//   - public:    dtRestoreTier2Serial (pg_restore tiers only)
	copyTierCeiling dataTankRestoreTier

	// dumpSubdir is the dump artefact location under backupDir (a directory for
	// the directory-format data-tank dump; a file for the custom-format public
	// dump).
	dumpSubdir string

	// listSchemas enumerates the schemas this shape migrates from a live source
	// connection (public => ["public"]; data tank => the <handle>/<handle>-parts
	// pairs).
	listSchemas func(ctx context.Context, conn *pgx.Conn) ([]string, error)

	// dumpFn, restoreTier1Fn, restoreTier2Fn supply the shape-specific insurance
	// dump and the two pg_restore tiers of the shared ladder. They differ only in
	// the dump FORMAT and the schema-selection flags: the data-tank shape uses a
	// directory-format dump restored with parallel / serial pg_restore over fresh
	// <handle> schemas; the public shape uses a custom-format --schema=public dump
	// restored with --single-transaction (which tolerates the pre-existing public
	// schema, where a directory restore's CREATE SCHEMA public would abort). The
	// reset-between-tiers, escalation, ceiling and (data-tank) COPY tiers are
	// shared and live in runTieredRestore.
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
		copyTierCeiling:       dtRestoreTier4PerPartition,
		dumpSubdir:            "data-tank",
		listSchemas:           listDataTankSchemas,
		dumpFn:                dumpDataTankSchemas,
		restoreTier1Fn:        restoreTier1ParallelDump,
		restoreTier2Fn:        restoreTier2SerialDump,
	}
}

// publicMigrationShape is the public-schema parameterisation of the shared
// engine. The COPY tiers are unreachable (a public schema carries arbitrary
// objects that COPY cannot rebuild), so the ladder stops after the pg_restore
// tiers; those use a custom-format --schema=public dump restored with
// --single-transaction.
func publicMigrationShape() migrationShape {
	return migrationShape{
		name:                  "public",
		collationPreCheck:     true,
		rowChecksumValidation: true,
		refreshPauseHook:      false,
		copyTierCeiling:       dtRestoreTier2Serial,
		dumpSubdir:            "public.dump",
		listSchemas: func(ctx context.Context, _ *pgx.Conn) ([]string, error) {
			return []string{"public"}, nil
		},
		dumpFn:         dumpPublicSchemaCustom,
		restoreTier1Fn: restorePublicSchemaCustom,
		restoreTier2Fn: func(ctx context.Context, target *pgClusterRef, dumpPath string) error {
			return restorePublicSchemaCustom(ctx, target, dumpPath, 1)
		},
	}
}

// dumpPublicSchemaCustom runs a custom-format pg_dump of the public schema to a
// single file, mirroring production's takeBackup. The shared engine reuses this
// dump for both pg_restore tiers.
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

// restorePublicSchemaCustom restores the custom-format public dump with
// --single-transaction. Unlike the directory-format restore, this tolerates the
// pre-existing public schema on the target. jobs is ignored (a single
// transaction cannot run parallel jobs); both ladder tiers map onto it because
// the public shape has no parallel-vs-serial distinction.
func restorePublicSchemaCustom(ctx context.Context, target *pgClusterRef, dumpPath string, _ int) error {
	args := []string{
		dumpPath,
		"--format=custom",
		"--schema=public",
		"--single-transaction",
		"--no-owner",
		"--dbname=" + target.dbName,
		"--username=" + target.user,
	}
	args = append(args, target.toolConnArgs()...)
	cmd := exec.CommandContext(ctx, target.tool("pg_restore"), args...)
	cmd.Env = target.env
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pg_restore (custom, public) failed: %w\n%s", err, out)
	}
	return nil
}

// runCopyFallbackMigration is the shared engine. Given a shape and the two
// cluster handles it runs the same flow for both data shapes:
//
//  1. enumerate the shape's schemas and (for the COPY tiers) the partitioned
//     parent tables;
//  2. disk pre-flight;
//  3. refresh-pause coordination (shape-gated);
//  4. optional collation pre-check (shape-gated);
//  5. reserved-word scan -> route decision;
//  6. insurance dump (directory format, reused by every restore tier);
//  7. the shared fallback restore ladder (capped at the shape's tier ceiling);
//  8. post-restore sanity check + optional row-checksum validation (shape-gated);
//  9. commit.
//
// On any failure or partial result the original is preserved on disk and the
// failure is surfaced (oldClusterRetained=true + a non-nil error). The deletion
// gate (removeOldDataDirOnMigrationSuccess) is NOT called here - it is the
// caller's single, success-only side effect - so the engine stays unit-testable
// against two real clusters without touching the install-dir filesystem.
func runCopyFallbackMigration(ctx context.Context, shape migrationShape, src, target *pgClusterRef, backupDir, statusPath string, jobs int, faults dataTankMigrationFaults) (dataTankMigrationResult, error) {
	res := dataTankMigrationResult{}

	srcConn, err := src.connect(ctx)
	if err != nil {
		return res, err
	}

	// Step 1: enumerate the schemas this shape migrates and the partitioned
	// parents the COPY tiers iterate.
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

	dumpDir := filepath.Join(backupDir, shape.dumpSubdir)
	res.dumpDir = dumpDir

	// Step 2: disk pre-flight.
	if faults.forceDiskPreflightFail {
		srcConn.Close(ctx)
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, "disk pre-flight: insufficient space for migration window")
		return res, errDataTankDiskPreflight
	}
	if shortfall, derr := dataTankDiskPreflight(ctx, srcConn, schemas, target.dataDir); derr == nil && shortfall > 0 {
		srcConn.Close(ctx)
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, fmt.Sprintf("disk pre-flight: need %d more bytes", shortfall))
		return res, errDataTankDiskPreflight
	}

	// Step 3: refresh-pause coordination (shape-gated). If the orchestrator did
	// not honour the pause before the dump, surface it rather than dumping over a
	// live refresh.
	if shape.refreshPauseHook && faults.refreshPauseNotHonoured {
		srcConn.Close(ctx)
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, "refresh-pause hook not honoured by orchestrator")
		return res, errDataTankRefreshPause
	}

	// Step 4: optional collation pre-check (shape-gated). A flagged risk OR a scan
	// error skips the restore conservatively and preserves the original. This is
	// the public shape's pre-flight gate; the data-tank shape skips it.
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
	if derr := shape.dumpFn(ctx, src, schemas, dumpDir, jobs); derr != nil {
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, derr.Error())
		return res, fmt.Errorf("%w: %v", errDataTankDumpFailed, derr)
	}

	// Step 7: the shared fallback restore ladder. An interrupted migration or an
	// unusable target preserves the original and surfaces the failure.
	if faults.interruptMidRestore || faults.targetUnusable {
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, "restore interrupted / target cluster unusable; original preserved on disk")
		return res, errDataTankAllTiersFailed
	}

	tier, partFailures, restoreErr := runTieredRestore(ctx, src, target, parents, dumpDir, jobs, res.reservedWordRouted, shape.copyTierCeiling, shape.restoreTier1Fn, shape.restoreTier2Fn, faults)
	res.tierReached = tier
	res.partitionFailures = partFailures
	if restoreErr != nil {
		// Every tier (up to this shape's ceiling) failed. Preserve the original;
		// alert the orchestrator.
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, restoreErr.Error())
		return res, errDataTankAllTiersFailed
	}

	// Step 8: post-restore sanity check (every expected parent + partition
	// present), then the optional row-checksum validation pass (shape-gated).
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

	// Step 9: commit - but only if the restore was FULLY complete. A non-empty
	// partitionFailures list means the per-partition COPY tier isolated and skipped
	// at least one partition: most rows landed, but that partition's rows never
	// reached the new cluster. The sanityCheckRestore above only confirms the
	// partition TABLE exists, not that its rows arrived, so it does not catch this.
	// This is a PARTIAL migration, not a success: leave committed=false so the
	// deletion gate preserves the old PG14 data dir alongside the safety dump, and
	// surface a distinct partial sentinel (the governing decision's hard invariant:
	// a partition that did not migrate must survive in the preserved original, not
	// only in the dump).
	if len(res.partitionFailures) > 0 {
		res.oldClusterRetained = true
		writeDataTankStatus(statusPath, res, fmt.Sprintf("%d partition(s) could not be migrated; original preserved on disk", len(res.partitionFailures)))
		return res, errDataTankPartialRestore
	}

	res.committed = true
	writeDataTankStatus(statusPath, res, "")
	return res, nil
}

// removeOldDataDirOnMigrationSuccess is THE single deletion gate. It is the only
// code that removes the old (source) data directory, and it does so only after a
// migration has been confirmed fully committed. Any caller that has not proven
// full success must not call it; on a failed or partial migration the old
// directory is the user's preserved copy of the original data.
//
// committed MUST be the engine's confirmed-success signal
// (dataTankMigrationResult.committed). A best-effort removal failure is logged
// and swallowed: a left-behind old directory is safe (data is preserved), only
// wasteful.
func removeOldDataDirOnMigrationSuccess(committed bool, oldDataDir string) {
	if !committed {
		// Not a confirmed full success - preserve the original. This is the
		// invariant the governing decision protects.
		return
	}
	if oldDataDir == "" {
		return
	}

	// oldDataDir is the old cluster's data directory. In Pipes it is a mounted
	// volume - delete its CONTENTS, not the directory itself (you cannot remove a
	// mount point). Single flow: identical behaviour on a laptop. The old binaries
	// (the sibling postgres/ dir) are left untouched - harmless, and on a baked
	// image they are an ephemeral layer anyway; the dir-selection scan ignores a
	// binaries-only old install because it requires data/PG_VERSION.

	// Commit (point of no return): unlink PG_VERSION FIRST. This single atomic
	// operation makes the directory stop reading as a migratable cluster (the
	// startup trigger keys on PG_VERSION), so a crash during the bulk cleanup
	// below can never leave a corrupt-but-migratable source behind.
	pgVersion := filepath.Join(oldDataDir, "PG_VERSION")
	if err := os.Remove(pgVersion); err != nil && !os.IsNotExist(err) {
		log.Printf("[WARN] migration commit: could not remove %s (%v); old data left in place", pgVersion, err)
		return
	}

	// Cleanup: remove the remaining contents. Crash-tolerant - any leftover files
	// are harmless now that PG_VERSION is gone. The directory itself is preserved.
	if err := putils.RemoveDirectoryContents(oldDataDir); err != nil {
		log.Printf("[WARN] migration commit: could not clear old data dir contents at %s: %v", oldDataDir, err)
	}
}
