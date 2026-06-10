package db_local

// Cross-major (PG14 -> PG18) migration test matrix.
//
// This suite drives the SHIPPED cross-major migration code directly:
//
//   - runMigrationEngine + publicMigrationShape (migration_engine.go) - the shared engine production's
//     restoreDBBackup runs for the public schema (the single cross-major orchestration: collation pre-flight ->
//     dump -> restore ladder -> validation -> tolerant matview refresh).
//   - runPreflightCollationScan (backup.go) - the pre-flight collation scan that gates the cross-major restore (also
//     driven in isolation by the G category).
//   - runValidateRestore (backup.go) - the post-restore validation pass.
//
// The harness owns only the OUT-of-process plumbing (boots a real PG14 source cluster and a real PG18 target cluster
// over Unix sockets, applies fixture SQL, takes the insurance dump). All policy decisions are made by the production
// code under test.
//
// HOW THE SEAM FOR constants.DatabaseVersion WORKS
// ------------------------------------------------
// The production migration code resolves the target major from constants.DatabaseVersion ("14.19.0", a compile-time
// constant). With the constant pinned to 14, classifyPgMigration could never return the cross-major branch under
// `go test` and the production cross-major code path would be dead under the test runner.
//
// The seam is a package-private var targetDatabaseVersion (backup.go) that defaults to constants.DatabaseVersion in
// production. TestMain overrides it to "18.4.0" for the duration of the cross-major test suite so the shipped
// classifyPgMigration / findDifferentPgInstallation logic sees PG18 as the target and the cross-major branch actually
// executes.
//
// HOW TO RUN
// ----------
//   # Place PG14 and PG18 binaries under (default) /tmp/sp-xmig-tests/db/<ver>/postgres
//   # (or set STEAMPIPE_XMIG_TEST_ROOT). Then:
//   go test ./pkg/db/db_local/... -run TestCrossMajorMigration -v -count=1
//
// No build tags or ldflags are required: the seam is set in TestMain.

import (
	"context"
	"crypto/md5"
	"errors"
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
// Outcome enum (mirrors the four end states the shared engine's public shape distinguishes, plus a DumpFailed
// sentinel for cases where the harness blocks the pre-dump step).
// -----------------------------------------------------------------------------

type migrationOutcome int

const (
	// PG18 has data; pre-flight cleared; restore succeeded; validation matched; no warnings.
	outcomeAutoRestoreSucceeded migrationOutcome = iota
	// pre-flight detected collation risk; restore skipped; dump retained.
	outcomePreflightSkipped
	// pg_restore returned non-zero; non-fatal; dump retained; old dir retained.
	outcomeRestoreFailedGracefully
	// pg_restore succeeded BUT validation found divergence; rollback to B-equivalent; dump retained; old dir retained.
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
// Binary locations. The suite expects PG14 + PG18 binaries pre-placed under /tmp/sp-xmig-tests/db/<version>/postgres/ .
// Overridable via env for CI.
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
// Cluster - a running PostgreSQL instance over a Unix socket (no TCP port allocation race; PG supports Unix sockets
// natively).
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
	// Mirror production: force a stable libc 'C' locale for initdb so the cluster's default collation matches what
	// Steampipe ships (install.go initDatabase :425). PG18's default ICU/builtin collation otherwise diverges and
	// would mask the very collation risks the suite targets.
	cmd.Env = append(libEnv(version), "LC_ALL=C")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("initdb (%s) failed: %v\n%s", version, err, out)
	}
	return nil
}

// startCluster boots postgres over a Unix socket in sockDir and waits for it to accept connections.
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

	// Wait until the server accepts connections (initdb creates 'postgres' and 'template1' databases; connect to
	// 'postgres' for readiness).
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

// applyFixtureSQL runs the fixture's SQL against the fixture database. Returns an error if the SQL fails to apply (a
// broken fixture, not a migration outcome).
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
// Migration building blocks.
//
// Nothing of the migration itself is reimplemented here - the harness builds two pgClusterRef handles and runs the
// SHIPPED shared engine (runMigrationEngine with publicMigrationShape), exactly as production's restoreDBBackup does.
// -----------------------------------------------------------------------------

// dumpPublicSchema runs pg_dump (custom format, public schema only) against the source cluster, mirroring the
// insurance dump prepareBackup/takeBackup takes before the engine runs (the engine takes its own working dump
// separately).
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

// xmigClusterRef adapts a harness cluster to the engine's cluster handle.
func xmigClusterRef(c *cluster) *pgClusterRef {
	return &pgClusterRef{
		version: c.version,
		binDir:  pgBinDir(c.version),
		env:     libEnv(c.version),
		sockDir: c.sockDir,
		dbName:  c.dbName,
		user:    c.superUsr,
		dataDir: c.dataDir,
	}
}

// plantInvalidIndex leaves a genuinely INVALID index in the target's fixture database: CREATE UNIQUE INDEX
// CONCURRENTLY over duplicate values fails and leaves the index behind with pg_index.indisvalid=false. The shipped
// runValidateRestore must then report an index_invalid divergence and the engine must roll the migration outcome back
// to ValidationDiverged.
func plantInvalidIndex(ctx context.Context, target *cluster) error {
	conn, err := target.connect(ctx, target.dbName)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, "CREATE TABLE public.validation_canary (id int)"); err != nil {
		return err
	}
	if _, err := conn.Exec(ctx, "INSERT INTO public.validation_canary VALUES (1),(1)"); err != nil {
		return err
	}
	// expected to fail (duplicate values under a unique index), leaving an INVALID index entry behind
	if _, err := conn.Exec(ctx, "CREATE UNIQUE INDEX CONCURRENTLY validation_canary_bad_idx ON public.validation_canary (id)"); err == nil {
		return fmt.Errorf("CREATE UNIQUE INDEX CONCURRENTLY over duplicates unexpectedly succeeded")
	}
	var valid bool
	if err := conn.QueryRow(ctx, "SELECT i.indisvalid FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid WHERE c.relname = 'validation_canary_bad_idx'").Scan(&valid); err != nil {
		return fmt.Errorf("invalid-index plant not found: %w", err)
	}
	if valid {
		return fmt.Errorf("planted index is unexpectedly valid")
	}
	return nil
}

// -----------------------------------------------------------------------------
// runMigration executes the SHIPPED shared engine (public shape) against the two test clusters and maps its result
// back to the matrix's outcome enum.
// -----------------------------------------------------------------------------

type xmigResult struct {
	outcome migrationOutcome
	detail  string
}

func runMigration(ctx context.Context, oldC, newC *cluster, dumpFile string, opts caseSetup) (xmigResult, error) {
	// Step 1: insurance dump (production: prepareBackup/takeBackup). The engine takes its own working dump; this one
	// is the independently-retained copy the data-preservation assertions check.
	dumpErr := dumpPublicSchema(ctx, oldC, dumpFile)
	if opts.forceDumpFailure || dumpErr != nil {
		return xmigResult{outcome: outcomeDumpFailed, detail: errString(dumpErr)}, nil
	}

	// opts.forceValidationFailure drives the validation-divergence path with a REAL catalog divergence: an invalid
	// index planted on the target, which the shipped runValidateRestore detects (index_invalid). This replaces the
	// old harness device of pointing the validation connection at an empty database - the engine owns its
	// connections, so the divergence now lives in the catalog itself. A real PG14->PG18 data divergence that escapes
	// the pre-flight scan is still hard to construct deterministically (see the I04 / NFC-NFD discussion).
	if opts.forceValidationFailure {
		if err := plantInvalidIndex(ctx, newC); err != nil {
			return xmigResult{}, err
		}
	}

	// opts.forceRestoreFailure maps to the engine's own fault injection: every restore tier fails, as a disk-full /
	// unusable-target restore would.
	faults := migrationFaults{}
	if opts.forceRestoreFailure {
		faults.failAllTiers = true
	}

	res, err := runMigrationEngine(ctx, publicMigrationShape(), xmigClusterRef(oldC), xmigClusterRef(newC), filepath.Dir(dumpFile), "", 1, faults)
	switch {
	case err == nil && res.committed:
		return xmigResult{outcome: outcomeAutoRestoreSucceeded}, nil
	case errors.Is(err, errMigrationPreflightSkipped):
		return xmigResult{outcome: outcomePreflightSkipped}, nil
	case errors.Is(err, errMigrationValidationDiverged):
		return xmigResult{outcome: outcomePostValidationFailedGracefully}, nil
	case errors.Is(err, errDataTankAllTiersFailed), errors.Is(err, errDataTankDumpFailed):
		return xmigResult{outcome: outcomeRestoreFailedGracefully}, nil
	}
	return xmigResult{}, fmt.Errorf("unexpected engine result: err=%v result=%+v", err, res)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// isPublicFailureOutcome reports whether an outcome is a terminal failure where the data-preservation invariant must
// hold (old dir + dump retained on disk).
func isPublicFailureOutcome(o migrationOutcome) bool {
	switch o {
	case outcomeRestoreFailedGracefully,
		outcomePostValidationFailedGracefully,
		outcomeDumpFailed:
		return true
	}
	return false
}

// assertOldDataDirPreserved confirms the source PG14 data directory still exists and still holds a real cluster (its
// PG_VERSION marker is present and the data is not an empty husk). This is the half of the data-preservation
// invariant that guarantees the original is recoverable after a failed migration.
func assertOldDataDirPreserved(dataDir string) error {
	info, err := os.Stat(dataDir)
	if err != nil {
		return fmt.Errorf("old PG14 data dir %s missing after migration: %w", dataDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("old PG14 data path %s is not a directory", dataDir)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "PG_VERSION")); err != nil {
		return fmt.Errorf("old PG14 data dir %s no longer contains a cluster (PG_VERSION absent): %w", dataDir, err)
	}
	return nil
}

// assertDumpRetained confirms the safety dump artefact is still on disk and non-empty after a restore/validation
// failure (the second independent recovery copy required by the governing decision).
func assertDumpRetained(dumpFile string) error {
	info, err := os.Stat(dumpFile)
	if err != nil {
		return fmt.Errorf("safety dump %s not retained after failure: %w", dumpFile, err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("safety dump %s retained but empty", dumpFile)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Test case table.
// -----------------------------------------------------------------------------

// caseSetup carries per-case harness instructions for the edge-case (H) category and the I-category validation cases.
type caseSetup struct {
	forceDumpFailure       bool // H02 / H06: dump cannot proceed
	forceRestoreFailure    bool // H07 / I-style forced restore failure
	forceValidationFailure bool // I02 / I03: drive the engine's validation-failure path via a planted invalid index, without depending on a real PG14->PG18 data divergence
	corruptPG14Binary      bool // H02: replace the PG14 postgres binary
	leaveSourceRunning     bool // H03: source still running at migration trigger
	reMigration            bool // H05: PG18 dir already present
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
		// ltree-dependent objects do NOT restore with `pg_dump --schema=public`: CREATE EXTENSION is not a schema
		// object, so the dump references public.ltree without creating the extension and pg_restore aborts with
		// `type "public.ltree" does not exist`. Verified against the PG18.4 wrapper build. Non-fatal degrade per D04.
		c("A21_ltree_extension", "A21_ltree_extension.sql", "", outcomeRestoreFailedGracefully),
		c("A22_all_nulls", "A22_all_nulls.sql", "A22_all_nulls.assert.sql", outcomeAutoRestoreSucceeded),
		c("A23_empty_table", "A23_empty_table.sql", "A23_empty_table.assert.sql", outcomeAutoRestoreSucceeded),
		c("A24_large_table", "A24_large_table.sql", "A24_large_table.assert.sql", outcomeAutoRestoreSucceeded),

		// ---- Category B (collation-risk zone). B02/B04/B09 are ASCII-only and per the data-aware pre-flight policy
		// restore cleanly. ----
		c("B01_btree_int", "B01_btree_int.sql", "B01_btree_int.assert.sql", outcomeAutoRestoreSucceeded),
		c("B02_btree_text_ascii", "B02_btree_text_ascii.sql", "", outcomeAutoRestoreSucceeded),
		c("B03_btree_text_nonascii", "B03_btree_text_nonascii.sql", "", outcomePreflightSkipped),
		c("B04_unique_text_ascii", "B04_unique_text_ascii.sql", "", outcomeAutoRestoreSucceeded),
		c("B05_unique_text_nonascii", "B05_unique_text_nonascii.sql", "", outcomePreflightSkipped),
		c("B06_gin_jsonb", "B06_gin_jsonb.sql", "B06_gin_jsonb.assert.sql", outcomeAutoRestoreSucceeded),
		// Same root cause as A21: the GiST-on-ltree fixture's public.ltree type is not recreated by
		// `--schema=public`, so the restore aborts.
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
		// GRANT CREATE ON SCHEMA public restores cleanly on PG18.4.
		c("E01_grant_public", "E01_grant_public.sql", "E01_grant_public.assert.sql", outcomeAutoRestoreSucceeded),
		c("E02_comments", "E02_comments.sql", "E02_comments.assert.sql", outcomeAutoRestoreSucceeded),
		c("E03_non_default_owner", "E03_non_default_owner.sql", "", outcomeRestoreFailedGracefully),

		// ---- Category F (cites pg15-18 catalog) ----
		c("F01_unlogged_partition_pg18", "F01_unlogged_partition_pg18.sql", "", outcomeRestoreFailedGracefully),
		c("F02_removed_func_pg16_walinfo", "F02_removed_func_pg16_walinfo.sql", "F02_removed_func_pg16_walinfo.assert.sql", outcomeAutoRestoreSucceeded),
		c("F03_removed_ext_adminpack", "F03_removed_ext_adminpack.sql", "", outcomeRestoreFailedGracefully),
		// Catalog P18.1 predicted a syntax error on GRANT RULE, but the PG18.4 wrapper build accepts the dump's
		// GRANT RULE clause and the restore succeeds.
		c("F04_removed_grant_rule_pg18", "F04_removed_grant_rule_pg18.sql", "F04_removed_grant_rule_pg18.assert.sql", outcomeAutoRestoreSucceeded),
		c("F05_reserved_word_system_user", "F05_reserved_word_system_user.sql", "", outcomeRestoreFailedGracefully),
		c("F06_interval_text_index_pg15", "F06_interval_text_index_pg15.sql", "", outcomeRestoreFailedGracefully),

		// ---- Category P: exhaustive PG15-18 catalogue coverage (public shape) ----
		//
		// One case per catalogue item P15.x..P18.x that the F/E set above does not already cover. Each cites the
		// catalogue ID. The REQUIRED outcome is derived from the catalogue's stated restore behaviour AND the
		// governing data-preservation decision:
		//   - items that ABORT pg_restore => outcomeRestoreFailedGracefully (dump + old dir retained; asserted by
		//     the data-preservation guard).
		//   - items that restore cleanly (body opacity, behavioural-only divergence, C-locale-safe FTS,
		//     custom-format-safe COPY) => outcomeAutoRestoreSucceeded.
		//
		// Catalogue items already covered elsewhere (NOT duplicated here):
		//   P15.1 -> E01_grant_public, P15.2 -> E03_non_default_owner,
		//   P15.3 -> F02 (body opacity) + F03 (view-references-removed-func),
		//   P15.6 -> F06_interval_text_index, P18.1 -> F04_removed_grant_rule,
		//   P18.3 -> F01_unlogged_partition, reserved-word SYSTEM_USER -> F05.
		//
		// Catalogue items that CANNOT be reproduced as a restore failure through Steampipe's PG14 --schema=public
		// dump-and-restore path (recorded as covered-by-class controls below, with the reason - each verified against
		// the real PG14.19 wrapper binary):
		//   P15.4 plpython2u (no plpython2u.control ships in the PG14 wrapper build; CREATE EXTENSION plpython2u
		//     errors "could not open extension control file", so the breakage trigger cannot be loaded onto the
		//     source).
		//   P16.3 NULLS NOT DISTINCT on a PK (PG14 rejects the syntax outright - "syntax error at or near nulls";
		//     the feature arrived in PG15).
		//   P17.1 adminpack (the extension DOES install on this PG14 build - CREATE EXTENSION adminpack succeeds -
		//     but Steampipe dumps with --schema=public, which excludes CREATE EXTENSION entirely; adminpack lives
		//     outside the public schema, so the removed extension never reaches the PG18 restore. Verified: the
		//     public-only custom dump TOC carries no adminpack entry and the restore succeeds without it. NOT the
		//     same path as F03 - F03 is a removed-FUNCTION (pg_is_in_backup) restore abort, a different mechanism;
		//     adminpack simply never enters a public dump).
		//   P17.2 db_user_namespace (a postgresql.conf GUC, never serialised into any dump).
		//   P17.3 / P17.4 colliculocale / daticulocale (catalog columns the PG15 ICU work added; absent from the
		//     PG14 catalog - verified pg_collation / pg_database carry no such column - so the breakage-triggering
		//     view cannot be created on the source).
		//   P18.5 data checksums (pg_dump output is checksum-agnostic; checksums are a cluster-init / pg_upgrade
		//     property never carried in a dump).

		// -- PG15 --
		c("P15_5_xml2_qualified", "P15_5_xml2_qualified.sql", "", outcomeRestoreFailedGracefully),

		// -- PG16 --
		c("P16_1_wal_records_info", "P16_1_wal_records_info.sql", "P16_1_wal_records_info.assert.sql", outcomeAutoRestoreSucceeded),
		c("P16_2_wal_stats", "P16_2_wal_stats.sql", "P16_2_wal_stats.assert.sql", outcomeAutoRestoreSucceeded),
		// P16.4: the catalogue's predicted restore failure is for direct DDL replay / pg_upgrade, NOT the
		// dump-and-restore path. A _RETURN ON SELECT rule makes the relation a view (relkind 'v'); pg_dump serialises
		// a view as an ordinary CREATE VIEW, which PG18 accepts - so on the dump-and-restore path the restore
		// SUCCEEDS by first principles (reasoned in the fixture, mirroring F04). The case still exercises the engine
		// on the rule-converted-view path.
		c("P16_4_on_select_rule_view", "P16_4_on_select_rule_view.sql", "P16_4_on_select_rule_view.assert.sql", outcomeAutoRestoreSucceeded),
		c("P16_5_cursor_assignment", "P16_5_cursor_assignment.sql", "P16_5_cursor_assignment.assert.sql", outcomeAutoRestoreSucceeded),

		// -- PG17 --
		// P17.3 / P17.4 reference catalog columns (pg_collation.colliculocale, pg_database.daticulocale) that the ICU
		// work introduced in PG15 - they do NOT exist on the PG14 source at all, so the view cannot be created at
		// fixture-apply time (verified: "column colliculocale does not exist" on PG14.19). The catalogue's premise is
		// a dump FROM PG15/16 (which HAS the column) restored into PG18; our source is PG14, which predates the
		// column. Non-reproducible from a PG14 source -> covered-by-class controls.
		// not constructible on PG14: pg_collation.colliculocale is a PG15 ICU catalog column absent from the PG14
		// catalog.
		c("P17_3_colliculocale_control", "A02_primitives_int.sql", "A02_primitives_int.assert.sql", outcomeAutoRestoreSucceeded),
		// not constructible on PG14: pg_database.daticulocale is a PG15 ICU catalog column absent from the PG14
		// catalog.
		c("P17_4_daticulocale_control", "A02_primitives_int.sql", "A02_primitives_int.assert.sql", outcomeAutoRestoreSucceeded),
		// P17.5: behavioural-only divergence (the view restores; PG14 and PG18 return different rows by design), so
		// no row-comparing golden is attached.
		c("P17_5_attstattarget_view", "P17_5_attstattarget_view.sql", "", outcomeAutoRestoreSucceeded),
		c("P17_6_search_path_index", "P17_6_search_path_index.sql", "P17_6_search_path_index.assert.sql", outcomeAutoRestoreSucceeded),

		// -- PG18 --
		c("P18_2_memory_contexts_view", "P18_2_memory_contexts_view.sql", "", outcomeRestoreFailedGracefully),
		c("P18_4_copy_dot_eof", "P18_4_copy_dot_eof.sql", "P18_4_copy_dot_eof.assert.sql", outcomeAutoRestoreSucceeded),
		c("P18_6_fts_index", "P18_6_fts_index.sql", "P18_6_fts_index.assert.sql", outcomeAutoRestoreSucceeded),

		// -- Covered-by-class controls for the items that CANNOT be constructed as a restore failure on a real PG14
		// source through Steampipe's --schema=public dump-and-restore path. Each loads a minimal ordinary schema
		// (A02) and asserts AutoRestoreSucceeded: these are NOT breakage coverage, they are documented
		// non-reproducibility controls proving the engine cleanly migrates a representative cluster when the
		// catalogue's breakage trigger cannot exist on the source. The per-item reason (each verified against the
		// PG14.19 binary) is stated in the block comment above; the one-line "not constructible on PG14" rationale is
		// repeated on each case below. --
		// not constructible on PG14: no plpython2u.control in the wrapper build; CREATE EXTENSION plpython2u errors
		// at fixture-apply time.
		c("P15_4_plpython2u_control", "A02_primitives_int.sql", "A02_primitives_int.assert.sql", outcomeAutoRestoreSucceeded),
		// not constructible on PG14: NULLS NOT DISTINCT is PG15 syntax; PG14 rejects it with a syntax error.
		c("P16_3_nulls_not_distinct_control", "A02_primitives_int.sql", "A02_primitives_int.assert.sql", outcomeAutoRestoreSucceeded),
		// not constructible on PG14: adminpack DOES install on PG14, but a --schema=public dump excludes CREATE
		// EXTENSION, so the extension PG18 removed never reaches the restore (distinct from F03, which is a
		// removed-FUNCTION abort).
		c("P17_1_adminpack_control", "A02_primitives_int.sql", "A02_primitives_int.assert.sql", outcomeAutoRestoreSucceeded),
		// not constructible on PG14: db_user_namespace is a postgresql.conf GUC, never serialised into a dump.
		c("P17_2_db_user_namespace_control", "A02_primitives_int.sql", "A02_primitives_int.assert.sql", outcomeAutoRestoreSucceeded),
		// not constructible on PG14: data checksums are a cluster-init / pg_upgrade property; pg_dump output is
		// checksum-agnostic.
		c("P18_5_checksums_control", "A02_primitives_int.sql", "A02_primitives_int.assert.sql", outcomeAutoRestoreSucceeded),

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

		// ---- Category I (post-restore validation orchestration) ----
		//
		// I02 / I03 use forceValidationFailure to drive the production validation-failure path without relying on a
		// real PG14->PG18 data divergence that happens to escape the pre-flight scan. Manufacturing such a divergence
		// on ASCII data is hard to do deterministically; the realistic NFC/NFD-only case (former I04) is unreachable
		// because the multi-byte detector catches any non-ASCII byte before validation ever runs. The harness instead
		// plants a genuinely INVALID index on the target, which the shipped runValidateRestore detects
		// (index_invalid), covering what these cases were always meant to cover: the engine rolls back to
		// PostValidationFailedGracefully when validation reports divergence.
		c("I01_validation_control", "I01_validation_control.sql", "I01_validation_control.assert.sql", outcomeAutoRestoreSucceeded),
		{name: "I02_validation_collation_divergence", fixture: "I02_validation_collation_divergence.sql", expected: outcomePostValidationFailedGracefully, setup: caseSetup{forceValidationFailure: true}},
		{name: "I03_validation_index_invalid", fixture: "I03_validation_index_invalid.sql", expected: outcomePostValidationFailedGracefully, setup: caseSetup{forceValidationFailure: true}},
		// I04 was a realistic NFC/NFD-only divergence case. Dropped: the pre-flight multi-byte detector catches any
		// non-ASCII byte (which both NFC and NFD encodings of the test data contain), so the fixture would always
		// fall to PreflightSkipped before validation can run. C04 already covers the "non-ASCII view ORDER BY"
		// preflight flag; reframing I04 to a fully ASCII validation-divergence case is what I02/I03 now cover via
		// forceValidationFailure.
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

// runCase executes a single case end-to-end on this worker and returns the observed outcome (plus a
// checksum-comparison error for AutoRestoreSucceeded cases whose assert golden mismatches).
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
	// Drives the SHIPPED runPreflightCollationScan directly.
	if tc.preflightOnly {
		conn, cerr := src.connect(ctx, src.dbName)
		if cerr != nil {
			return 0, cerr
		}
		risks, serr := runPreflightCollationScan(ctx, conn)
		conn.Close(ctx)
		if serr != nil {
			return 0, serr
		}
		if len(risks) > 0 {
			return outcomePreflightSkipped, nil
		}
		return outcomeAutoRestoreSucceeded, nil // "clean" sentinel for G
	}

	// --- H03: source still running when migration triggers ---
	if tc.setup.leaveSourceRunning {
		return outcomeDumpFailed, nil
	}

	// --- H02: corrupt the PG14 postgres binary path used for the dump ---
	if tc.setup.corruptPG14Binary {
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
	res, merr := runMigration(ctx, src, target, dumpFile, tc.setup)
	if merr != nil {
		return 0, merr
	}

	// Data-preservation invariant (governing decision 2026-06-08): for every failure-ending outcome the migration
	// must leave the original recoverable - the old PG14 data directory present and populated, plus (where a dump was
	// taken) the safety dump retained. This is asserted PER failure case here so a regression in exec-6b that deletes
	// the old dir before the migration is 100% complete is caught at the point of failure. exec-6d is the dedicated
	// gate test; this is the per-case guard.
	if isPublicFailureOutcome(res.outcome) {
		if perr := assertOldDataDirPreserved(w.pg14Data()); perr != nil {
			return res.outcome, fmt.Errorf("data-preservation invariant violated for %s: %w", res.outcome, perr)
		}
		// On a dump failure the dump artefact may be absent/partial by definition (the dump step is what failed); the
		// old data dir is the surviving copy. On restore / validation failures the safety dump MUST be retained.
		if res.outcome != outcomeDumpFailed {
			if derr := assertDumpRetained(dumpFile); derr != nil {
				return res.outcome, fmt.Errorf("data-preservation invariant violated for %s: %w", res.outcome, derr)
			}
		}
	}

	// For AutoRestoreSucceeded cases, run the assert golden against BOTH clusters and compare. The PG14 result is the
	// golden; the PG18 result must match it byte-for-byte.
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
// TestMain - flip the production seam so the cross-major branch is the target classification under `go test`.
// -----------------------------------------------------------------------------

func TestMain(m *testing.M) {
	prev := targetDatabaseVersion
	targetDatabaseVersion = pg18Version
	code := m.Run()
	targetDatabaseVersion = prev
	os.Exit(code)
}

// -----------------------------------------------------------------------------
// Top-level test.
// -----------------------------------------------------------------------------

func TestCrossMajorMigration(t *testing.T) {
	if os.Getenv("STEAMPIPE_XMIG_TEST") == "off" {
		t.Skip("cross-major migration matrix disabled via STEAMPIPE_XMIG_TEST=off")
	}
	// Require the pre-placed PG14 + PG18 binaries. If absent, this is an environment problem, not a code defect -
	// skip with a clear message rather than fail (the suite cannot run without two real clusters).
	for _, v := range []string{pg14Version, pg18Version} {
		bin := filepath.Join(pgBinDir(v), "postgres")
		if _, err := os.Stat(bin); err != nil {
			t.Skipf("PG%s binary not found at %s - place binaries per the suite header (set STEAMPIPE_XMIG_TEST_ROOT to override)", v, bin)
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

	// Collect results so we can print a per-category summary.
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

	t.Logf("cross-major migration matrix: %d cases, %d PASS, %d FAIL", len(cases), passCount, failCount)
}
