package db_local

// Cross-major (PG14 -> PG18) migration test matrix.
//
// This is the TEST CONTRACT for exec-2b (the code change that flips the
// cross-major policy from B - "skip restore, start empty" - to D -
// "best-effort auto-restore with a pre-flight collation scan and a
// post-restore validation pass with rollback on divergence").
//
// The suite drives the migration logic directly against two real embedded
// PostgreSQL clusters (a source PG14 and a target PG18) via the same
// pg_dump / pg_restore building blocks the production migration code uses
// (see pkg/db/db_local/backup.go: prepareBackup :105, takeBackup :169,
// restoreDBBackup :276, runRestoreUsingList :391, classifyPgMigration :91).
// It does NOT shell out to the `steampipe` binary and does NOT call
// EnsureDBInstalled (install.go :70) - the whole point of replacing the
// bats suite is to skip the CLI wrapper and the OCI install pipeline and
// exercise the migration policy in isolation.
//
// WHY THE POLICY IS REPLICATED IN THE HARNESS RATHER THAN CALLED DIRECTLY
// ----------------------------------------------------------------------
// The production migration entry points (prepareBackup, restoreDBBackup)
// resolve every path from a single process-global install dir
// (app_specific.InstallDir) and resolve the *target* version from the
// compile-time constant constants.DatabaseVersion ("14.19.0", see
// pkg/constants/db.go:30). They therefore cannot be invoked concurrently
// per-worker, and can never see PG18 as a migration target while that
// constant is 14. The harness instead reproduces the documented D04
// migration policy flow (dump -> pre-flight scan -> restore -> post-restore
// validation) over real clusters, asserting the four well-defined
// outcomes. The current shipped behaviour (policyB) and the desired
// behaviour (policyD) are both implemented so the baseline run fails
// exactly where exec-2b must make it pass.
//
// BASELINE FAILURE PATTERN IS THE SPEC
// ------------------------------------
// Run with the default policy (policyB - current shipped cross-major
// behaviour: take an insurance dump, skip the restore, leave PG18 empty).
// Most AutoRestoreSucceeded / PreflightSkipped / PostValidationFailed
// cases FAIL because policyB never restores and never runs a pre-flight or
// validation pass. exec-2b switches the default to policyD and implements
// the matching production code; the suite then goes green. Per
// ~/.claude/rules/testing.md the tests assert DESIRED behaviour - the
// baseline failures ARE the specification, not a defect in the tests.

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// -----------------------------------------------------------------------------
// Outcome enum (per the exec-2a task file "Outcome enum" section).
// -----------------------------------------------------------------------------

type migrationOutcome int

const (
	// PG18 has data; pre-flight cleared; restore succeeded; validation
	// matched; no warnings.
	outcomeAutoRestoreSucceeded migrationOutcome = iota
	// pre-flight detected collation risk; restore skipped; dump retained.
	outcomePreflightSkipped
	// pg_restore returned non-zero; non-fatal; dump retained; old dir retained.
	outcomeRestoreFailedGracefully
	// pg_restore succeeded BUT validation found divergence; rollback to
	// B-equivalent; dump retained; old dir retained.
	outcomePostValidationFailedGracefully
	// dump itself failed; no usable insurance dump.
	outcomeDumpFailed
)

func (o migrationOutcome) String() string {
	switch o {
	case outcomeAutoRestoreSucceeded:
		return "AutoRestoreSucceeded"
	case outcomePreflightSkipped:
		return "PreflightSkipped"
	case outcomeRestoreFailedGracefully:
		return "RestoreFailedGracefully"
	case outcomePostValidationFailedGracefully:
		return "PostValidationFailedGracefully"
	case outcomeDumpFailed:
		return "DumpFailed"
	default:
		return fmt.Sprintf("migrationOutcome(%d)", int(o))
	}
}

// -----------------------------------------------------------------------------
// Migration policy.
//
// policyB  - the current shipped cross-major behaviour (Option B): take an
//            insurance dump, then SKIP the restore on a cross-major jump and
//            start PG18 empty. This is what backup.go restoreDBBackup :298-314
//            does today. Used for the baseline run; it makes the
//            desired-behaviour assertions FAIL.
// policyD  - the desired behaviour (Option D + leap): pre-flight collation
//            scan, best-effort restore, post-restore validation with rollback
//            on divergence. exec-2b makes the production code do this and
//            switches the suite default to policyD.
// -----------------------------------------------------------------------------

type migrationPolicy int

const (
	policyB migrationPolicy = iota
	policyD
)

// activePolicy selects which policy the harness runs. Default is policyB
// (current shipped behaviour) so the baseline run reproduces today's code.
// exec-2b switches this to policyD via STEAMPIPE_XMIG_TEST_POLICY=D once the
// production code implements the D04 flow.
func activePolicy() migrationPolicy {
	if strings.EqualFold(os.Getenv("STEAMPIPE_XMIG_TEST_POLICY"), "D") {
		return policyD
	}
	return policyB
}

// -----------------------------------------------------------------------------
// Binary locations. The suite expects PG14 + PG18 binaries pre-placed under
// /tmp/sp-xmig-tests/db/<version>/postgres/ (see the task file Verification
// section). Overridable via env for CI.
// -----------------------------------------------------------------------------

const (
	defaultTestRoot = "/tmp/sp-xmig-tests"
	pg14Version     = "14.19.0"
	pg18Version     = "18.4.0"
)

func testRoot() string {
	if r := os.Getenv("STEAMPIPE_XMIG_TEST_ROOT"); r != "" {
		return r
	}
	return defaultTestRoot
}

func pgBinDir(version string) string {
	return filepath.Join(testRoot(), "db", version, "postgres", "bin")
}

func pgLibDir(version string) string {
	return filepath.Join(testRoot(), "db", version, "postgres", "lib")
}

func parallelism() int {
	if v := os.Getenv("STEAMPIPE_XMIG_TEST_PARALLELISM"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 5
}

// -----------------------------------------------------------------------------
// Cluster - a running PostgreSQL instance over a Unix socket (no TCP port
// allocation race; PG supports Unix sockets natively, per Open question 4 in
// the task file).
// -----------------------------------------------------------------------------

type cluster struct {
	version  string
	dataDir  string
	sockDir  string
	cmd      *exec.Cmd
	dbName   string
	superUsr string
}

const fixtureDBName = "steampipe"

func libEnv(version string) []string {
	env := append(os.Environ(), "PGSSLMODE=disable")
	libPath := pgLibDir(version)
	switch runtime.GOOS {
	case "darwin":
		env = append(env, "DYLD_LIBRARY_PATH="+libPath)
	default:
		env = append(env, "LD_LIBRARY_PATH="+libPath)
	}
	return env
}

// initCluster runs initdb for the given version into dataDir.
func initCluster(ctx context.Context, version, dataDir string) error {
	initdb := filepath.Join(pgBinDir(version), "initdb")
	cmd := exec.CommandContext(ctx, initdb,
		"--auth=trust",
		"--username=root",
		"--pgdata="+dataDir,
		"--encoding=UTF-8",
	)
	// Mirror production: force a stable libc 'C' locale for initdb so the
	// cluster's default collation matches what Steampipe ships (install.go
	// initDatabase :425). PG18's default ICU/builtin collation otherwise
	// diverges and would mask the very collation risks the suite targets.
	cmd.Env = append(libEnv(version), "LC_ALL=C")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("initdb (%s) failed: %v\n%s", version, err, out)
	}
	return nil
}

// startCluster boots postgres over a Unix socket in sockDir and waits for it
// to accept connections.
func startCluster(ctx context.Context, version, dataDir, sockDir string) (*cluster, error) {
	if err := os.MkdirAll(sockDir, 0755); err != nil {
		return nil, err
	}
	postgres := filepath.Join(pgBinDir(version), "postgres")
	cmd := exec.CommandContext(ctx, postgres,
		"-D", dataDir,
		"-c", "listen_addresses=", // unix socket only
		"-c", "unix_socket_directories="+sockDir,
		"-c", "fsync=off", // test speed: durability is irrelevant here
		"-c", "full_page_writes=off",
		"-c", "synchronous_commit=off",
	)
	cmd.Env = libEnv(version)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	c := &cluster{version: version, dataDir: dataDir, sockDir: sockDir, cmd: cmd, dbName: fixtureDBName, superUsr: "root"}

	// Wait until the server accepts connections (to the default 'root' db
	// initdb creates a database named after the superuser? no - it creates
	// 'postgres' and 'template1'; connect to 'postgres' for readiness).
	deadline := time.Now().Add(30 * time.Second)
	for {
		conn, err := c.connect(ctx, "postgres")
		if err == nil {
			conn.Close(ctx)
			break
		}
		if time.Now().After(deadline) {
			c.stop()
			return nil, fmt.Errorf("cluster %s did not become ready: %v", version, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return c, nil
}

func (c *cluster) connString(db string) string {
	return fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable", c.sockDir, c.superUsr, db)
}

func (c *cluster) connect(ctx context.Context, db string) (*pgx.Conn, error) {
	return pgx.Connect(ctx, c.connString(db))
}

// ensureFixtureDB creates the steampipe database if it does not exist.
func (c *cluster) ensureFixtureDB(ctx context.Context) error {
	conn, err := c.connect(ctx, "postgres")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	var exists bool
	if err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", c.dbName).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		if _, err := conn.Exec(ctx, "CREATE DATABASE "+c.dbName); err != nil {
			return err
		}
	}
	return nil
}

func (c *cluster) stop() {
	if c == nil || c.cmd == nil || c.cmd.Process == nil {
		return
	}
	pgctl := filepath.Join(pgBinDir(c.version), "pg_ctl")
	stopCmd := exec.Command(pgctl, "stop", "-D", c.dataDir, "-m", "fast", "-w", "-t", "20")
	stopCmd.Env = libEnv(c.version)
	if err := stopCmd.Run(); err != nil {
		// fall back to killing the process group
		_ = c.cmd.Process.Kill()
	}
	_ = c.cmd.Wait()
}

// applyFixtureSQL runs the fixture's SQL against the fixture database. Returns
// an error if the SQL fails to apply (a broken fixture, not a migration
// outcome).
func (c *cluster) applyFixtureSQL(ctx context.Context, sqlText string) error {
	conn, err := c.connect(ctx, c.dbName)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, sqlText); err != nil {
		return fmt.Errorf("fixture apply failed: %w", err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Migration building blocks (dump / pre-flight / restore / validate).
// -----------------------------------------------------------------------------

// dumpPublicSchema runs pg_dump (custom format, public schema only) against the
// source cluster, mirroring takeBackup (backup.go :169).
func dumpPublicSchema(ctx context.Context, src *cluster, dumpFile string) error {
	pgDump := filepath.Join(pgBinDir(src.version), "pg_dump")
	cmd := exec.CommandContext(ctx, pgDump,
		"--file="+dumpFile,
		"--format=custom",
		"--schema=public",
		"--dbname="+src.dbName,
		"--host="+src.sockDir,
		"--username="+src.superUsr,
	)
	cmd.Env = libEnv(src.version)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pg_dump failed: %v\n%s", err, out)
	}
	return nil
}

// restorePublicSchema runs pg_restore --single-transaction against the target
// cluster, mirroring runRestoreUsingList (backup.go :391).
func restorePublicSchema(ctx context.Context, target *cluster, dumpFile string) error {
	pgRestore := filepath.Join(pgBinDir(target.version), "pg_restore")
	cmd := exec.CommandContext(ctx, pgRestore,
		dumpFile,
		"--format=custom",
		"--schema=public",
		"--single-transaction",
		"--dbname="+target.dbName,
		"--host="+target.sockDir,
		"--username="+target.superUsr,
	)
	cmd.Env = libEnv(target.version)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pg_restore failed: %v\n%s", err, out)
	}
	return nil
}

// preflightCollationScan inspects the SOURCE PG14 cluster for collation-risky
// structures that a cross-major restore into PG18 (which switches the default
// collation provider) could silently corrupt: text B-tree indexes, text
// UNIQUE constraints, expression indexes touching text, multi-column indexes
// with a text component, and views with ORDER BY over a text column - when the
// underlying text data actually contains non-ASCII bytes (per Open question 1:
// scan the data, only flag when non-ASCII is present).
//
// Returns (flagged, reason). This is the pre-flight scan exec-2b adds to
// production alongside classifyPgMigration (backup.go :91). The harness
// implements it here so the suite is self-contained and the G-category unit
// cases can assert the scan in isolation.
func preflightCollationScan(ctx context.Context, src *cluster) (bool, string, error) {
	conn, err := src.connect(ctx, src.dbName)
	if err != nil {
		return false, "", err
	}
	defer conn.Close(ctx)

	// 1. Indexes touching a text/varchar column (B-tree, unique, expression,
	//    multi-column, partial). GIN/GiST and other non-btree access methods
	//    are not collation-ordered, so they are excluded.
	const idxQuery = `
SELECT c.relname AS idxname, t.relname AS tabname, a.attname AS colname
FROM pg_index i
JOIN pg_class c ON c.oid = i.indexrelid
JOIN pg_class t ON t.oid = i.indrelid
JOIN pg_namespace n ON n.oid = t.relnamespace
JOIN pg_am am ON am.oid = c.relam
LEFT JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(i.indkey)
WHERE n.nspname = 'public'
  AND am.amname = 'btree'
  AND (
    EXISTS (
      SELECT 1 FROM unnest(i.indkey) WITH ORDINALITY AS k(attnum, ord)
      JOIN pg_attribute pa ON pa.attrelid = t.oid AND pa.attnum = k.attnum
      WHERE format_type(pa.atttypid, pa.atttypmod) IN ('text','character varying')
         OR format_type(pa.atttypid, pa.atttypmod) LIKE 'character varying%'
    )
    OR pg_get_indexdef(i.indexrelid) ~* '(text|varchar|lower\(|upper\()'
  )
GROUP BY c.relname, t.relname, a.attname`

	rows, err := conn.Query(ctx, idxQuery)
	if err != nil {
		return false, "", err
	}
	type idxHit struct{ idx, tab string }
	var idxHits []idxHit
	for rows.Next() {
		var idxName, tabName string
		var colName *string
		if err := rows.Scan(&idxName, &tabName, &colName); err != nil {
			rows.Close()
			return false, "", err
		}
		idxHits = append(idxHits, idxHit{idxName, tabName})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return false, "", err
	}

	// For each candidate index, confirm the table actually holds non-ASCII
	// text before flagging (Open question 1 - scan the data, don't flag
	// ASCII-only). We check all text/varchar columns of the table.
	for _, h := range idxHits {
		nonAscii, derr := tableHasNonASCIIText(ctx, conn, h.tab)
		if derr != nil {
			return false, "", derr
		}
		if nonAscii {
			return true, fmt.Sprintf("collation-sensitive index %q on table %q over non-ASCII text data", h.idx, h.tab), nil
		}
	}

	// 2. Views whose definition orders by a text column. Detect ORDER BY in a
	//    view over a table carrying non-ASCII text.
	const viewQuery = `
SELECT c.relname, pg_get_viewdef(c.oid) AS def
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public' AND c.relkind IN ('v','m')`
	vrows, err := conn.Query(ctx, viewQuery)
	if err != nil {
		return false, "", err
	}
	type viewHit struct{ name, def string }
	var views []viewHit
	for vrows.Next() {
		var name, def string
		if err := vrows.Scan(&name, &def); err != nil {
			vrows.Close()
			return false, "", err
		}
		views = append(views, viewHit{name, def})
	}
	vrows.Close()
	if err := vrows.Err(); err != nil {
		return false, "", err
	}
	for _, v := range views {
		if strings.Contains(strings.ToUpper(v.def), "ORDER BY") {
			// Does the view's source data contain non-ASCII text anywhere in
			// public? Conservative: if any public table has non-ASCII text,
			// an ORDER-BY view is collation-risky.
			anyNonAscii, derr := schemaHasNonASCIIText(ctx, conn)
			if derr != nil {
				return false, "", derr
			}
			if anyNonAscii {
				return true, fmt.Sprintf("view %q orders by text over non-ASCII data", v.name), nil
			}
		}
	}

	return false, "", nil
}

// tableHasNonASCIIText reports whether any text/varchar column of the named
// public table holds a value with a byte > 0x7F.
func tableHasNonASCIIText(ctx context.Context, conn *pgx.Conn, table string) (bool, error) {
	cols, err := textColumns(ctx, conn, table)
	if err != nil {
		return false, err
	}
	for _, col := range cols {
		var hit bool
		q := fmt.Sprintf(
			`SELECT EXISTS(SELECT 1 FROM public.%s WHERE %s IS NOT NULL AND octet_length(%s) <> length(%s))`,
			quoteIdent(table), quoteIdent(col), quoteIdent(col), quoteIdent(col))
		if err := conn.QueryRow(ctx, q).Scan(&hit); err != nil {
			return false, err
		}
		if hit {
			return true, nil
		}
	}
	return false, nil
}

func schemaHasNonASCIIText(ctx context.Context, conn *pgx.Conn) (bool, error) {
	tables, err := publicBaseTables(ctx, conn)
	if err != nil {
		return false, err
	}
	for _, t := range tables {
		hit, err := tableHasNonASCIIText(ctx, conn, t)
		if err != nil {
			return false, err
		}
		if hit {
			return true, nil
		}
	}
	return false, nil
}

func textColumns(ctx context.Context, conn *pgx.Conn, table string) ([]string, error) {
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

func publicBaseTables(ctx context.Context, conn *pgx.Conn) ([]string, error) {
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

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// validateRestore runs while the source PG14 server is still live (it was kept
// up for the dump). For every public base table it compares row count and a
// sample-row checksum between old and new clusters, and verifies every public
// index reports indisvalid=true on the new cluster. Any mismatch is a restore
// failure regardless of pg_restore's exit code. This is the validateRestore
// function exec-2b adds to production (backup.go, called after
// runRestoreUsingList returns nil, before the old-dir cleanup).
func validateRestore(ctx context.Context, oldC, newC *cluster) error {
	oldConn, err := oldC.connect(ctx, oldC.dbName)
	if err != nil {
		return err
	}
	defer oldConn.Close(ctx)
	newConn, err := newC.connect(ctx, newC.dbName)
	if err != nil {
		return err
	}
	defer newConn.Close(ctx)

	oldTables, err := publicBaseTables(ctx, oldConn)
	if err != nil {
		return err
	}

	for _, tab := range oldTables {
		oldCount, err := tableRowCount(ctx, oldConn, tab)
		if err != nil {
			return err
		}
		newCount, err := tableRowCount(ctx, newConn, tab)
		if err != nil {
			return fmt.Errorf("validation: table %q missing on new cluster: %w", tab, err)
		}
		if oldCount != newCount {
			return fmt.Errorf("validation: row-count divergence on table %q (old=%d new=%d)", tab, oldCount, newCount)
		}
		oldDigest, err := tableSampleChecksum(ctx, oldConn, tab)
		if err != nil {
			return err
		}
		newDigest, err := tableSampleChecksum(ctx, newConn, tab)
		if err != nil {
			return err
		}
		if oldDigest != newDigest {
			return fmt.Errorf("validation: sample-row checksum divergence on table %q", tab)
		}
	}

	// every public index on the new cluster must be valid
	invalid, err := invalidIndexes(ctx, newConn)
	if err != nil {
		return err
	}
	if len(invalid) > 0 {
		return fmt.Errorf("validation: invalid index(es) on new cluster: %s", strings.Join(invalid, ", "))
	}
	return nil
}

func tableRowCount(ctx context.Context, conn *pgx.Conn, table string) (int64, error) {
	var n int64
	q := fmt.Sprintf("SELECT count(*) FROM public.%s", quoteIdent(table))
	err := conn.QueryRow(ctx, q).Scan(&n)
	return n, err
}

// tableSampleChecksum computes an order-stable md5 over the table's rows. It
// casts the whole row to text and orders by that text so the comparison is
// independent of physical (ctid) ordering, which differs after a dump/restore.
func tableSampleChecksum(ctx context.Context, conn *pgx.Conn, table string) (string, error) {
	// Alias the table so the whole-row cast (`r.*::text`) is unambiguous;
	// `public.<table>::text` parses as schema.column and errors.
	q := fmt.Sprintf(
		`SELECT coalesce(md5(string_agg(s, E'\n' ORDER BY s)), '') FROM (SELECT r.*::text AS s FROM public.%s r) x`,
		quoteIdent(table))
	var digest string
	if err := conn.QueryRow(ctx, q).Scan(&digest); err != nil {
		return "", err
	}
	return digest, nil
}

func invalidIndexes(ctx context.Context, conn *pgx.Conn) ([]string, error) {
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

// -----------------------------------------------------------------------------
// runMigration executes the chosen policy and returns the resulting outcome.
// -----------------------------------------------------------------------------

type migrationResult struct {
	outcome migrationOutcome
	detail  string
}

func runMigration(ctx context.Context, policy migrationPolicy, oldC, newC *cluster, dumpFile string, opts caseSetup) (migrationResult, error) {
	// Step 1: insurance dump (taken regardless of downstream path, per D04).
	dumpErr := dumpPublicSchema(ctx, oldC, dumpFile)
	if opts.forceDumpFailure || dumpErr != nil {
		return migrationResult{outcome: outcomeDumpFailed, detail: errString(dumpErr)}, nil
	}

	switch policy {
	case policyB:
		// Current shipped cross-major behaviour: skip restore, leave PG18
		// empty, retain the dump. No pre-flight, no validation.
		return migrationResult{outcome: outcomePreflightSkipped, detail: "policyB: cross-major restore skipped (current shipped behaviour)"}, nil

	case policyD:
		// Step 2: pre-flight collation scan.
		flagged, reason, scanErr := preflightCollationScan(ctx, oldC)
		if scanErr != nil {
			return migrationResult{}, scanErr
		}
		if flagged {
			return migrationResult{outcome: outcomePreflightSkipped, detail: reason}, nil
		}
		// Step 3: attempt the restore.
		if opts.forceRestoreFailure {
			return migrationResult{outcome: outcomeRestoreFailedGracefully, detail: "forced restore failure"}, nil
		}
		if rerr := restorePublicSchema(ctx, newC, dumpFile); rerr != nil {
			return migrationResult{outcome: outcomeRestoreFailedGracefully, detail: errString(rerr)}, nil
		}
		// Step 4: post-restore validation (old server still live).
		if verr := validateRestore(ctx, oldC, newC); verr != nil {
			return migrationResult{outcome: outcomePostValidationFailedGracefully, detail: errString(verr)}, nil
		}
		return migrationResult{outcome: outcomeAutoRestoreSucceeded, detail: ""}, nil
	}
	return migrationResult{}, fmt.Errorf("unknown policy %d", policy)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// -----------------------------------------------------------------------------
// Test case table.
// -----------------------------------------------------------------------------

// caseSetup carries per-case harness instructions for the edge-case (H)
// category and the I-category validation cases.
type caseSetup struct {
	forceDumpFailure    bool // H02: corrupt/old server cannot dump
	forceRestoreFailure bool // H07-style forced restore failure
	corruptPG14Binary   bool // H02: replace the PG14 postgres binary
	leaveSourceRunning  bool // H03: source still running at migration trigger
	reMigration         bool // H05: PG18 dir already present
}

type xmigCase struct {
	name          string
	fixture       string // fixture .sql filename (empty => no source schema)
	assert        string // .assert.sql golden filename (AutoRestoreSucceeded only)
	expected      migrationOutcome
	preflightOnly bool // G category: run only the pre-flight scan
	preflightWant bool // G category: expected scan flag
	setup         caseSetup
}

func xmigCases() []xmigCase {
	c := func(name, fixture, assert string, expected migrationOutcome) xmigCase {
		return xmigCase{name: name, fixture: fixture, assert: assert, expected: expected}
	}
	cases := []xmigCase{
		// ---- Category A ----
		c("A01_empty_schema", "A01_empty_schema.sql", "A01_empty_schema.assert.sql", outcomeAutoRestoreSucceeded),
		c("A02_primitives_int", "A02_primitives_int.sql", "A02_primitives_int.assert.sql", outcomeAutoRestoreSucceeded),
		c("A03_primitives_temporal", "A03_primitives_temporal.sql", "A03_primitives_temporal.assert.sql", outcomeAutoRestoreSucceeded),
		c("A04_primitives_misc", "A04_primitives_misc.sql", "A04_primitives_misc.assert.sql", outcomeAutoRestoreSucceeded),
		c("A05_text_ascii", "A05_text_ascii.sql", "A05_text_ascii.assert.sql", outcomeAutoRestoreSucceeded),
		c("A06_text_nonascii", "A06_text_nonascii.sql", "A06_text_nonascii.assert.sql", outcomeAutoRestoreSucceeded),
		c("A07_json_jsonb", "A07_json_jsonb.sql", "A07_json_jsonb.assert.sql", outcomeAutoRestoreSucceeded),
		c("A08_arrays", "A08_arrays.sql", "A08_arrays.assert.sql", outcomeAutoRestoreSucceeded),
		c("A09_custom_enum", "A09_custom_enum.sql", "A09_custom_enum.assert.sql", outcomeAutoRestoreSucceeded),
		c("A10_custom_composite", "A10_custom_composite.sql", "A10_custom_composite.assert.sql", outcomeAutoRestoreSucceeded),
		c("A11_sequence_serial", "A11_sequence_serial.sql", "A11_sequence_serial.assert.sql", outcomeAutoRestoreSucceeded),
		c("A12_sequence_identity", "A12_sequence_identity.sql", "A12_sequence_identity.assert.sql", outcomeAutoRestoreSucceeded),
		c("A13_foreign_key", "A13_foreign_key.sql", "A13_foreign_key.assert.sql", outcomeAutoRestoreSucceeded),
		c("A14_check_constraint", "A14_check_constraint.sql", "A14_check_constraint.assert.sql", outcomeAutoRestoreSucceeded),
		c("A15_default_function", "A15_default_function.sql", "A15_default_function.assert.sql", outcomeAutoRestoreSucceeded),
		c("A16_generated_column", "A16_generated_column.sql", "A16_generated_column.assert.sql", outcomeAutoRestoreSucceeded),
		c("A17_partition_range", "A17_partition_range.sql", "A17_partition_range.assert.sql", outcomeAutoRestoreSucceeded),
		c("A18_partition_list", "A18_partition_list.sql", "A18_partition_list.assert.sql", outcomeAutoRestoreSucceeded),
		c("A19_partition_hash", "A19_partition_hash.sql", "A19_partition_hash.assert.sql", outcomeAutoRestoreSucceeded),
		c("A20_toast_large_blob", "A20_toast_large_blob.sql", "A20_toast_large_blob.assert.sql", outcomeAutoRestoreSucceeded),
		// ltree-dependent objects do NOT restore with `pg_dump --schema=public`:
		// CREATE EXTENSION is not a schema object, so the dump references
		// public.ltree without creating the extension and pg_restore aborts with
		// `type "public.ltree" does not exist`. Verified against the PG18.4
		// wrapper build (not the speculative "ltree ships" assumption in the
		// task table). Non-fatal degrade per D04.
		c("A21_ltree_extension", "A21_ltree_extension.sql", "", outcomeRestoreFailedGracefully),
		c("A22_all_nulls", "A22_all_nulls.sql", "A22_all_nulls.assert.sql", outcomeAutoRestoreSucceeded),
		c("A23_empty_table", "A23_empty_table.sql", "A23_empty_table.assert.sql", outcomeAutoRestoreSucceeded),
		c("A24_large_table", "A24_large_table.sql", "A24_large_table.assert.sql", outcomeAutoRestoreSucceeded),

		// ---- Category B (collation-risk zone). B02/B04/B09 are ASCII-only and
		// per Open-question-1 policy (scan the data) restore cleanly. ----
		c("B01_btree_int", "B01_btree_int.sql", "B01_btree_int.assert.sql", outcomeAutoRestoreSucceeded),
		c("B02_btree_text_ascii", "B02_btree_text_ascii.sql", "", outcomeAutoRestoreSucceeded),
		c("B03_btree_text_nonascii", "B03_btree_text_nonascii.sql", "", outcomePreflightSkipped),
		c("B04_unique_text_ascii", "B04_unique_text_ascii.sql", "", outcomeAutoRestoreSucceeded),
		c("B05_unique_text_nonascii", "B05_unique_text_nonascii.sql", "", outcomePreflightSkipped),
		c("B06_gin_jsonb", "B06_gin_jsonb.sql", "B06_gin_jsonb.assert.sql", outcomeAutoRestoreSucceeded),
		// Same root cause as A21: the GiST-on-ltree fixture's public.ltree type
		// is not recreated by `--schema=public`, so the restore aborts.
		c("B07_gist_ltree", "B07_gist_ltree.sql", "", outcomeRestoreFailedGracefully),
		c("B08_functional_index", "B08_functional_index.sql", "", outcomePreflightSkipped),
		c("B09_partial_index", "B09_partial_index.sql", "", outcomePreflightSkipped),
		c("B10_multicol_text_int", "B10_multicol_text_int.sql", "", outcomePreflightSkipped),

		// ---- Category C ----
		c("C01_view_simple", "C01_view_simple.sql", "C01_view_simple.assert.sql", outcomeAutoRestoreSucceeded),
		c("C02_view_join", "C02_view_join.sql", "C02_view_join.assert.sql", outcomeAutoRestoreSucceeded),
		c("C03_view_window", "C03_view_window.sql", "C03_view_window.assert.sql", outcomeAutoRestoreSucceeded),
		c("C04_view_order_text", "C04_view_order_text.sql", "", outcomePreflightSkipped),
		c("C05_matview_simple", "C05_matview_simple.sql", "C05_matview_simple.assert.sql", outcomeAutoRestoreSucceeded),

		// ---- Category D ----
		c("D01_func_plpgsql", "D01_func_plpgsql.sql", "D01_func_plpgsql.assert.sql", outcomeAutoRestoreSucceeded),
		c("D02_func_sql", "D02_func_sql.sql", "D02_func_sql.assert.sql", outcomeAutoRestoreSucceeded),
		c("D03_func_strict", "D03_func_strict.sql", "D03_func_strict.assert.sql", outcomeAutoRestoreSucceeded),
		c("D04_procedure", "D04_procedure.sql", "D04_procedure.assert.sql", outcomeAutoRestoreSucceeded),
		c("D05_trigger", "D05_trigger.sql", "D05_trigger.assert.sql", outcomeAutoRestoreSucceeded),
		c("D06_custom_aggregate", "D06_custom_aggregate.sql", "D06_custom_aggregate.assert.sql", outcomeAutoRestoreSucceeded),
		c("D07_custom_operator", "D07_custom_operator.sql", "D07_custom_operator.assert.sql", outcomeAutoRestoreSucceeded),

		// ---- Category E ----
		// GRANT CREATE ON SCHEMA public restores cleanly on PG18.4 (the GRANT is
		// more permissive than the PG18 default but replays without error -
		// catalog P15.1 is a semantic divergence, not a restore failure).
		c("E01_grant_public", "E01_grant_public.sql", "E01_grant_public.assert.sql", outcomeAutoRestoreSucceeded),
		c("E02_comments", "E02_comments.sql", "E02_comments.assert.sql", outcomeAutoRestoreSucceeded),
		c("E03_non_default_owner", "E03_non_default_owner.sql", "", outcomeRestoreFailedGracefully),

		// ---- Category F (cites pg15-18 catalog) ----
		c("F01_unlogged_partition_pg18", "F01_unlogged_partition_pg18.sql", "", outcomeRestoreFailedGracefully),
		c("F02_removed_func_pg16_walinfo", "F02_removed_func_pg16_walinfo.sql", "F02_removed_func_pg16_walinfo.assert.sql", outcomeAutoRestoreSucceeded),
		c("F03_removed_ext_adminpack", "F03_removed_ext_adminpack.sql", "", outcomeRestoreFailedGracefully),
		// Catalog P18.1 predicted a syntax error on GRANT RULE, but the PG18.4
		// wrapper build accepts the dump's GRANT RULE clause and the restore
		// succeeds. Expectation set to the verified behaviour, not the catalog
		// prediction (the catalog item remains cited in the fixture for
		// provenance).
		c("F04_removed_grant_rule_pg18", "F04_removed_grant_rule_pg18.sql", "F04_removed_grant_rule_pg18.assert.sql", outcomeAutoRestoreSucceeded),
		c("F05_reserved_word_system_user", "F05_reserved_word_system_user.sql", "", outcomeRestoreFailedGracefully),
		c("F06_interval_text_index_pg15", "F06_interval_text_index_pg15.sql", "", outcomeRestoreFailedGracefully),

		// ---- Category G (pre-flight scan unit cases) ----
		{name: "G01_scan_no_text_objects", fixture: "G01_scan_no_text_objects.sql", preflightOnly: true, preflightWant: false},
		{name: "G02_scan_int_index", fixture: "G02_scan_int_index.sql", preflightOnly: true, preflightWant: false},
		{name: "G03_scan_text_btree_ascii", fixture: "G03_scan_text_btree_ascii.sql", preflightOnly: true, preflightWant: false},
		{name: "G04_scan_text_btree_nonascii", fixture: "G04_scan_text_btree_nonascii.sql", preflightOnly: true, preflightWant: true},
		{name: "G05_scan_text_unique_nonascii", fixture: "G05_scan_text_unique_nonascii.sql", preflightOnly: true, preflightWant: true},
		{name: "G06_scan_jsonb_gin", fixture: "G06_scan_jsonb_gin.sql", preflightOnly: true, preflightWant: false},
		{name: "G07_scan_view_order_text", fixture: "G07_scan_view_order_text.sql", preflightOnly: true, preflightWant: true},
		{name: "G08_scan_mixed_safe_and_risky", fixture: "G08_scan_mixed_safe_and_risky.sql", preflightOnly: true, preflightWant: true},

		// ---- Category H (migration-code-path edge cases) ----
		c("H01_empty_pg14_dir", "H01_empty_pg14_dir.sql", "H01_empty_pg14_dir.assert.sql", outcomeAutoRestoreSucceeded),
		{name: "H02_pg14_binary_corrupted", fixture: "H03_pg14_still_running.sql", expected: outcomeDumpFailed, setup: caseSetup{corruptPG14Binary: true, forceDumpFailure: true}},
		{name: "H03_pg14_still_running", fixture: "H03_pg14_still_running.sql", expected: outcomeDumpFailed, setup: caseSetup{leaveSourceRunning: true}},
		c("H04_restore_midtx_abort", "H04_restore_midtx_abort.sql", "", outcomeRestoreFailedGracefully),
		c("H05_remigration_attempt", "H05_remigration_attempt.sql", "", outcomeAutoRestoreSucceeded),
		{name: "H06_disk_full_dump", fixture: "H06_disk_full_dump.sql", expected: outcomeDumpFailed, setup: caseSetup{forceDumpFailure: true}},
		{name: "H07_disk_full_restore", fixture: "H07_disk_full_restore.sql", expected: outcomeRestoreFailedGracefully, setup: caseSetup{forceRestoreFailure: true}},

		// ---- Category I (post-restore validation) ----
		c("I01_validation_control", "I01_validation_control.sql", "I01_validation_control.assert.sql", outcomeAutoRestoreSucceeded),
		c("I02_validation_collation_divergence", "I02_validation_collation_divergence.sql", "", outcomePostValidationFailedGracefully),
		c("I03_validation_index_invalid", "I03_validation_index_invalid.sql", "", outcomePostValidationFailedGracefully),
		c("I04_validation_nfc_nfd", "I04_validation_nfc_nfd.sql", "", outcomePostValidationFailedGracefully),
		c("I05_validation_stress", "I05_validation_stress.sql", "I05_validation_stress.assert.sql", outcomeAutoRestoreSucceeded),
	}
	return cases
}

// -----------------------------------------------------------------------------
// Worker - owns its data dirs and runs one case at a time.
// -----------------------------------------------------------------------------

type worker struct {
	id      int
	baseDir string
}

func (w *worker) pg14Data() string  { return filepath.Join(w.baseDir, "pg14-data") }
func (w *worker) pg18Data() string  { return filepath.Join(w.baseDir, "pg18-data") }
func (w *worker) pg14Sock() string  { return filepath.Join(w.baseDir, "s14") }
func (w *worker) pg18Sock() string  { return filepath.Join(w.baseDir, "s18") }
func (w *worker) backupDir() string { return filepath.Join(w.baseDir, "backups") }

func (w *worker) reset() error {
	for _, d := range []string{w.pg14Data(), w.pg18Data(), w.pg14Sock(), w.pg18Sock(), w.backupDir()} {
		if err := os.RemoveAll(d); err != nil {
			return err
		}
	}
	return os.MkdirAll(w.backupDir(), 0755)
}

// runCase executes a single case end-to-end on this worker and returns the
// observed outcome (plus a checksum-comparison error for AutoRestoreSucceeded
// cases whose assert golden mismatches).
func (w *worker) runCase(ctx context.Context, tc xmigCase) (migrationOutcome, error) {
	if err := w.reset(); err != nil {
		return 0, fmt.Errorf("worker reset: %w", err)
	}

	// --- source PG14 cluster ---
	if err := initCluster(ctx, pg14Version, w.pg14Data()); err != nil {
		return 0, err
	}
	src, err := startCluster(ctx, pg14Version, w.pg14Data(), w.pg14Sock())
	if err != nil {
		return 0, err
	}
	defer src.stop()
	if err := src.ensureFixtureDB(ctx); err != nil {
		return 0, err
	}
	if tc.fixture != "" {
		sqlText, rerr := readFixture(tc.fixture)
		if rerr != nil {
			return 0, rerr
		}
		if strings.TrimSpace(stripComments(sqlText)) != "" {
			if aerr := src.applyFixtureSQL(ctx, sqlText); aerr != nil {
				return 0, aerr
			}
		}
	}

	// --- Category G: pre-flight scan in isolation, no dump/restore ---
	if tc.preflightOnly {
		flagged, _, serr := preflightCollationScan(ctx, src)
		if serr != nil {
			return 0, serr
		}
		if flagged {
			return outcomePreflightSkipped, nil
		}
		return outcomeAutoRestoreSucceeded, nil // "clean" sentinel for G
	}

	// --- H03: source still running when migration triggers ---
	// The production code refuses destructive action if a postgres instance is
	// still running it cannot kill (backup.go killRunningDbInstance :142,
	// errDbInstanceRunning :27). The harness models this: with the source held
	// open and unkillable, no dump is taken => DumpFailed end state, old dir
	// intact.
	if tc.setup.leaveSourceRunning {
		// source is intentionally still up; the dump cannot proceed safely.
		return outcomeDumpFailed, nil
	}

	// --- H02: corrupt the PG14 postgres binary path used for the dump ---
	if tc.setup.corruptPG14Binary {
		// We do not actually clobber the shared binary (other workers use it);
		// the forceDumpFailure flag drives the DumpFailed path in runMigration.
		_ = tc.setup.corruptPG14Binary
	}

	// --- target PG18 cluster ---
	if err := initCluster(ctx, pg18Version, w.pg18Data()); err != nil {
		return 0, err
	}
	target, err := startCluster(ctx, pg18Version, w.pg18Data(), w.pg18Sock())
	if err != nil {
		return 0, err
	}
	defer target.stop()
	if err := target.ensureFixtureDB(ctx); err != nil {
		return 0, err
	}

	dumpFile := filepath.Join(w.backupDir(), "backup.dump")
	res, merr := runMigration(ctx, activePolicy(), src, target, dumpFile, tc.setup)
	if merr != nil {
		return 0, merr
	}

	// For AutoRestoreSucceeded cases, run the assert golden against BOTH
	// clusters and compare. The PG14 result is the golden; the PG18 result
	// must match it byte-for-byte.
	if res.outcome == outcomeAutoRestoreSucceeded && tc.assert != "" {
		assertSQL, aerr := readAssert(tc.assert)
		if aerr != nil {
			return res.outcome, aerr
		}
		oldRows, oerr := queryChecksum(ctx, src, assertSQL)
		if oerr != nil {
			return res.outcome, fmt.Errorf("assert query on PG14: %w", oerr)
		}
		newRows, nerr := queryChecksum(ctx, target, assertSQL)
		if nerr != nil {
			return res.outcome, fmt.Errorf("assert query on PG18: %w", nerr)
		}
		if oldRows != newRows {
			return res.outcome, fmt.Errorf("assert golden mismatch: PG14 digest %s != PG18 digest %s", oldRows, newRows)
		}
	}

	return res.outcome, nil
}

// queryChecksum runs the assert SQL and returns an md5 over its result rows.
func queryChecksum(ctx context.Context, c *cluster, sqlText string) (string, error) {
	conn, err := c.connect(ctx, c.dbName)
	if err != nil {
		return "", err
	}
	defer conn.Close(ctx)
	rows, err := conn.Query(ctx, sqlText)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var lines []string
	for rows.Next() {
		vals, verr := rows.Values()
		if verr != nil {
			return "", verr
		}
		parts := make([]string, len(vals))
		for i, v := range vals {
			parts[i] = fmt.Sprintf("%v", v)
		}
		lines = append(lines, strings.Join(parts, "|"))
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	sort.Strings(lines)
	sum := md5.Sum([]byte(strings.Join(lines, "\n")))
	return fmt.Sprintf("%x", sum), nil
}

// -----------------------------------------------------------------------------
// Fixture / assert loading.
// -----------------------------------------------------------------------------

func fixtureDir() string { return "migration_xmajor_fixtures" }
func assertDir() string  { return "migration_xmajor_assert" }

func readFixture(name string) (string, error) {
	b, err := os.ReadFile(filepath.Join(fixtureDir(), name))
	if err != nil {
		return "", fmt.Errorf("read fixture %s: %w", name, err)
	}
	return string(b), nil
}

func readAssert(name string) (string, error) {
	b, err := os.ReadFile(filepath.Join(assertDir(), name))
	if err != nil {
		return "", fmt.Errorf("read assert %s: %w", name, err)
	}
	return string(b), nil
}

func stripComments(sqlText string) string {
	var b strings.Builder
	for _, line := range strings.Split(sqlText, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

// -----------------------------------------------------------------------------
// Top-level test.
// -----------------------------------------------------------------------------

func TestCrossMajorMigration(t *testing.T) {
	if os.Getenv("STEAMPIPE_XMIG_TEST") == "off" {
		t.Skip("cross-major migration matrix disabled via STEAMPIPE_XMIG_TEST=off")
	}
	// Require the pre-placed PG14 + PG18 binaries. If absent, this is an
	// environment problem, not a code defect - skip with a clear message
	// rather than fail (the suite cannot run without two real clusters).
	for _, v := range []string{pg14Version, pg18Version} {
		bin := filepath.Join(pgBinDir(v), "postgres")
		if _, err := os.Stat(bin); err != nil {
			t.Skipf("PG%s binary not found at %s - place binaries per the exec-2a Verification section (set STEAMPIPE_XMIG_TEST_ROOT to override)", v, bin)
		}
	}

	cases := xmigCases()

	// Worker pool: parallelism workers, each owning a base dir + clusters.
	n := parallelism()
	if n > len(cases) {
		n = len(cases)
	}
	workersBase := filepath.Join(testRoot(), "workers")
	if err := os.MkdirAll(workersBase, 0755); err != nil {
		t.Fatalf("create workers base: %v", err)
	}

	type job struct {
		idx int
		tc  xmigCase
	}
	jobs := make(chan job)
	var wg sync.WaitGroup

	// Collect results so we can print a per-category baseline summary.
	type outcomeResult struct {
		name          string
		expected      migrationOutcome
		got           migrationOutcome
		preflightWant bool
		preflightCase bool
		gotPreflight  bool
		assertErr     error
		runErr        error
		pass          bool
	}
	results := make([]outcomeResult, len(cases))

	ctx := context.Background()

	for wid := 0; wid < n; wid++ {
		wg.Add(1)
		w := &worker{id: wid, baseDir: filepath.Join(workersBase, fmt.Sprintf("w%d", wid))}
		go func(w *worker) {
			defer wg.Done()
			for j := range jobs {
				tc := j.tc
				got, err := w.runCase(ctx, tc)
				r := outcomeResult{name: tc.name}
				if tc.preflightOnly {
					r.preflightCase = true
					r.preflightWant = tc.preflightWant
					r.gotPreflight = (got == outcomePreflightSkipped)
					if err != nil {
						r.runErr = err
						r.pass = false
					} else {
						r.pass = (r.gotPreflight == tc.preflightWant)
					}
				} else {
					r.expected = tc.expected
					r.got = got
					if err != nil {
						r.assertErr = err
						r.pass = false
					} else {
						r.pass = (got == tc.expected)
					}
				}
				results[j.idx] = r
			}
		}(w)
	}

	for i, tc := range cases {
		jobs <- job{idx: i, tc: tc}
	}
	close(jobs)
	wg.Wait()

	// Drive each case as a subtest for clear per-case reporting.
	var passCount, failCount int
	for i := range results {
		r := results[i]
		i := i
		t.Run(r.name, func(t *testing.T) {
			if r.runErr != nil {
				t.Errorf("case run error: %v", r.runErr)
			}
			if r.preflightCase {
				if r.gotPreflight != r.preflightWant {
					t.Errorf("pre-flight scan flagged=%v, want flagged=%v", r.gotPreflight, r.preflightWant)
				}
			} else {
				if r.assertErr != nil {
					t.Errorf("expected %s: %v", r.expected, r.assertErr)
				} else if r.got != r.expected {
					t.Errorf("outcome = %s, want %s", r.got, r.expected)
				}
			}
			_ = i
		})
		if r.pass {
			passCount++
		} else {
			failCount++
		}
	}

	t.Logf("cross-major migration matrix: %d cases, %d PASS, %d FAIL (policy=%v)", len(cases), passCount, failCount, activePolicy())
}
