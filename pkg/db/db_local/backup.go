package db_local

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/jackc/pgx/v5"
	"github.com/shirou/gopsutil/process"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	putils "github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

var (
	errDbInstanceRunning = fmt.Errorf("cannot start DB backup - a postgres instance is still running and Steampipe could not kill it. Please kill this manually and restart Steampipe")
)

// targetDatabaseVersion is the version this build targets as the embedded
// PostgreSQL. In production it is constants.DatabaseVersion. It exists as a
// package-level variable purely so the in-package cross-major migration test
// matrix (migration_xmajor_test.go) can drive the cross-major branch in
// restoreDBBackup at run time - the compile-time constant would otherwise pin
// the test to whatever major matches the current shipped DB and the
// classifyPgMigration branch under test would never be reached. NEVER override
// outside tests in the same package.
var targetDatabaseVersion = constants.DatabaseVersion

const (
	backupFormat            = "custom"
	backupDumpFileExtension = "dump"
	backupTextFileExtension = "sql"
)

// pgRunningInfo represents a running pg instance that we need to startup to create the
// backup archive and the name of the installed database
type pgRunningInfo struct {
	cmd    *exec.Cmd
	port   int
	dbName string
}

// stop is used for shutting down postgres instance spun up for extracting dump
// it uses signals as suggested by https://www.postgresql.org/docs/12/server-shutdown.html
// to try to shutdown the db process process.
// It is not expected that any client is connected to the instance when 'stop' is called.
// Connected clients will be forcefully disconnected
func (r *pgRunningInfo) stop(ctx context.Context) error {
	p, err := process.NewProcess(int32(r.cmd.Process.Pid))
	if err != nil {
		return err
	}
	return doThreeStepPostgresExit(ctx, p)
}

const (
	noMatViewRefreshListFileName   = "without_refresh.lst"
	onlyMatViewRefreshListFileName = "only_refresh.lst"
)

// pgMigrationKind classifies an old on-disk PostgreSQL install relative to the
// target (constants.DatabaseVersion).
type pgMigrationKind int

const (
	// pgMigrationMinor - same major, different (older) minor, e.g. 14.17 -> 14.19.
	// The automatic pg_dump+pg_restore path is safe and is retained; only a
	// restore *failure* is made non-fatal. This is the zero value: when in
	// doubt (e.g. an unparseable version) we preserve the historical
	// automatic behaviour rather than strand data.
	pgMigrationMinor pgMigrationKind = iota
	// pgMigrationMajor - different (older) major, e.g. 14 -> 18. pg_restore
	// cannot load a lower-major dump into a higher-major server, so the
	// automatic restore is NOT attempted: an insurance dump is retained,
	// the old data directory is kept, the service starts fresh, and the
	// user is told how to restore manually. (Blocking startup until the
	// user acknowledges is a possible alternative policy, retained as a
	// documented fallback rather than implemented here.)
	pgMigrationMajor
)

// classifyPgMigration decides how an old install (oldVersion, e.g. "14.17.0")
// relates to the target (targetVersion, e.g. "14.19.0").
//
// On a parse failure it returns pgMigrationMinor: that preserves the historical
// automatic dump+restore behaviour (and keeps the green same-major migration
// tests green) rather than stranding data on a routine bump. A cross-major
// jump is only ever concluded from two successfully-parsed versions with
// differing majors.
func classifyPgMigration(oldVersion, targetVersion string) pgMigrationKind {
	ov, errOld := semver.NewVersion(oldVersion)
	tv, errTarget := semver.NewVersion(targetVersion)
	if errOld != nil || errTarget != nil {
		return pgMigrationMinor
	}
	if ov.Major() != tv.Major() {
		return pgMigrationMajor
	}
	return pgMigrationMinor
}

// collationRisk describes a single collation-dependent object found by the
// pre-flight scan. A cross-major restore changes the default collation
// provider, so these objects can have their index ordering / uniqueness or
// view ordering silently change after a restore.
type collationRisk struct {
	kind      string // "text_btree_index" | "text_unique_constraint" | "ordered_view_text"
	schemaObj string // e.g. "public.my_table.my_idx"
	sample    string // sample value / reason showing why flagged
}

// validationDivergence describes a single post-restore mismatch between the old
// PG14 cluster and the new PG18 cluster.
type validationDivergence struct {
	kind   string // "row_count" | "checksum" | "index_invalid"
	target string // e.g. "public.my_table" or "public.my_idx"
	detail string // e.g. "old=1234 new=1230" or "md5 mismatch"
}

// nonAsciiTextDetector is the cheap multi-byte detector used by the pre-flight
// scan. With the embedded DB initdb'd LC_ALL=C --encoding=UTF-8, a value whose
// octet_length differs from its character length necessarily contains a
// multi-byte UTF-8 sequence and therefore non-ASCII content. See the encoding
// guard in runPreflightCollationScan.
const nonAsciiTextDetector = `octet_length(%[1]s) <> length(%[1]s)`

// runPreflightCollationScan inspects the old (source) cluster for
// collation-dependent objects whose underlying text data actually contains
// non-ASCII bytes. It is data-aware (per the locked 2026-06-05 B02/B04 policy):
// a text index/constraint/view over purely ASCII data is NOT flagged, because
// ASCII sorts identically under every collation provider. Returns an empty
// slice when the schema is clean.
func runPreflightCollationScan(ctx context.Context, conn *pgx.Conn) ([]collationRisk, error) {
	// Encoding guard: the non-ASCII detector relies on a multi-byte UTF-8
	// encoding. Under a single-byte encoding (SQL_ASCII / LATIN1) the detector
	// silently returns false negatives, so we conservatively flag-all rather
	// than under-report.
	var encoding string
	if err := conn.QueryRow(ctx, "SELECT pg_encoding_to_char(encoding) FROM pg_database WHERE datname = current_database()").Scan(&encoding); err != nil {
		return nil, err
	}
	multiByteEncoding := encoding == "UTF8" || encoding == "UTF-8"

	var risks []collationRisk

	// 1. B-tree indexes (including unique and expression indexes) on public
	//    tables that touch a text/varchar column. GIN/GiST and other non-btree
	//    access methods are not collation-ordered, so they are excluded.
	const idxQuery = `
SELECT c.relname AS idxname, t.relname AS tabname, i.indisunique
FROM pg_index i
JOIN pg_class c ON c.oid = i.indexrelid
JOIN pg_class t ON t.oid = i.indrelid
JOIN pg_namespace n ON n.oid = t.relnamespace
JOIN pg_am am ON am.oid = c.relam
WHERE n.nspname = 'public'
  AND am.amname = 'btree'
  AND (
    EXISTS (
      SELECT 1 FROM unnest(i.indkey) AS k(attnum)
      JOIN pg_attribute pa ON pa.attrelid = t.oid AND pa.attnum = k.attnum
      WHERE format_type(pa.atttypid, pa.atttypmod) IN ('text','character varying')
         OR format_type(pa.atttypid, pa.atttypmod) LIKE 'character varying%'
    )
    OR pg_get_indexdef(i.indexrelid) ~* '(text|varchar|lower\(|upper\()'
  )
GROUP BY c.relname, t.relname, i.indisunique`

	idxRows, err := conn.Query(ctx, idxQuery)
	if err != nil {
		return nil, err
	}
	type idxHit struct {
		idx, tab string
		unique   bool
	}
	var idxHits []idxHit
	for idxRows.Next() {
		var h idxHit
		if err := idxRows.Scan(&h.idx, &h.tab, &h.unique); err != nil {
			idxRows.Close()
			return nil, err
		}
		idxHits = append(idxHits, h)
	}
	idxRows.Close()
	if err := idxRows.Err(); err != nil {
		return nil, err
	}

	for _, h := range idxHits {
		flag := !multiByteEncoding
		if multiByteEncoding {
			nonAscii, derr := pfTableHasNonASCIIText(ctx, conn, h.tab)
			if derr != nil {
				return nil, derr
			}
			flag = nonAscii
		}
		if flag {
			kind := "text_btree_index"
			if h.unique {
				kind = "text_unique_constraint"
			}
			risks = append(risks, collationRisk{
				kind:      kind,
				schemaObj: fmt.Sprintf("public.%s.%s", h.tab, h.idx),
				sample:    "non-ASCII text data under a collation-ordered index",
			})
		}
	}

	// 2. Views / materialized views whose definition orders by a text column
	//    over non-ASCII data.
	const viewQuery = `
SELECT c.relname, pg_get_viewdef(c.oid) AS def
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public' AND c.relkind IN ('v','m')`
	viewRows, err := conn.Query(ctx, viewQuery)
	if err != nil {
		return nil, err
	}
	type viewHit struct{ name, def string }
	var views []viewHit
	for viewRows.Next() {
		var v viewHit
		if err := viewRows.Scan(&v.name, &v.def); err != nil {
			viewRows.Close()
			return nil, err
		}
		views = append(views, v)
	}
	viewRows.Close()
	if err := viewRows.Err(); err != nil {
		return nil, err
	}
	for _, v := range views {
		if !strings.Contains(strings.ToUpper(v.def), "ORDER BY") {
			continue
		}
		flag := !multiByteEncoding
		if multiByteEncoding {
			anyNonAscii, derr := pfSchemaHasNonASCIIText(ctx, conn)
			if derr != nil {
				return nil, derr
			}
			flag = anyNonAscii
		}
		if flag {
			risks = append(risks, collationRisk{
				kind:      "ordered_view_text",
				schemaObj: fmt.Sprintf("public.%s", v.name),
				sample:    "view orders by text over non-ASCII data",
			})
		}
	}

	return risks, nil
}

// pfTableHasNonASCIIText reports whether any text/varchar column of the named
// public table holds a value with a byte > 0x7F.
func pfTableHasNonASCIIText(ctx context.Context, conn *pgx.Conn, table string) (bool, error) {
	cols, err := pfTextColumns(ctx, conn, table)
	if err != nil {
		return false, err
	}
	for _, col := range cols {
		detector := fmt.Sprintf(nonAsciiTextDetector, pfQuoteIdent(col))
		q := fmt.Sprintf(
			`SELECT EXISTS(SELECT 1 FROM public.%s WHERE %s IS NOT NULL AND %s)`,
			pfQuoteIdent(table), pfQuoteIdent(col), detector)
		var hit bool
		if err := conn.QueryRow(ctx, q).Scan(&hit); err != nil {
			return false, err
		}
		if hit {
			return true, nil
		}
	}
	return false, nil
}

func pfSchemaHasNonASCIIText(ctx context.Context, conn *pgx.Conn) (bool, error) {
	tables, err := pfPublicBaseTables(ctx, conn)
	if err != nil {
		return false, err
	}
	for _, t := range tables {
		hit, err := pfTableHasNonASCIIText(ctx, conn, t)
		if err != nil {
			return false, err
		}
		if hit {
			return true, nil
		}
	}
	return false, nil
}

func pfTextColumns(ctx context.Context, conn *pgx.Conn, table string) ([]string, error) {
	const q = `
SELECT a.attname
FROM pg_attribute a
JOIN pg_class c ON c.oid = a.attrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname='public' AND c.relname=$1 AND a.attnum > 0 AND NOT a.attisdropped
  AND format_type(a.atttypid, a.atttypmod) IN ('text','character varying')`
	rows, err := conn.Query(ctx, q, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func pfPublicBaseTables(ctx context.Context, conn *pgx.Conn) ([]string, error) {
	const q = `
SELECT c.relname
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname='public' AND c.relkind IN ('r','p')
ORDER BY c.relname`
	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func pfQuoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// runValidateRestore runs while the old (source) server is still live. For every
// public base table it compares row count and an order-stable sample-row
// checksum between the old and new clusters, and verifies every public index
// reports indisvalid=true on the new cluster. Any mismatch is a divergence
// regardless of pg_restore's exit code. Returns an empty slice on a full
// match.
func runValidateRestore(ctx context.Context, oldConn, newConn *pgx.Conn) ([]validationDivergence, error) {
	oldTables, err := pfPublicBaseTables(ctx, oldConn)
	if err != nil {
		return nil, err
	}

	var divergences []validationDivergence
	for _, tab := range oldTables {
		oldCount, err := vrTableRowCount(ctx, oldConn, tab)
		if err != nil {
			return nil, err
		}
		newCount, err := vrTableRowCount(ctx, newConn, tab)
		if err != nil {
			divergences = append(divergences, validationDivergence{
				kind:   "row_count",
				target: fmt.Sprintf("public.%s", tab),
				detail: fmt.Sprintf("table missing on new cluster: %v", err),
			})
			continue
		}
		if oldCount != newCount {
			divergences = append(divergences, validationDivergence{
				kind:   "row_count",
				target: fmt.Sprintf("public.%s", tab),
				detail: fmt.Sprintf("old=%d new=%d", oldCount, newCount),
			})
			continue
		}
		oldDigest, err := vrTableSampleChecksum(ctx, oldConn, tab)
		if err != nil {
			return nil, err
		}
		newDigest, err := vrTableSampleChecksum(ctx, newConn, tab)
		if err != nil {
			return nil, err
		}
		if oldDigest != newDigest {
			divergences = append(divergences, validationDivergence{
				kind:   "checksum",
				target: fmt.Sprintf("public.%s", tab),
				detail: "md5 mismatch",
			})
		}
	}

	// every public index on the new cluster must be valid
	invalid, err := vrInvalidIndexes(ctx, newConn)
	if err != nil {
		return nil, err
	}
	for _, idx := range invalid {
		divergences = append(divergences, validationDivergence{
			kind:   "index_invalid",
			target: fmt.Sprintf("public.%s", idx),
			detail: "pg_index.indisvalid = false",
		})
	}

	return divergences, nil
}

func vrTableRowCount(ctx context.Context, conn *pgx.Conn, table string) (int64, error) {
	var n int64
	q := fmt.Sprintf("SELECT count(*) FROM public.%s", pfQuoteIdent(table))
	err := conn.QueryRow(ctx, q).Scan(&n)
	return n, err
}

// vrTableSampleChecksum computes an order-stable md5 over the table's rows. It
// casts the whole row to text and orders by that text so the comparison is
// independent of physical (ctid) ordering, which differs after a dump/restore.
func vrTableSampleChecksum(ctx context.Context, conn *pgx.Conn, table string) (string, error) {
	q := fmt.Sprintf(
		`SELECT coalesce(md5(string_agg(s, E'\n' ORDER BY s)), '') FROM (SELECT r.*::text AS s FROM public.%s r) x`,
		pfQuoteIdent(table))
	var digest string
	if err := conn.QueryRow(ctx, q).Scan(&digest); err != nil {
		return "", err
	}
	return digest, nil
}

func vrInvalidIndexes(ctx context.Context, conn *pgx.Conn) ([]string, error) {
	const q = `
SELECT c.relname
FROM pg_index i
JOIN pg_class c ON c.oid = i.indexrelid
JOIN pg_class t ON t.oid = i.indrelid
JOIN pg_namespace n ON n.oid = t.relnamespace
WHERE n.nspname='public' AND NOT i.indisvalid`
	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bad []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		bad = append(bad, name)
	}
	return bad, rows.Err()
}

// crossMajorOutcome enumerates the possible end states of the cross-major
// migration orchestration (pre-flight scan -> restore -> post-restore
// validation). See runCrossMajorMigration.
type crossMajorOutcome int

const (
	// crossMajorOutcomeSuccess: pre-flight clear, restore succeeded, validation
	// matched. The new cluster carries the old data; the caller should proceed
	// with the post-restore steps (matview refresh, retain backup, remove old
	// dir).
	crossMajorOutcomeSuccess crossMajorOutcome = iota
	// crossMajorOutcomePreflightSkipped: pre-flight detected collation risk
	// (or could not run); restore was not attempted. Fall-back warning:
	// crossMajorPreflightSkippedWarning.
	crossMajorOutcomePreflightSkipped
	// crossMajorOutcomeRestoreFailed: pre-flight clear but restore returned a
	// non-zero exit (or a precursor step failed). Fall-back warning:
	// crossMajorRestoreFailedWarning.
	crossMajorOutcomeRestoreFailed
	// crossMajorOutcomeValidationDiverged: restore succeeded but the
	// post-restore validation pass found divergence (row count, sample-row
	// checksum, or invalid index) - or the validation query itself failed.
	// Fall-back warning: crossMajorValidationDivergedWarning.
	crossMajorOutcomeValidationDiverged
)

func (o crossMajorOutcome) String() string {
	switch o {
	case crossMajorOutcomeSuccess:
		return "Success"
	case crossMajorOutcomePreflightSkipped:
		return "PreflightSkipped"
	case crossMajorOutcomeRestoreFailed:
		return "RestoreFailed"
	case crossMajorOutcomeValidationDiverged:
		return "ValidationDiverged"
	default:
		return fmt.Sprintf("crossMajorOutcome(%d)", int(o))
	}
}

// runCrossMajorMigration runs the pre-flight collation scan, restore, and
// post-restore validation pass that together define the cross-major migration
// policy. The caller (restoreDBBackup in production; the cross-major test
// matrix under go test) supplies:
//
//   - oldConn: open pgx connection to the source (old major) cluster. The
//     scan + validation pass query this. The caller owns and closes oldConn.
//   - runRestore: closure that performs the actual restore of the dump into
//     the new (target major) cluster. Production wraps the TOC-partition +
//     runRestoreUsingList flow; the test wraps a single pg_restore call.
//   - newConnFn: factory that opens a fresh pgx connection to the new
//     cluster, called once after a successful restore for the validation
//     pass. The returned connection is closed inside this function. Returns
//     nil + an error on connect failure (treated as a validation divergence).
//
// The caller is responsible for producing the cause-specific warning from the
// returned outcome (e.g. crossMajorPreflightSkippedWarning) and for the
// fall-back side effects (retain dump, leave old dir in place, etc.) - those
// are intentionally NOT in this helper so it stays unit-testable against two
// real clusters without touching the install-dir filesystem.
//
// This helper exists because exec-2a's test matrix needs to exercise the
// shipped pre-flight and validation functions (runPreflightCollationScan,
// runValidateRestore) end-to-end as the production code calls them, not via a
// harness reimplementation. See migration_xmajor_test.go.
func runCrossMajorMigration(
	ctx context.Context,
	oldConn *pgx.Conn,
	runRestore func() error,
	newConnFn func() (*pgx.Conn, error),
) (crossMajorOutcome, error) {
	// Step 1: the insurance dump is assumed to have already been taken by the
	// caller (production: prepareBackup; test: dumpPublicSchema).

	// Step 2: pre-flight collation scan against the old data.
	risks, perr := runPreflightCollationScan(ctx, oldConn)
	if perr != nil {
		log.Printf("[WARN] cross-major migration: pre-flight scan failed: %v", perr)
		return crossMajorOutcomePreflightSkipped, nil
	}
	if len(risks) > 0 {
		log.Printf("[TRACE] cross-major migration: pre-flight flagged %d collation risk(s); skipping restore", len(risks))
		return crossMajorOutcomePreflightSkipped, nil
	}

	// Step 3: restore.
	if rerr := runRestore(); rerr != nil {
		log.Printf("[WARN] cross-major migration: restore failed: %v", rerr)
		return crossMajorOutcomeRestoreFailed, nil
	}

	// Step 4: post-restore validation, old server still live.
	newConn, nerr := newConnFn()
	if nerr != nil {
		log.Printf("[WARN] cross-major migration: could not connect to new cluster for validation: %v", nerr)
		return crossMajorOutcomeValidationDiverged, nil
	}
	defer newConn.Close(ctx)
	divergences, verr := runValidateRestore(ctx, oldConn, newConn)
	if verr != nil {
		log.Printf("[WARN] cross-major migration: validation query failed: %v", verr)
		return crossMajorOutcomeValidationDiverged, nil
	}
	if len(divergences) > 0 {
		log.Printf("[TRACE] cross-major migration: validation found %d divergence(s); rolling back", len(divergences))
		return crossMajorOutcomeValidationDiverged, nil
	}

	// Step 5: success.
	return crossMajorOutcomeSuccess, nil
}

// retainedOldServer holds the still-running old (source) cluster on a
// cross-major migration. On a same-major (minor) migration prepareBackup tears
// the old server down internally as before. On a cross-major jump it is left
// running and stashed here so restoreDBBackup can run the pre-flight collation
// scan (against the old data) and the post-restore validation pass (old vs new)
// while the old server is still live. restoreDBBackup is responsible for
// stopping it (see stopRetainedOldServer) once those steps complete or the
// fall-back message is emitted.
var retainedOldServer *pgRunningInfo

// connectOldServer opens a pgx connection to the retained old (source) cluster.
func connectOldServer(ctx context.Context) (*pgx.Conn, error) {
	if retainedOldServer == nil {
		return nil, fmt.Errorf("no retained old server to connect to")
	}
	connStr := fmt.Sprintf("host=127.0.0.1 port=%d user=%s dbname=%s sslmode=disable",
		retainedOldServer.port, constants.DatabaseSuperUser, retainedOldServer.dbName)
	return pgx.Connect(ctx, connStr)
}

// stopRetainedOldServer stops the retained old (source) cluster, if any, and
// clears the handle. Safe to call multiple times.
func stopRetainedOldServer(ctx context.Context) {
	if retainedOldServer == nil {
		return
	}
	//nolint:golint,errcheck // best-effort shutdown of the old cluster
	retainedOldServer.stop(ctx)
	retainedOldServer = nil
}

// prepareBackup creates a backup file of the public schema for the current database, if we are migrating
// if a backup was taken, this returns the name of the database that was backed up
func prepareBackup(ctx context.Context) (*string, error) {
	found, location, err := findDifferentPgInstallation(ctx)
	if err != nil {
		log.Println("[TRACE] Error while finding different PG Version:", err)
		return nil, err
	}
	// nothing found - nothing to do
	if !found {
		return nil, nil
	}

	// ensure there is no orphaned instance of postgres running
	// (if the service state file was in-tact, we would already have found it and
	// failed before now with a suitable message
	// - to get here the state file must be missing/invalid, so just kill the postgres process)
	// ignore error - just proceed with installation
	if err := killRunningDbInstance(ctx); err != nil {
		return nil, err
	}

	runConfig, err := startDatabaseInLocation(ctx, location)
	if err != nil {
		log.Printf("[TRACE] Error while starting old db in %s: %v", location, err)
		return nil, err
	}

	// On a cross-major jump the old server must stay up so restoreDBBackup can
	// run the pre-flight collation scan and the post-restore validation pass
	// against the old data while it is still live. For a same-major (minor)
	// migration there is no pre-flight/validation step, so tear it down here as
	// before.
	crossMajor := classifyPgMigration(filepath.Base(location), targetDatabaseVersion) == pgMigrationMajor

	takeErr := takeBackup(ctx, runConfig)
	if takeErr != nil {
		// the dump failed - the old server is no longer needed; tear it down
		// and surface the error.
		//nolint:golint,errcheck // best-effort shutdown
		runConfig.stop(ctx)
		return &runConfig.dbName, takeErr
	}

	if crossMajor {
		// leave the old server running; restoreDBBackup stops it.
		retainedOldServer = runConfig
	} else {
		//nolint:golint,errcheck // this will probably never error - if it does, it's not something we can recover from with code
		runConfig.stop(ctx)
	}

	return &runConfig.dbName, nil
}

// killRunningDbInstance searches for a postgres instance running in the install dir
// and if found tries to kill it
func killRunningDbInstance(ctx context.Context) error {
	processes, err := FindAllSteampipePostgresInstances(ctx)
	if err != nil {
		log.Println("[TRACE] FindAllSteampipePostgresInstances failed with", err)
		return err
	}

	for _, p := range processes {
		cmdLine, err := p.CmdlineWithContext(ctx)
		if err != nil {
			continue
		}

		// check if the name of the process is prefixed with the $STEAMPIPE_INSTALL_DIR
		// that means this is a steampipe service from this installation directory
		if strings.HasPrefix(cmdLine, app_specific.InstallDir) {
			log.Println("[TRACE] Terminating running postgres process")
			if err := p.Kill(); err != nil {
				error_helpers.ShowWarning(fmt.Sprintf("Failed to kill orphan postgres process PID %d", p.Pid))
				return errDbInstanceRunning
			}
		}
	}
	return nil
}

// backup the old pg instance public schema using pg_dump
func takeBackup(ctx context.Context, config *pgRunningInfo) error {
	cmd := pgDumpCmd(
		ctx,
		fmt.Sprintf("--file=%s", filepaths.DatabaseBackupFilePath()),
		fmt.Sprintf("--format=%s", backupFormat),
		// of the public schema only
		"--schema=public",
		// only backup the database used by steampipe
		fmt.Sprintf("--dbname=%s", config.dbName),
		// connection parameters
		"--host=127.0.0.1",
		fmt.Sprintf("--port=%d", config.port),
		fmt.Sprintf("--username=%s", constants.DatabaseSuperUser),
	)
	log.Println("[TRACE] starting pg_dump command:", cmd.String())

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("[TRACE] pg_dump process output:", string(output))
		return err
	}

	return nil
}

// startDatabaseInLocation starts up the postgres binary in a specific installation directory
// returns a pgRunningInfo instance
func startDatabaseInLocation(ctx context.Context, location string) (*pgRunningInfo, error) {
	binaryLocation := filepath.Join(location, "postgres", "bin", "postgres")
	dataLocation := filepath.Join(location, "data")
	port, err := putils.GetNextFreePort()
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(
		ctx,
		binaryLocation,
		// by this time, we are sure that the port is free to listen to
		"-p", fmt.Sprint(port),
		"-c", "listen_addresses=127.0.0.1",
		// NOTE: If quoted, the application name includes the quotes. Worried about
		// having spaces in the APPNAME, but leaving it unquoted since currently
		// the APPNAME is hardcoded to be steampipe.
		"-c", fmt.Sprintf("application_name=%s", app_specific.AppName),
		"-c", fmt.Sprintf("cluster_name=%s", app_specific.AppName),

		// Data Directory
		"-D", dataLocation,
	)

	log.Println("[TRACE]", cmd.String())

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	runConfig := &pgRunningInfo{cmd: cmd, port: port}

	dbName, err := getDatabaseName(ctx, port)
	if err != nil {
		runConfig.stop(ctx)
		return nil, err
	}

	runConfig.dbName = dbName

	return runConfig, nil
}

// findDifferentPgInstallation checks whether the '$STEAMPIPE_INSTALL_DIR/db' directory contains any database installation
// other than desired version.
// it's called from `prepareBackup` to decide whether `pg_dump` needs to run,
// and from `restoreDBBackup` both to classify the old install (same-major vs
// cross-major) and, on a successful same-major restore, to locate the old
// installation for removal. On the cross-major path the old dir is
// deliberately NOT removed (it is the user's only copy of the old data).
func findDifferentPgInstallation(ctx context.Context) (bool, string, error) {
	dbBaseDirectory := filepaths.EnsureDatabaseDir()
	entries, err := os.ReadDir(dbBaseDirectory)
	if err != nil {
		return false, "", err
	}

	// The version this build targets, parsed once for comparison. If it is not
	// valid semver we simply skip the "older than target" guard below.
	targetVersion, targetErr := semver.NewVersion(targetDatabaseVersion)

	// When several old installs are on disk (e.g. fossils left by earlier
	// upgrades), the one to migrate from is the most-recent prior version - the
	// one that was live before this upgrade - NOT whichever happens to sort
	// first. Pick the highest version that is older than the target and that
	// holds real data.
	var bestPath string
	var bestVersion *semver.Version
	for _, de := range entries {
		if !de.IsDir() || de.Name() == targetDatabaseVersion {
			continue
		}
		dir := filepath.Join(dbBaseDirectory, de.Name())

		// A migration source must have both a postgres binary AND an initialised
		// data directory (data/PG_VERSION) - a binary-only dir has nothing to dump.
		hasBinary := files.FileExists(filepath.Join(dir, "postgres", "bin", "postgres"))
		hasData := files.FileExists(filepath.Join(dir, "data", "PG_VERSION"))
		if !hasBinary || !hasData {
			continue
		}

		// The directory name is the embedded-PG version.
		v, err := semver.NewVersion(de.Name())
		if err != nil {
			log.Printf("[TRACE] findDifferentPgInstallation - skipping non-semver db dir %q: %s", de.Name(), err.Error())
			continue
		}
		// Never migrate "down" from a leftover install newer than the target.
		if targetErr == nil && !v.LessThan(targetVersion) {
			continue
		}
		if bestVersion == nil || v.GreaterThan(bestVersion) {
			bestVersion = v
			bestPath = dir
		}
	}

	if bestVersion == nil {
		return false, "", nil
	}
	return true, bestPath, nil
}

// restoreDBBackup loads the back up file into the database
func restoreDBBackup(ctx context.Context) error {
	backupFilePath := filepaths.DatabaseBackupFilePath()
	if !files.FileExists(backupFilePath) {
		// nothing to do here
		return nil
	}
	log.Printf("[TRACE] restoreDBBackup: backup file '%s' found, restoring", backupFilePath)

	// load the db status
	runningInfo, err := GetState()
	if err != nil {
		return err
	}
	if runningInfo == nil {
		return fmt.Errorf("steampipe service is not running")
	}

	// Determine whether the on-disk old install is a same-major (minor) or a
	// cross-major migration. On a cross-major jump we run the five-step
	// best-effort flow (insurance dump -> pre-flight collation scan -> restore
	// -> post-restore validation -> success-or-fall-back) for the public schema,
	// then, on public success and while the old cluster is still live, run the
	// data-tank migration (migrateDataTankSchemasOnStartup); the old data dir is
	// removed only when both commit. The fall-back state (hardened-B: retain the
	// dump and the old data dir, start the service on an empty public schema,
	// warn the user) is rolled to whenever pre-flight detects collation risk,
	// pg_restore returns non-zero, or validation finds divergence.
	var crossMajor bool
	var oldVersion string
	if found, location, ferr := findDifferentPgInstallation(ctx); ferr == nil && found {
		if classifyPgMigration(filepath.Base(location), targetDatabaseVersion) == pgMigrationMajor {
			crossMajor = true
			oldVersion = filepath.Base(location)
		}
	}
	newVersion := targetDatabaseVersion

	// fallBackCrossMajor rolls to the hardened-B fall-back state with a
	// cause-specific warning. The insurance dump is retained and the old data
	// directory is intentionally NOT removed - it is the user's only copy of the
	// old data.
	// KNOWN LIMITATION: the old dir is retained indefinitely and
	// findDifferentPgInstallation returns the first non-target version dir it
	// finds (no recency ordering), so a *subsequent* upgrade could detect this
	// now-stale dir instead of the current one. Acceptable for the embedded DB
	// (low-churn); revisit if multi-generation upgrade chains become common.
	fallBackCrossMajor := func(warning string) error {
		stopRetainedOldServer(ctx)
		if err := retainBackup(ctx); err != nil {
			error_helpers.ShowWarning(fmt.Sprintf("Failed to save backup file: %v", err))
		}
		error_helpers.ShowWarning(warning)
		return nil
	}

	if crossMajor {
		// Step 1: the insurance dump was already taken by prepareBackup and the
		// backup file exists on disk (checked above). It is retained as part of
		// every fall-back path and at the end of the success path.

		// The pre-flight and validation steps need a live connection to the old
		// (source) cluster, which prepareBackup left running on the cross-major
		// path. If it is unavailable (unexpected), fall back conservatively to
		// the pre-flight-skipped state rather than risk an unvalidated restore.
		oldConn, oerr := connectOldServer(ctx)
		if oerr != nil {
			log.Printf("[WARN] cross-major migration: could not connect to old cluster for pre-flight/validation: %v", oerr)
			return fallBackCrossMajor(crossMajorPreflightSkippedWarning(oldVersion, newVersion))
		}

		// runRestore wraps the TOC-partition + runRestoreUsingList machinery
		// (same restore steps as the same-major path) so the orchestration can
		// own restore-failure recovery without duplicating the flow.
		var objectListFile, matviewListFile string
		runRestore := func() error {
			toc, err := getTableOfContentsFromBackup(ctx)
			if err != nil {
				return err
			}
			objectListFile, matviewListFile, err = partitionTableOfContents(ctx, toc)
			if err != nil {
				return err
			}
			return runRestoreUsingList(ctx, runningInfo, objectListFile)
		}
		newConnFn := func() (*pgx.Conn, error) {
			return createMaintenanceClient(ctx, runningInfo.Port)
		}

		outcome, oerr2 := runCrossMajorMigration(ctx, oldConn, runRestore, newConnFn)
		oldConn.Close(ctx)
		// Clean up TOC list files that may have been written by the restore
		// closure regardless of the outcome path taken.
		defer func() {
			if objectListFile != "" {
				os.Remove(objectListFile)
			}
			if matviewListFile != "" {
				os.Remove(matviewListFile)
			}
		}()
		if oerr2 != nil {
			// Internal error in the orchestration helper itself (not a
			// migration outcome). Surface to the caller.
			return oerr2
		}
		switch outcome {
		case crossMajorOutcomePreflightSkipped:
			return fallBackCrossMajor(crossMajorPreflightSkippedWarning(oldVersion, newVersion))
		case crossMajorOutcomeRestoreFailed:
			return fallBackCrossMajor(crossMajorRestoreFailedWarning(oldVersion, newVersion))
		case crossMajorOutcomeValidationDiverged:
			return fallBackCrossMajor(crossMajorValidationDivergedWarning(oldVersion, newVersion))
		case crossMajorOutcomeSuccess:
			// Public-schema restore + validation clean. Before stopping the old
			// server, migrate any data-tank schemas through the shared engine - it
			// needs the old cluster live to read partition topology and stream the
			// data old -> new. On the normal CLI workspace there are no data-tank
			// schemas and this is a clean no-op.
			//
			// The old data directory is removed only after BOTH the public-schema
			// migration AND the data-tank migration confirm full success (the single
			// deletion gate). A data-tank failure preserves the old dir and warns,
			// but does NOT revert the public-schema success: the new version still
			// runs (the 2026-06-08 data-preservation decision).
			dtCommitted := true
			if found, location, ferr := findDifferentPgInstallation(ctx); ferr == nil && found {
				old := oldClusterRef(oldVersion, location, retainedOldServer.dbName, retainedOldServer.port)
				newRef := newClusterRef(runningInfo.Port, runningInfo.Database)
				dtRes, dtErr := migrateDataTankSchemasOnStartup(ctx, old, newRef, filepaths.EnsureDatabaseDir())
				if dtErr != nil || !dtRes.committed {
					// A data-tank failure does NOT revert the public-schema success
					// (the new version still runs). The old data dir + the retained
					// data-tank dump are the two preserved recovery copies.
					dtCommitted = false
					if dtErr != nil {
						log.Printf("[WARN] cross-major data-tank migration failed: %v", dtErr)
					}
					error_helpers.ShowWarning(dataTankMigrationDataPreservedWarning(dtRes.dumpDir))
				}
			}

			// Refresh materialized views (Step 6), retain the backup, and (only if
			// the data-tank migration also fully succeeded) remove the old install
			// dir through the single deletion gate.
			stopRetainedOldServer(ctx)
			if matviewListFile != "" {
				if err := runRestoreUsingList(ctx, runningInfo, matviewListFile); err != nil {
					error_helpers.ShowWarning("Could not REFRESH Materialized Views while restoring data. Please REFRESH manually.")
				}
			}
			if err := retainBackup(ctx); err != nil {
				error_helpers.ShowWarning(fmt.Sprintf("Failed to save backup file: %v", err))
			}
			found, location, err := findDifferentPgInstallation(ctx)
			if err != nil {
				return err
			}
			if found {
				removeOldDataDirOnMigrationSuccess(dtCommitted, location)
			}
			return nil
		}
		return fmt.Errorf("cross-major migration: unknown outcome %v", outcome)
	}

	// ---- Same-major (minor) migration: existing flow ----

	// extract the Table of Contents from the Backup Archive
	toc, err := getTableOfContentsFromBackup(ctx)
	if err != nil {
		return err
	}

	// create separate TableOfContent files - one containing only DB OBJECT CREATION (with static data) instructions and another containing only REFRESH MATERIALIZED VIEW instructions
	objectAndStaticDataListFile, matviewRefreshListFile, err := partitionTableOfContents(ctx, toc)
	if err != nil {
		return err
	}
	defer func() {
		// remove both files before returning
		// if the restoration fails, these will be regenerated at the next run
		os.Remove(objectAndStaticDataListFile)
		os.Remove(matviewRefreshListFile)
	}()

	// restore everything, but don't refresh Materialized views.
	err = runRestoreUsingList(ctx, runningInfo, objectAndStaticDataListFile)
	if err != nil {
		// Same-major restore failed. Do NOT brick the service: retain the
		// insurance dump, keep the old data directory in place, warn the
		// user, and let the service start (the data did not carry over but
		// is recoverable from the retained dump / preserved old directory).
		if rerr := retainBackup(ctx); rerr != nil {
			error_helpers.ShowWarning(fmt.Sprintf("Failed to save backup file: %v", rerr))
		}
		error_helpers.ShowWarning(restoreFailedWarning(targetDatabaseVersion))
		return nil
	}

	//
	// make an attempt at refreshing the materialized views as part of restoration
	// we are doing this separately, since we do not want the whole restoration to fail if we can't refresh
	//
	// we may not be able to restore when the materilized views contain transitive references to unqualified
	// table names
	//
	// since 'pg_dump' always set a blank 'search_path', it will not be able to resolve the aforementioned transitive
	// dependencies and will inevitably fail to refresh
	//
	err = runRestoreUsingList(ctx, runningInfo, matviewRefreshListFile)
	if err != nil {
		//
		// we could not refresh the Materialized views
		// this is probably because the Materialized views
		// contain transitive references to unqualified table names
		//
		// WARN the user.
		//
		error_helpers.ShowWarning("Could not REFRESH Materialized Views while restoring data. Please REFRESH manually.")
	}

	if err := retainBackup(ctx); err != nil {
		error_helpers.ShowWarning(fmt.Sprintf("Failed to save backup file: %v", err))
	}

	// get the location of the other instance which was backed up
	found, location, err := findDifferentPgInstallation(ctx)
	if err != nil {
		return err
	}

	// remove it through the single deletion gate (the only code that removes the
	// old data dir). A same-major restore that reaches here has succeeded, so the
	// removal is unlocked; the behaviour is identical to the previous inline
	// os.RemoveAll.
	if found {
		removeOldDataDirOnMigrationSuccess(true, location)
	}

	return nil
}

func runRestoreUsingList(ctx context.Context, info *RunningDBInstanceInfo, listFile string) error {
	cmd := pgRestoreCmd(
		ctx,
		filepaths.DatabaseBackupFilePath(),
		fmt.Sprintf("--format=%s", backupFormat),
		// only the public schema is backed up
		"--schema=public",
		// Execute the restore as a single transaction (that is, wrap the emitted commands in BEGIN/COMMIT).
		// This ensures that either all the commands complete successfully, or no changes are applied.
		// This option implies --exit-on-error.
		"--single-transaction",
		// Restore only those archive elements that are listed in list-file, and restore them in the order they appear in the file.
		fmt.Sprintf("--use-list=%s", listFile),
		// the database name
		fmt.Sprintf("--dbname=%s", info.Database),
		// connection parameters
		"--host=127.0.0.1",
		fmt.Sprintf("--port=%d", info.Port),
		fmt.Sprintf("--username=%s", constants.DatabaseSuperUser),
	)

	log.Println("[TRACE] pg_restore command:", cmd.String())

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Println("[TRACE] runRestoreUsingList process:", string(output))
		return err
	}

	return nil
}

// partitionTableOfContents writes back the TableOfContents into a two temporary TableOfContents files:
//
// 1. without REFRESH MATERIALIZED VIEWS commands and 2. only REFRESH MATERIALIZED VIEWS commands
//
// This needs to be done because the pg_dump will always set a blank search path in the backup archive
// and backed up MATERIALIZED VIEWS may have functions with unqualified table names
func partitionTableOfContents(ctx context.Context, tableOfContentsOfBackup []string) (string, string, error) {
	onlyRefresh, withoutRefresh := putils.Partition(tableOfContentsOfBackup, func(v string) bool {
		return strings.Contains(strings.ToUpper(v), "MATERIALIZED VIEW DATA")
	})

	withoutFile := filepath.Join(filepaths.EnsureDatabaseDir(), noMatViewRefreshListFileName)
	onlyFile := filepath.Join(filepaths.EnsureDatabaseDir(), onlyMatViewRefreshListFileName)

	err := error_helpers.CombineErrors(
		os.WriteFile(withoutFile, []byte(strings.Join(withoutRefresh, "\n")), 0644),
		os.WriteFile(onlyFile, []byte(strings.Join(onlyRefresh, "\n")), 0644),
	)

	return withoutFile, onlyFile, err
}

// getTableOfContentsFromBackup uses pg_restore to read the TableOfContents from the
// back archive
func getTableOfContentsFromBackup(ctx context.Context) ([]string, error) {
	cmd := pgRestoreCmd(
		ctx,
		filepaths.DatabaseBackupFilePath(),
		fmt.Sprintf("--format=%s", backupFormat),
		// only the public schema is backed up
		"--schema=public",
		"--list",
	)
	log.Println("[TRACE] TableOfContent extraction command: ", cmd.String())

	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(b)))
	scanner.Split(bufio.ScanLines)

	/* start with an extra comment line */
	lines := []string{";"}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ";") {
			// no use of comments
			continue
		}
		lines = append(lines, scanner.Text())
	}
	/* an extra comment line at the end */
	lines = append(lines, ";")

	return lines, err
}

// retainBackup creates a text dump of the backup binary and saves both in the $STEAMPIPE_INSTALL_DIR/backups directory
// the backups are saved as:
//
//	binary: 'database-yyyy-MM-dd-hh-mm-ss.dump'
//	text:   'database-yyyy-MM-dd-hh-mm-ss.sql'
func retainBackup(ctx context.Context) error {
	now := time.Now()
	backupBaseFileName := fmt.Sprintf(
		"database-%s",
		now.Format("2006-01-02-15-04-05"),
	)
	binaryBackupRetentionFileName := fmt.Sprintf("%s.%s", backupBaseFileName, backupDumpFileExtension)
	textBackupRetentionFileName := fmt.Sprintf("%s.%s", backupBaseFileName, backupTextFileExtension)

	backupDir := filepaths.EnsureBackupsDir()
	binaryBackupFilePath := filepath.Join(backupDir, binaryBackupRetentionFileName)
	textBackupFilePath := filepath.Join(backupDir, textBackupRetentionFileName)

	log.Println("[TRACE] moving database back up to", binaryBackupFilePath)
	if err := putils.MoveFile(filepaths.DatabaseBackupFilePath(), binaryBackupFilePath); err != nil {
		return err
	}
	log.Println("[TRACE] converting database back up to", textBackupFilePath)
	txtConvertCmd := pgRestoreCmd(
		ctx,
		binaryBackupFilePath,
		fmt.Sprintf("--file=%s", textBackupFilePath),
	)

	if output, err := txtConvertCmd.CombinedOutput(); err != nil {
		log.Println("[TRACE] pg_restore convertion process output:", string(output))
		return err
	}

	// limit the number of old backups
	trimBackups()

	return nil
}

func pgDumpCmd(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(
		ctx,
		filepaths.PgDumpBinaryExecutablePath(),
		args...,
	)
	cmd.Env = append(os.Environ(), "PGSSLMODE=disable")

	// set the library path for the pg_dump command
	// this is required for the pg_dump to work correctly since we build the pg_dump binary
	// from source(zonkyio does not package it), they are incorrectly linked, so the correct
	// library path must be set before running it
	cmd.Env = append(cmd.Env, fmt.Sprintf("DYLD_LIBRARY_PATH=%s", filepaths.GetDatabaseLibPath()))

	log.Println("[TRACE] pg_dump command:", cmd.String())
	return cmd
}

func pgRestoreCmd(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(
		ctx,
		filepaths.PgRestoreBinaryExecutablePath(),
		args...,
	)
	cmd.Env = append(os.Environ(), "PGSSLMODE=disable")

	// set the library path for the pg_restore command
	// this is required for the pg_restore to work correctly since we build the pg_restore binary
	// from source(zonkyio does not package it), they are incorrectly linked, so the correct
	// library path must be set before running it
	cmd.Env = append(cmd.Env, fmt.Sprintf("DYLD_LIBRARY_PATH=%s", filepaths.GetDatabaseLibPath()))

	log.Println("[TRACE] pg_restore command:", cmd.String())
	return cmd
}

// trimBackups trims the number of backups to the most recent constants.MaxBackups
func trimBackups() {
	backupDir := filepaths.BackupsDir()
	files, err := os.ReadDir(backupDir)
	if err != nil {
		error_helpers.ShowWarning(fmt.Sprintf("Failed to trim backups folder: %s", err.Error()))
		return
	}

	// retain only the .dump files (just to get the unique backups)
	files = putils.Filter(files, func(v fs.DirEntry) bool {
		if v.Type().IsDir() {
			return false
		}
		// retain only the .dump files
		return strings.HasSuffix(v.Name(), backupDumpFileExtension)
	})

	// map to the names of the backups, without extensions
	names := putils.Map(files, func(v fs.DirEntry) string {
		return strings.TrimSuffix(v.Name(), filepath.Ext(v.Name()))
	})

	// just sorting should work, since these names are suffixed by date of the format yyyy-MM-dd-hh-mm-ss
	sort.Strings(names)

	for len(names) > constants.MaxBackups {
		// shift the first element
		trim := names[0]

		// remove the first element from the array
		names = names[1:]

		// get back the names
		dumpFilePath := filepath.Join(backupDir, fmt.Sprintf("%s.%s", trim, backupDumpFileExtension))
		textFilePath := filepath.Join(backupDir, fmt.Sprintf("%s.%s", trim, backupTextFileExtension))

		removeErr := error_helpers.CombineErrors(os.Remove(dumpFilePath), os.Remove(textFilePath))
		if removeErr != nil {
			error_helpers.ShowWarning(fmt.Sprintf("Could not remove backup: %s", trim))
		}
	}
}
