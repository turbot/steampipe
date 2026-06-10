package db_local

// Data-tank cross-major (PG14 -> PG18) migration test matrix.
//
// This suite is the TEST CONTRACT for the data-tank cross-major migration path (exec-5b). It is written against the
// DESIRED behaviour described in data-tank-storage-patterns.md (exec-5):
//
//   - Each data tank is a <handle> + <handle>-parts schema PAIR. The migration must dump/restore every data-tank schema
//     pair, preserving the declarative PARTITION BY LIST(_cloud_partition) topology and re-attaching every partition.
//   - The light-migration variant applies: no collation pre-flight blocking, no row-level checksum gate. The ONLY
//     catalog risk is the SYSTEM_USER reserved word (F05), which must route the migration to tier 3 rather than
//     blocking it.
//   - The restore is TIERED. Tier 1 (parallel pg_restore) is the normal path; tiers 2-4 escalate as each prior tier
//     fails. The outcome names the tier reached, since the operational cost differs per tier (tier 4 = degraded
//     service). When every tier fails the terminal outcome is dtOutcomeDataPreservedOnDisk: under the 2026-06-08
//     governing decision the old data directory and the safety dump are both kept on disk untouched so nothing is lost
//     (no version-revert). In PRODUCTION such a failure fail-stops startup (restoreDBBackup); the harness here keeps
//     its test cluster running, which is harness mechanics, not the production contract - what these cases assert is
//     the preserved-on-disk invariant.
//
// WHAT THE SUITE DRIVES
// ---------------------
// The matrix drives the PRODUCTION engine. runDataTankMigration below is a thin adapter: it maps the harness's
// dtCluster handles and dtCaseSetup fault-injection flags onto migrateDataTank (migration_datatank.go, itself a shape
// adapter over the shared runMigrationEngine) and folds the engine's result + sentinel errors back onto this suite's
// outcome enum and report. All policy decisions - tier escalation, reserved-word routing, data preservation - are made
// by the production code under test.
//
// HOW TO RUN
// ----------
//   # Place PG14 and PG18 binaries under (default) /tmp/sp-dt-xmig-tests/db/<ver>/postgres
//   # OR set STEAMPIPE_DT_XMIG_TEST_ROOT. (The exec-2a suite's default root
//   # /tmp/sp-xmig-tests already has the binaries; point the env var there.)
//   go test ./pkg/db/db_local/... -run TestDataTankMigration -v -count=1
//
// The harness owns only the out-of-process plumbing (boots a real PG14 source cluster + a real PG18 target cluster over
// Unix sockets, applies fixture SQL, runs pg_dump / pg_restore). Policy decisions belong to exec-5b's code.

import (
	"context"
	"crypto/md5"
	"encoding/json"
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
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// -----------------------------------------------------------------------------
// Outcome enum. Separate type / separate prefix from exec-2a's migrationOutcome because the data-tank value set names
// the restore TIER reached. (Spec: exec-5a task file, "Outcome enum" section.)
// -----------------------------------------------------------------------------

type dataTankMigrationOutcome int

const (
	// Success outcomes - name the tier reached, since the cost of each tier matters operationally (tier 4 is degraded
	// service).
	dtOutcomeAutoRestoreSucceededAtTier1 dataTankMigrationOutcome = iota // parallel pg_restore
	dtOutcomeAutoRestoreSucceededAtTier2                                 // serial pg_restore
	dtOutcomeAutoRestoreSucceededAtTier3                                 // per-table COPY
	// dtOutcomePartialRestoreDataPreserved is NOT a success: the per-partition COPY tier (tier 4) climbed and most rows
	// landed, but at least one partition could not be migrated, so that partition's rows never reached the new cluster.
	// Under the 2026-06-08 governing decision a partial result preserves the original on disk (old PG14 data dir +
	// safety dump) and is reported as a failure outcome - it is NOT a success, because the deletion gate must not fire.
	// The engine signals this via errDataTankPartialRestore with committed=false and oldClusterRetained=true.
	dtOutcomePartialRestoreDataPreserved // per-partition COPY, ≥1 partition failed; data preserved
	// Failure outcomes.
	dtOutcomeDumpFailed
	dtOutcomeRefreshPauseFailed  // could not pause refreshes within budget
	dtOutcomeDiskPreflightFailed // insufficient disk to migrate
	// dtOutcomeDataPreservedOnDisk is the terminal failure outcome under the 2026-06-08 governing decision
	// (data-preservation over version-revert): the migration did NOT complete and the original data is preserved on
	// disk in two independent forms - the untouched old PG14 data directory plus the retained safety dump - so it is
	// recoverable. In PRODUCTION this failure fail-stops startup (restoreDBBackup); the harness's test cluster keeps
	// running, which is harness mechanics, not the production contract. This REPLACES the former
	// dtOutcomeRolledBackToPG14 ("stay on PG14"), which encoded the version-revert policy the decision supersedes. The
	// engine signals this via errDataTankAllTiersFailed with oldClusterRetained=true; the harness asserts the old dir +
	// dump are still on disk.
	dtOutcomeDataPreservedOnDisk
)

func (o dataTankMigrationOutcome) String() string {
	switch o {
	case dtOutcomeAutoRestoreSucceededAtTier1:
		return "AutoRestoreSucceededAtTier1"
	case dtOutcomeAutoRestoreSucceededAtTier2:
		return "AutoRestoreSucceededAtTier2"
	case dtOutcomeAutoRestoreSucceededAtTier3:
		return "AutoRestoreSucceededAtTier3"
	case dtOutcomePartialRestoreDataPreserved:
		return "PartialRestoreDataPreserved"
	case dtOutcomeDumpFailed:
		return "DumpFailed"
	case dtOutcomeRefreshPauseFailed:
		return "RefreshPauseFailed"
	case dtOutcomeDiskPreflightFailed:
		return "DiskPreflightFailed"
	case dtOutcomeDataPreservedOnDisk:
		return "DataPreservedOnDisk"
	default:
		return fmt.Sprintf("dataTankMigrationOutcome(%d)", int(o))
	}
}

// -----------------------------------------------------------------------------
// Binary locations. Mirrors the exec-2a layout under a data-tank-specific root so the two suites don't share a worker
// tree.
// -----------------------------------------------------------------------------

const (
	dtDefaultTestRoot = "/tmp/sp-dt-xmig-tests"
	dtPG14Version     = "14.19.0"
	dtPG18Version     = "18.4.0"
	dtFixtureDBName   = "steampipe"
)

func dtTestRoot() string {
	if r := os.Getenv("STEAMPIPE_DT_XMIG_TEST_ROOT"); r != "" {
		return r
	}
	return dtDefaultTestRoot
}

func dtPGBinDir(version string) string {
	return filepath.Join(dtTestRoot(), "db", version, "postgres", "bin")
}

func dtPGLibDir(version string) string {
	return filepath.Join(dtTestRoot(), "db", version, "postgres", "lib")
}

func dtParallelism() int {
	if v := os.Getenv("STEAMPIPE_DT_XMIG_TEST_PARALLELISM"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 5
}

// dtVolumeMultiplier scales the programmatic DT-B volume fixtures down for CI / dev runs without editing the case
// table. Default 1 = full size. Set STEAMPIPE_DT_XMIG_VOLUME=0.01 to run DT-B at 1% for a smoke pass.
func dtVolumeMultiplier() float64 {
	if v := os.Getenv("STEAMPIPE_DT_XMIG_VOLUME"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			return f
		}
	}
	return 1.0
}

func dtScaleRows(n int) int {
	scaled := int(float64(n) * dtVolumeMultiplier())
	if scaled < 1 {
		scaled = 1
	}
	return scaled
}

// dtScalePartitions scales partition COUNTS by the same volume knob, but only for counts above a small floor - so a
// smoke run (STEAMPIPE_DT_XMIG_VOLUME small) does not spend minutes building tens of thousands of partition tables just
// to load a fixture. At the default multiplier (1.0) the production-scale counts (DT-B5 = 600, DT-B6 = 33,600) are
// preserved exactly, which is what feeds exec-5b's ATTACH-cost budget.
func dtScalePartitions(n int) int {
	if n <= 4 {
		return n // small fixtures keep their exact shape
	}
	scaled := int(float64(n) * dtVolumeMultiplier())
	if scaled < 4 {
		scaled = 4 // keep enough partitions to still exercise multi-partition logic
	}
	return scaled
}

// -----------------------------------------------------------------------------
// dtCluster - a running PostgreSQL instance over a Unix socket.
// -----------------------------------------------------------------------------

type dtCluster struct {
	version  string
	dataDir  string
	sockDir  string
	cmd      *exec.Cmd
	dbName   string
	superUsr string
}

func dtLibEnv(version string) []string {
	env := append(os.Environ(), "PGSSLMODE=disable")
	libPath := dtPGLibDir(version)
	switch runtime.GOOS {
	case "darwin":
		env = append(env, "DYLD_LIBRARY_PATH="+libPath)
	default:
		env = append(env, "LD_LIBRARY_PATH="+libPath)
	}
	return env
}

func dtInitCluster(ctx context.Context, version, dataDir string) error {
	initdb := filepath.Join(dtPGBinDir(version), "initdb")
	cmd := exec.CommandContext(ctx, initdb,
		"--auth=trust",
		"--username=root",
		"--pgdata="+dataDir,
		"--encoding=UTF-8",
	)
	// Mirror production: force libc 'C' locale so the cluster default collation matches what Steampipe ships
	// (install.go initDatabase). PG18's default ICU/builtin collation otherwise diverges.
	cmd.Env = append(dtLibEnv(version), "LC_ALL=C")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("initdb (%s) failed: %v\n%s", version, err, out)
	}
	return nil
}

func dtStartCluster(ctx context.Context, version, dataDir, sockDir string) (*dtCluster, error) {
	if err := os.MkdirAll(sockDir, 0755); err != nil {
		return nil, err
	}
	postgres := filepath.Join(dtPGBinDir(version), "postgres")
	cmd := exec.CommandContext(ctx, postgres,
		"-D", dataDir,
		"-c", "listen_addresses=",
		"-c", "unix_socket_directories="+sockDir,
		"-c", "fsync=off",
		"-c", "full_page_writes=off",
		"-c", "synchronous_commit=off",
		// DT-B6 attaches ~33k partitions; the default max_locks_per_transaction (64) overflows the lock table when a
		// single transaction touches that many relations. Raise it so the fixture can load. (exec-5b's real
		// partition-batch ATTACH spreads this across transactions instead.)
		"-c", "max_locks_per_transaction=50000",
	)
	cmd.Env = dtLibEnv(version)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	c := &dtCluster{version: version, dataDir: dataDir, sockDir: sockDir, cmd: cmd, dbName: dtFixtureDBName, superUsr: "root"}

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

func (c *dtCluster) connString(db string) string {
	return fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable", c.sockDir, c.superUsr, db)
}

func (c *dtCluster) connect(ctx context.Context, db string) (*pgx.Conn, error) {
	return pgx.Connect(ctx, c.connString(db))
}

func (c *dtCluster) ensureFixtureDB(ctx context.Context) error {
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

func (c *dtCluster) stop() {
	if c == nil || c.cmd == nil || c.cmd.Process == nil {
		return
	}
	pgctl := filepath.Join(dtPGBinDir(c.version), "pg_ctl")
	stopCmd := exec.Command(pgctl, "stop", "-D", c.dataDir, "-m", "fast", "-w", "-t", "20")
	stopCmd.Env = dtLibEnv(c.version)
	if err := stopCmd.Run(); err != nil {
		_ = c.cmd.Process.Kill()
	}
	_ = c.cmd.Wait()
}

func (c *dtCluster) applyFixtureSQL(ctx context.Context, sqlText string) error {
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

// dtListDataTankSchemas returns the data-tank schema names present on the source cluster: every <handle> schema and its
// <handle>-parts sibling. This is the enumeration the real migration must drive over (instead of just "public").
func dtListDataTankSchemas(ctx context.Context, c *dtCluster) ([]string, error) {
	conn, err := c.connect(ctx, c.dbName)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)
	rows, err := conn.Query(ctx, `
		SELECT nspname FROM pg_namespace
		WHERE nspname NOT IN ('public', 'pg_catalog', 'information_schema', 'pg_toast')
		  AND nspname NOT LIKE 'pg_temp%'
		  AND nspname NOT LIKE 'pg_toast_temp%'
		ORDER BY nspname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schemas []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		schemas = append(schemas, s)
	}
	return schemas, rows.Err()
}

// -----------------------------------------------------------------------------
// Adapter onto the production engine.
//
// runDataTankMigration maps the harness's dtCluster handles + dtCaseSetup fault-injection flags onto migrateDataTank
// (migration_datatank.go) and folds the engine's result and sentinel errors back onto the suite's outcome enum +
// report. No migration policy lives here.
// -----------------------------------------------------------------------------

// dtMigrationReport carries WHAT the migration actually did, beyond the final outcome enum. The reserved-word /
// tier-escalation / recovery cases assert on this so that an accidental match of the outcome enum does NOT count as a
// pass. The fields are copied straight from the engine's migrationResult, so they record how the engine actually
// routed and escalated.
type dtMigrationReport struct {
	// highestTierAttempted is the last restore tier the migration tried (1..4). Zero means the migration aborted
	// before the restore ladder ran (pre-flight / refresh-pause / dump failure).
	highestTierAttempted int
	// reservedWordRouted is true if a reserved-word column was detected and the migration routed straight to tier 3.
	// The reserved-word cases assert this.
	reservedWordRouted bool
	// oldClusterRetained is true if, on a failure, the migration left the old PG14 data directory in place on disk (the
	// preserved-on-disk guarantee these tests assert; in production the failure also fail-stops startup - no
	// version-revert). DT-E3 / DT-E4 / DT-F5 assert this.
	oldClusterRetained bool
}

// runDataTankMigration is the SEAM exec-5b implements. It adapts the test harness's dtCluster handles + dtCaseSetup
// injection flags onto the real tiered data-tank migration engine (migration_datatank.go) and maps the engine's result
// back onto the suite's outcome enum + report.
func runDataTankMigration(ctx context.Context, oldC, newC *dtCluster, backupDir string, setup dtCaseSetup) (dataTankMigrationOutcome, dtMigrationReport, error) {
	var report dtMigrationReport

	src := dtClusterRef(oldC)
	target := dtClusterRef(newC)
	faults := migrationFaults{
		forceDumpFailure:        setup.forceDumpFailure,
		forceDiskPreflightFail:  setup.forceDiskPreflightFail,
		refreshPauseNotHonoured: setup.refreshDuringDump && !setup.respectPauseHook,
		failTier1:               setup.failTier1,
		failTier2:               setup.failTier2,
		failTier3:               setup.failTier3,
		failTier3MidTable:       setup.failTier3MidTable,
		corruptOnePartition:     setup.corruptOnePartition,
		failAllTiers:            setup.failAllTiers,
		interruptMidRestore:     setup.interruptMidRestore,
		targetUnusable:          setup.pg14WontStart,
	}

	statusPath := filepath.Join(backupDir, "data-tank-migration-status.json")
	res, err := migrateDataTank(ctx, src, target, backupDir, statusPath, dtParallelism(), faults)

	report.reservedWordRouted = res.reservedWordRouted
	report.highestTierAttempted = int(res.tierReached)
	report.oldClusterRetained = res.oldClusterRetained

	if err != nil {
		switch {
		case errors.Is(err, errDataTankDiskPreflight):
			return dtOutcomeDiskPreflightFailed, report, nil
		case errors.Is(err, errDataTankRefreshPause):
			return dtOutcomeRefreshPauseFailed, report, nil
		case errors.Is(err, errDataTankDumpFailed):
			return dtOutcomeDumpFailed, report, nil
		case errors.Is(err, errDataTankPartialRestore):
			// PARTIAL: the per-partition COPY tier climbed and most rows landed, but ≥1 partition could not be
			// migrated. committed=false, so the deletion gate preserves the old PG14 data dir alongside the dump. This
			// is NOT a success.
			return dtOutcomePartialRestoreDataPreserved, report, nil
		case errors.Is(err, errDataTankAllTiersFailed):
			return dtOutcomeDataPreservedOnDisk, report, nil
		default:
			return 0, report, err
		}
	}

	// No error => the engine fully committed (partition failures already routed to errDataTankPartialRestore above).
	// Name the tier reached.
	switch res.tierReached {
	case dtRestoreTier1Parallel:
		return dtOutcomeAutoRestoreSucceededAtTier1, report, nil
	case dtRestoreTier2Serial:
		return dtOutcomeAutoRestoreSucceededAtTier2, report, nil
	case dtRestoreTier3PerTable:
		return dtOutcomeAutoRestoreSucceededAtTier3, report, nil
	case dtRestoreTier4PerPartition:
		// Tier 4 with zero partition failures landed every partition - a full success at the per-partition COPY tier.
		return dtOutcomeAutoRestoreSucceededAtTier3, report, nil
	default:
		return dtOutcomeDataPreservedOnDisk, report, nil
	}
}

// dtClusterRef adapts a test-harness dtCluster into the engine's pgClusterRef.
func dtClusterRef(c *dtCluster) *pgClusterRef {
	return &pgClusterRef{
		version: c.version,
		binDir:  dtPGBinDir(c.version),
		env:     dtLibEnv(c.version),
		sockDir: c.sockDir,
		dbName:  c.dbName,
		user:    c.superUsr,
		dataDir: c.dataDir,
	}
}

// -----------------------------------------------------------------------------
// Volume fixture generation (DT-B). Generated programmatically, never shipped as multi-MB .sql files (acceptance
// criterion: volume fixtures use programmatic row generation).
// -----------------------------------------------------------------------------

// dtGenVolumeSQL builds a single-tank data-tank fixture with `partitions` partitions and `rowsPerPart` rows in each
// (both pre-scaled by the volume multiplier). Tank handle "vol_aws", table "v".
func dtGenVolumeSQL(partitions, rowsPerPart int) string {
	var b strings.Builder
	b.WriteString(`create schema if not exists "vol_aws";`)
	b.WriteString("\n")
	b.WriteString(`create schema if not exists "vol_aws-parts";`)
	b.WriteString("\n")
	b.WriteString(`create table "vol_aws"."v" (
		id bigint, title text, _cloud_partition text, _ctx jsonb,
		constraint v_pk primary key (id, _cloud_partition)
	) partition by list (_cloud_partition);`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`do $$
declare
	p int; pname text;
begin
	for p in 1..%d loop
		pname := 'part_conn_' || lpad(p::text, 6, '0') || '-20260101000000';
		execute format('create table %%I.%%I (like %%I.%%I including all)',
			'vol_aws-parts', pname, 'vol_aws', 'v');
		execute format('alter table %%I.%%I attach partition %%I.%%I for values in (%%L)',
			'vol_aws', 'v', 'vol_aws-parts', pname, pname);
		execute format('insert into %%I.%%I (id, title, _cloud_partition, _ctx)
			select g::bigint, ''r-'' || g, %%L, ''{}''::jsonb
			from generate_series(1, %d) g',
			'vol_aws', 'v', pname);
	end loop;
end $$;`, partitions, rowsPerPart))
	b.WriteString("\n")
	return b.String()
}

// dtGenVolumeSQLMultiTable builds a multi-table volume fixture: `tables` tables, each with `partitionsPerTable`
// partitions and `rowsPerPart` rows. Used by DT-B6 (state-farm workspace aggregate: ~33k partitions across 60 tables).
func dtGenVolumeSQLMultiTable(tables, partitionsPerTable, rowsPerPart int) string {
	var b strings.Builder
	b.WriteString(`create schema if not exists "vol_aws";`)
	b.WriteString("\n")
	b.WriteString(`create schema if not exists "vol_aws-parts";`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`do $$
declare
	t int; p int; tname text; pname text;
begin
	for t in 1..%d loop
		tname := 't_' || lpad(t::text, 3, '0');
		execute format('create table %%I.%%I (
			id bigint, title text, _cloud_partition text, _ctx jsonb,
			constraint %%I primary key (id, _cloud_partition)
		) partition by list (_cloud_partition)',
			'vol_aws', tname, tname || '_pk');
		for p in 1..%d loop
			pname := tname || '_part_' || lpad(p::text, 6, '0') || '-20260101000000';
			execute format('create table %%I.%%I (like %%I.%%I including all)',
				'vol_aws-parts', pname, 'vol_aws', tname);
			execute format('alter table %%I.%%I attach partition %%I.%%I for values in (%%L)',
				'vol_aws', tname, 'vol_aws-parts', pname, pname);
			execute format('insert into %%I.%%I (id, title, _cloud_partition, _ctx)
				select g, ''r-'' || g, %%L, ''{}''::jsonb
				from generate_series(1, %d) g',
				'vol_aws', tname, pname);
		end loop;
	end loop;
end $$;`, tables, partitionsPerTable, rowsPerPart))
	b.WriteString("\n")
	return b.String()
}

// -----------------------------------------------------------------------------
// Test case table.
// -----------------------------------------------------------------------------

// dtCaseSetup carries per-case harness instructions for the failure-injection cases (DT-E operational, DT-F tier
// escalation, DT-D refresh-mid-flight).
type dtCaseSetup struct {
	forceDumpFailure       bool // DT-E1 disk full during dump
	forceDiskPreflightFail bool // DT-E preflight disk check fails

	// DT-F tier injection: force the named tier(s) to fail so the next tier must pick up. The adapter maps these onto
	// the engine's migrationFaults so each tier's failure is simulated deterministically inside the production ladder.
	failTier1           bool // DT-F2: parallel restore fails
	failTier2           bool // DT-F3: serial restore fails
	failTier3           bool // DT-F4 setup half: per-table COPY fails for one partition
	failTier3MidTable   bool // DT-F7: tier 3 dies after half a parent's partitions are fully loaded
	corruptOnePartition bool // DT-F4: exactly one partition's data is unrestorable
	failAllTiers        bool // DT-F5: hostile - every tier fails

	// DT-D refresh-mid-flight.
	refreshDuringDump  bool // DT-D1 / DT-D2
	respectPauseHook   bool // DT-D1 true (queued), DT-D2 false (race)
	swapInFlight       bool // DT-D3: ATTACH PARTITION mid-transaction at migration start
	mgrCleanupInFlight bool // DT-D4: a _mgr_ intermediate-table cleanup is running

	// DT-E lifecycle.
	interruptMidRestore bool // DT-E3: SIGKILL mid-restore, re-run on next start
	pg14WontStart       bool // DT-E4: old cluster binary issue

	// DT-B volume generation (mutually exclusive with `fixture`).
	volPartitions  int
	volRowsPerPart int
	volTables      int // >0 => multi-table generator (DT-B6)
}

type dtCase struct {
	name     string
	fixture  string // fixture .sql filename ("" => generated volume fixture)
	assert   string // .assert.sql golden filename (success cases only)
	expected dataTankMigrationOutcome
	setup    dtCaseSetup

	// Report-level assertions. These make the reserved-word, tier-escalation, and preserve/failure cases assert HOW the
	// outcome was reached, not just the final enum - the report fields are copied from the engine's migrationResult, so
	// an accidental enum match cannot pass without the engine having actually routed/escalated that way.
	wantReservedWordRouted bool // DT-C1 / DT-F6: reserved-word scan routed to tier 3
	wantMinTier            int  // DT-F2..F4: the escalation must reach at least this tier
	wantOldClusterRetained bool // DT-E3 / DT-E4 / DT-F5: a preserve outcome must retain the old PG14 dir

	// loadTest marks the exec-5c LE-* cases. runCase captures wall-clock + disk metrics for these and appends them to
	// the per-run metrics file so the load-test report (output/data-tank-load-test-<date>.md) is built from measured
	// numbers, not estimates.
	loadTest bool
}

func dtCases() []dtCase {
	c := func(name, fixture, assert string, expected dataTankMigrationOutcome) dtCase {
		return dtCase{name: name, fixture: fixture, assert: assert, expected: expected}
	}
	cases := []dtCase{
		// ---- Category DT-A: schema shape ----
		c("DT-A1", "DT-A1_single_table_no_partition.sql", "DT-A1_single_table_no_partition.assert.sql", dtOutcomeAutoRestoreSucceededAtTier1),
		c("DT-A2", "DT-A2_list_partition_4.sql", "DT-A2_list_partition_4.assert.sql", dtOutcomeAutoRestoreSucceededAtTier1),
		c("DT-A3", "DT-A3_list_partition_10.sql", "DT-A3_list_partition_10.assert.sql", dtOutcomeAutoRestoreSucceededAtTier1),
		c("DT-A4", "DT-A4_aggregator_single_part.sql", "DT-A4_aggregator_single_part.assert.sql", dtOutcomeAutoRestoreSucceededAtTier1),
		c("DT-A5", "DT-A5_multiple_data_tanks.sql", "DT-A5_multiple_data_tanks.assert.sql", dtOutcomeAutoRestoreSucceededAtTier1),

		// ---- Category DT-B: volume (programmatic generation) ----
		{name: "DT-B1", expected: dtOutcomeAutoRestoreSucceededAtTier1, setup: dtCaseSetup{volPartitions: 1, volRowsPerPart: dtScaleRows(1000)}},
		{name: "DT-B2", expected: dtOutcomeAutoRestoreSucceededAtTier1, setup: dtCaseSetup{volPartitions: 1, volRowsPerPart: dtScaleRows(100000)}},
		{name: "DT-B3", expected: dtOutcomeAutoRestoreSucceededAtTier1, setup: dtCaseSetup{volPartitions: 1, volRowsPerPart: dtScaleRows(1000000)}},
		{name: "DT-B4", expected: dtOutcomeAutoRestoreSucceededAtTier1, setup: dtCaseSetup{volPartitions: dtScalePartitions(60), volRowsPerPart: dtScaleRows(100000)}},
		// DT-B5: state-farm single-table max (600 partitions x 60k rows = 36M).
		{name: "DT-B5", expected: dtOutcomeAutoRestoreSucceededAtTier1, setup: dtCaseSetup{volPartitions: dtScalePartitions(600), volRowsPerPart: dtScaleRows(60000)}},
		// DT-B6: state-farm workspace aggregate (~33k partitions across 60 tables). 60 tables x 560 partitions ~=
		// 33,600 partitions. Small rows-per-part so the ATTACH cost - not the row volume - is what is exercised.
		{name: "DT-B6", expected: dtOutcomeAutoRestoreSucceededAtTier1, setup: dtCaseSetup{volTables: dtScalePartitions(60), volPartitions: dtScalePartitions(560), volRowsPerPart: dtScaleRows(10)}},

		// ---- Category DT-C: cross-major risk surface (reserved word) ----
		// Reserved-word scan routes directly to tier 3.
		{name: "DT-C1", fixture: "DT-C1_reserved_word_system_user.sql", expected: dtOutcomeAutoRestoreSucceededAtTier3, wantReservedWordRouted: true, wantMinTier: 3},

		// ---- Category DT-D: refresh-mid-flight ----
		{name: "DT-D1", fixture: "DT-D_base.sql", assert: "DT-D_base.assert.sql", expected: dtOutcomeAutoRestoreSucceededAtTier1, setup: dtCaseSetup{refreshDuringDump: true, respectPauseHook: true}},
		// DT-D2: pause hook NOT respected => the orchestration contract must prevent this; the desired outcome is the
		// refresh-pause failure, surfaced rather than silently corrupting the dump.
		{name: "DT-D2", fixture: "DT-D_base.sql", expected: dtOutcomeRefreshPauseFailed, setup: dtCaseSetup{refreshDuringDump: true, respectPauseHook: false}},
		// DT-D3: migration starts while a partition swap is in flight. Desired: the dump captures a consistent
		// (post-swap-or-pre-swap) snapshot and the restore succeeds at tier 1 (ATTACH transactions are atomic
		// per-table).
		{name: "DT-D3", fixture: "DT-D_base.sql", assert: "DT-D_base.assert.sql", expected: dtOutcomeAutoRestoreSucceededAtTier1, setup: dtCaseSetup{swapInFlight: true}},
		// DT-D4: _mgr_ intermediate-table cleanup in flight. Desired: exec-5b detects _mgr_ infix tables and migrates
		// cleanly (skip-or-wait path).
		{name: "DT-D4", fixture: "DT-D4_mgr_intermediate.sql", assert: "DT-D4_mgr_intermediate.assert.sql", expected: dtOutcomeAutoRestoreSucceededAtTier1, setup: dtCaseSetup{mgrCleanupInFlight: true}},

		// ---- Category DT-E: operational edge cases ----
		{name: "DT-E1", fixture: "DT-F_base.sql", expected: dtOutcomeDumpFailed, setup: dtCaseSetup{forceDumpFailure: true}},
		{name: "DT-E2", fixture: "DT-F_base.sql", expected: dtOutcomeDiskPreflightFailed, setup: dtCaseSetup{forceDiskPreflightFail: true}},
		{name: "DT-E3", fixture: "DT-F_base.sql", expected: dtOutcomeDataPreservedOnDisk, setup: dtCaseSetup{interruptMidRestore: true}, wantOldClusterRetained: true},
		{name: "DT-E4", fixture: "DT-F_base.sql", expected: dtOutcomeDataPreservedOnDisk, setup: dtCaseSetup{pg14WontStart: true}, wantOldClusterRetained: true},

		// ---- Category DT-F: fallback tier transitions ----
		{name: "DT-F1", fixture: "DT-F_base.sql", assert: "DT-F_base.assert.sql", expected: dtOutcomeAutoRestoreSucceededAtTier1, wantMinTier: 1},
		{name: "DT-F2", fixture: "DT-F_base.sql", expected: dtOutcomeAutoRestoreSucceededAtTier2, setup: dtCaseSetup{failTier1: true}, wantMinTier: 2},
		{name: "DT-F3", fixture: "DT-F_base.sql", expected: dtOutcomeAutoRestoreSucceededAtTier3, setup: dtCaseSetup{failTier1: true, failTier2: true}, wantMinTier: 3},
		// DT-F4: tiers 1-3 fail, tier 4 (per-partition COPY) climbs but one corrupt partition never migrates - a
		// PARTIAL result, NOT a success. committed=false, the old PG14 data dir + safety dump are preserved (the
		// unmigrated partition survives in the original), the deletion gate does not fire.
		{name: "DT-F4", fixture: "DT-F4_partition_corrupt_base.sql", expected: dtOutcomePartialRestoreDataPreserved, setup: dtCaseSetup{failTier1: true, failTier2: true, failTier3: true, corruptOnePartition: true}, wantMinTier: 4, wantOldClusterRetained: true},
		{name: "DT-F5", fixture: "DT-F_base.sql", expected: dtOutcomeDataPreservedOnDisk, setup: dtCaseSetup{failAllTiers: true}, wantOldClusterRetained: true},
		// DT-F6: reserved-word fixture re-run; assert the TIER reached (3), via the reserved-word escalation path (not
		// the F3 path).
		{name: "DT-F6", fixture: "DT-C1_reserved_word_system_user.sql", expected: dtOutcomeAutoRestoreSucceededAtTier3, wantReservedWordRouted: true},
		// DT-F7: tier 3 dies AFTER fully loading half a parent's partitions; tier 4 then rebuilds that parent from
		// scratch and succeeds. The golden compares source vs target row counts, so any tier-3 leftovers surviving
		// into tier 4 (COPY appends - rows would double) fail the assert.
		{name: "DT-F7", fixture: "DT-A3_list_partition_10.sql", assert: "DT-A3_list_partition_10.assert.sql", expected: dtOutcomeAutoRestoreSucceededAtTier3, setup: dtCaseSetup{failTier1: true, failTier2: true, failTier3MidTable: true}, wantMinTier: 4},
		// DT-F8: a PLAIN (non-partitioned) table lives in the tank schema alongside a partitioned one - the
		// detached-partition-awaiting-cleanup / interrupted-refresh shape. The copy tiers must migrate its rows
		// directly (they live in the table, not in partitions); the golden asserts both tables' contents.
		{name: "DT-F8", fixture: "DT-F8_plain_table_in_tank.sql", assert: "DT-F8_plain_table_in_tank.assert.sql", expected: dtOutcomeAutoRestoreSucceededAtTier3, setup: dtCaseSetup{failTier1: true, failTier2: true}, wantMinTier: 3},

		// ---- Category DT-P: PG15-18 catalogue coverage (data-tank shape) ----
		//
		// Data tank's schema surface is deliberately narrow: ordinary partitioned tables (PARTITION BY
		// LIST(_cloud_partition)) with bigint/text/jsonb columns and a single primary-key index. It has NO functions,
		// views, rules, extensions, procedural languages, expression/partial/GIN/GiST indexes, or public-schema GRANTs.
		// So most catalogue items simply CANNOT appear in a data-tank schema. The catalogue items that DO map to the
		// data shape get a real fixture; every other item is non-applicable and is covered by the DT-P-NonApplicable-*
		// control (a plain tank that must clean-migrate, proving the engine is not tripped by something irrelevant).
		//
		// Data-shape-relevant catalogue items:
		//   - SYSTEM_USER reserved word -> DT-C1 / DT-F6 (already above): the one true cross-major risk for data tank;
		//     routes to tier 3.
		//   - P18.3 unlogged partitioned table -> DT-P18_3: a data tank could be created UNLOGGED on PG14; pg_restore
		//     of the unlogged DDL fails on PG18, so the tiered engine must escalate to the per-table COPY tier (tier
		//     3), which recreates the parent LOGGED and lands the data. No fault flags: the catalogue trigger itself
		//     forces the tier-1/2 failure.
		//   - P18.4 \. -in-CSV -> DT-P18_4: a text value containing "\." round-trips because the engine streams
		//     partition data with binary COPY; tier 1.
		{name: "DT-P18_3_unlogged_tank", fixture: "DT-P18_3_unlogged_tank.sql", assert: "DT-P18_3_unlogged_tank.assert.sql", expected: dtOutcomeAutoRestoreSucceededAtTier3, wantMinTier: 3},
		{name: "DT-P18_4_copy_dot_eof_tank", fixture: "DT-P18_4_copy_dot_eof_tank.sql", assert: "DT-P18_4_copy_dot_eof_tank.assert.sql", expected: dtOutcomeAutoRestoreSucceededAtTier1, wantMinTier: 1},

		// Non-applicable control. Stands in for every catalogue item that cannot occur in a data-tank schema: P15.1,
		// P15.2, P15.4, P15.5, P15.6, P16.1, P16.2, P16.3, P16.4, P16.5, P17.1, P17.2, P17.3, P17.4, P17.5, P17.6,
		// P18.1, P18.2, P18.5, P18.6. Each of those requires a function / view / rule / extension / language /
		// expression-index / public-GRANT that data tank never has. A plain tank must clean-migrate at tier 1: the
		// engine is not tripped by the absence of these triggers.
		c("DT-P-NonApplicable-control", "DT-A1_single_table_no_partition.sql", "DT-A1_single_table_no_partition.assert.sql", dtOutcomeAutoRestoreSucceededAtTier1),
	}
	cases = append(cases, dtLoadCases()...)
	return cases
}

// dtLoadCases returns the LE-1..LE-11 load / resilience / tier-escalation cases (exec-5c). They reuse the same harness,
// volume generators and fault-injection flags as the DT-* matrix but at production-calibrated scale (state-farm: 600
// partitions in the largest single table, ~33,600 partitions per workspace). Every LE case carries loadTest=true so
// runCase captures wall-clock + disk metrics into the per-run metrics file (see dtRecordMetrics).
//
// Scale knobs (LE_* volume): the LE-3-derived cases (LE-3, LE-5..LE-11) all use the same 600x60k = 36M-row single-tank
// fixture so their wall-clocks are directly comparable (LE-8 = 1.5xLE-3, LE-9 = 2xLE-3, etc. budgets in the task).
func dtLoadCases() []dtCase {
	const (
		le3Partitions  = 600   // state-farm largest single table
		le3RowsPerPart = 60000 // 600 x 60k = 36M rows
	)
	le3Setup := func() dtCaseSetup {
		return dtCaseSetup{volPartitions: dtScalePartitions(le3Partitions), volRowsPerPart: dtScaleRows(le3RowsPerPart)}
	}
	withFlags := func(base dtCaseSetup, mut func(*dtCaseSetup)) dtCaseSetup {
		mut(&base)
		return base
	}
	return []dtCase{
		// LE-1: single tank, 1M rows, single partition. Tier 1 happy path.
		{name: "LE-1", loadTest: true, expected: dtOutcomeAutoRestoreSucceededAtTier1, wantMinTier: 1,
			setup: dtCaseSetup{volPartitions: 1, volRowsPerPart: dtScaleRows(1000000)}},
		// LE-2: 60 partitions x 100k = 6M rows (mid-size single-table). Tier 1.
		{name: "LE-2", loadTest: true, expected: dtOutcomeAutoRestoreSucceededAtTier1, wantMinTier: 1,
			setup: dtCaseSetup{volPartitions: dtScalePartitions(60), volRowsPerPart: dtScaleRows(100000)}},
		// LE-3: 600 partitions x 60k = 36M rows (state-farm largest single table). Tier 1.
		{name: "LE-3", loadTest: true, expected: dtOutcomeAutoRestoreSucceededAtTier1, wantMinTier: 1, setup: le3Setup()},
		// LE-4: state-farm workspace aggregate: ~33,600 partitions across 60 tables. Tier 1.
		{name: "LE-4", loadTest: true, expected: dtOutcomeAutoRestoreSucceededAtTier1, wantMinTier: 1,
			setup: dtCaseSetup{volTables: dtScalePartitions(60), volPartitions: dtScalePartitions(560), volRowsPerPart: dtScaleRows(10)}},
		// LE-5: LE-3 + concurrent refresh attempt during migration, pause hook honoured. Refresh-pause respected =>
		// migration completes at tier 1.
		{name: "LE-5", loadTest: true, expected: dtOutcomeAutoRestoreSucceededAtTier1, wantMinTier: 1,
			setup: withFlags(le3Setup(), func(s *dtCaseSetup) { s.refreshDuringDump = true; s.respectPauseHook = true })},
		// LE-6: LE-3 + simulated disk pressure. Disk pre-flight aborts cleanly; PG14 retained.
		{name: "LE-6", loadTest: true, expected: dtOutcomeDiskPreflightFailed, wantOldClusterRetained: true,
			setup: withFlags(le3Setup(), func(s *dtCaseSetup) { s.forceDiskPreflightFail = true })},
		// LE-7: LE-3 + SIGKILL mid-restore. Old data dir + dump preserved on disk, old cluster intact (in production
		// this failure fail-stops startup; the harness's running test cluster is harness mechanics).
		{name: "LE-7", loadTest: true, expected: dtOutcomeDataPreservedOnDisk, wantOldClusterRetained: true,
			setup: withFlags(le3Setup(), func(s *dtCaseSetup) { s.interruptMidRestore = true })},
		// LE-8: LE-3 + tier-1 failure injected. Tier 2 (serial pg_restore) picks up. Budget 1.5xLE-3.
		{name: "LE-8", loadTest: true, expected: dtOutcomeAutoRestoreSucceededAtTier2, wantMinTier: 2,
			setup: withFlags(le3Setup(), func(s *dtCaseSetup) { s.failTier1 = true })},
		// LE-9: LE-3 + tier-1+2 failure. Tier 3 (per-table COPY) picks up. Budget 2xLE-3.
		{name: "LE-9", loadTest: true, expected: dtOutcomeAutoRestoreSucceededAtTier3, wantMinTier: 3,
			setup: withFlags(le3Setup(), func(s *dtCaseSetup) { s.failTier1 = true; s.failTier2 = true })},
		// LE-10: LE-3 + tier-1+2+3 failure + one corrupt partition. Tier 4 lands the rest, but the one bad partition
		// never migrates - a PARTIAL result, NOT a success: committed=false, the old PG14 data dir + safety dump are
		// preserved (the unmigrated partition survives in the original), the deletion gate does not fire.
		{name: "LE-10", loadTest: true, expected: dtOutcomePartialRestoreDataPreserved, wantMinTier: 4, wantOldClusterRetained: true,
			setup: withFlags(le3Setup(), func(s *dtCaseSetup) {
				s.failTier1 = true
				s.failTier2 = true
				s.failTier3 = true
				s.corruptOnePartition = true
			})},
		// LE-11: LE-3 + catalog corruption that defeats all tiers. Aborts cleanly; PG14 stays runnable.
		{name: "LE-11", loadTest: true, expected: dtOutcomeDataPreservedOnDisk, wantOldClusterRetained: true,
			setup: withFlags(le3Setup(), func(s *dtCaseSetup) { s.failAllTiers = true })},
		// LE-MEGA: the combined worst case - all three stress axes maxed at once, which no single other case does. LE-3
		// maxes rows in ONE table (36M / 600 partitions); LE-4 maxes tables+partitions but with tiny rows (~33.6k
		// partitions x 10). LE-MEGA combines them: 60 tables x 600 partitions x 1,000 rows = 36,000 partitions and 36M
		// rows spread across 60 tables - the partition/table fan-out of the aggregate AND the row volume of the biggest
		// single table, together. Strictly heavier than either (same row volume, more restore units, more per-table +
		// ATTACH overhead).
		//
		// The 1,000 rows/partition is a stress-test choice, not a measured client number - bump volMegaRowsPerPart to
		// match a real workspace if you have it. At full volume this is a multi-minute offline run; the
		// STEAMPIPE_DT_XMIG_VOLUME knob scales every axis down (0.01 -> 4 tables x 6 partitions x 10 rows) so
		// normal/smoke runs stay fast. Expected: clean tier-1 restore.
		{name: "LE-MEGA", loadTest: true, expected: dtOutcomeAutoRestoreSucceededAtTier1, wantMinTier: 1,
			setup: dtCaseSetup{
				volTables:      dtScalePartitions(60),
				volPartitions:  dtScalePartitions(600),
				volRowsPerPart: dtScaleRows(volMegaRowsPerPart),
			}},
	}
}

// volMegaRowsPerPart is the rows-per-partition for the LE-MEGA combined stress case. A stress-test target, not a
// measured client figure; raise it to match a real largest-workspace shape. At 1,000 the case is 60 tables x 600
// partitions x 1,000 rows = 36M rows over 36,000 partitions at full volume.
const volMegaRowsPerPart = 1000

// -----------------------------------------------------------------------------
// Worker - owns its data dirs and runs one case at a time.
// -----------------------------------------------------------------------------

type dtWorker struct {
	id      int
	baseDir string
}

func (w *dtWorker) pg14Data() string  { return filepath.Join(w.baseDir, "pg14-data") }
func (w *dtWorker) pg18Data() string  { return filepath.Join(w.baseDir, "pg18-data") }
func (w *dtWorker) pg14Sock() string  { return filepath.Join(w.baseDir, "s14") }
func (w *dtWorker) pg18Sock() string  { return filepath.Join(w.baseDir, "s18") }
func (w *dtWorker) backupDir() string { return filepath.Join(w.baseDir, "backups") }

func (w *dtWorker) reset() error {
	for _, d := range []string{w.pg14Data(), w.pg18Data(), w.pg14Sock(), w.pg18Sock(), w.backupDir()} {
		if err := os.RemoveAll(d); err != nil {
			return err
		}
	}
	return os.MkdirAll(w.backupDir(), 0755)
}

func (w *dtWorker) fixtureSQLFor(tc dtCase) (string, error) {
	if tc.setup.volTables > 0 {
		return dtGenVolumeSQLMultiTable(tc.setup.volTables, tc.setup.volPartitions, tc.setup.volRowsPerPart), nil
	}
	if tc.setup.volPartitions > 0 {
		return dtGenVolumeSQL(tc.setup.volPartitions, tc.setup.volRowsPerPart), nil
	}
	if tc.fixture != "" {
		return dtReadFixture(tc.fixture)
	}
	return "", nil
}

// runCase executes one case end-to-end and returns the observed outcome plus an assertion error (for success cases
// whose per-schema/per-partition golden mismatches between PG14 and PG18).
func (w *dtWorker) runCase(ctx context.Context, tc dtCase) (dataTankMigrationOutcome, error) {
	if err := w.reset(); err != nil {
		return 0, fmt.Errorf("worker reset: %w", err)
	}

	// --- source PG14 cluster ---
	if err := dtInitCluster(ctx, dtPG14Version, w.pg14Data()); err != nil {
		return 0, err
	}
	src, err := dtStartCluster(ctx, dtPG14Version, w.pg14Data(), w.pg14Sock())
	if err != nil {
		return 0, err
	}
	defer src.stop()
	if err := src.ensureFixtureDB(ctx); err != nil {
		return 0, err
	}

	sqlText, ferr := w.fixtureSQLFor(tc)
	if ferr != nil {
		return 0, ferr
	}
	if strings.TrimSpace(dtStripComments(sqlText)) != "" {
		if aerr := src.applyFixtureSQL(ctx, sqlText); aerr != nil {
			return 0, aerr
		}
	}

	// --- target PG18 cluster ---
	if err := dtInitCluster(ctx, dtPG18Version, w.pg18Data()); err != nil {
		return 0, err
	}
	target, err := dtStartCluster(ctx, dtPG18Version, w.pg18Data(), w.pg18Sock())
	if err != nil {
		return 0, err
	}
	defer target.stop()
	if err := target.ensureFixtureDB(ctx); err != nil {
		return 0, err
	}

	migStart := time.Now()
	got, report, merr := runDataTankMigration(ctx, src, target, w.backupDir(), tc.setup)
	migElapsed := time.Since(migStart)
	if merr != nil {
		return 0, merr
	}

	// Load-test metric capture (exec-5c). Measured AFTER the migration returns so the dump dir + target data dir
	// reflect peak on-disk footprint of the run.
	if tc.loadTest {
		dumpBytes := dtDirSize(w.backupDir())
		pg18Bytes := dtDirSize(w.pg18Data())
		pg14Bytes := dtDirSize(w.pg14Data())
		dtRecordMetrics(dtLoadMetric{
			Case:               tc.name,
			Outcome:            got.String(),
			TierReached:        report.highestTierAttempted,
			MigrationSecs:      migElapsed.Seconds(),
			DumpDirBytes:       dumpBytes,
			PG18DataBytes:      pg18Bytes,
			PG14DataBytes:      pg14Bytes,
			PeakDiskBytes:      pg14Bytes + dumpBytes + pg18Bytes,
			ReservedWordRouted: report.reservedWordRouted,
			OldClusterRetained: report.oldClusterRetained,
		})
	}

	// Report-level assertions: HOW the outcome was reached. The report is populated from the engine's migrationResult,
	// so these verify the engine really routed/escalated as the case demands, not just that the final enum matched.
	if tc.wantReservedWordRouted && !report.reservedWordRouted {
		return got, fmt.Errorf("expected reserved-word routing to tier 3, but migration did not route on a reserved word (report.reservedWordRouted=false)")
	}
	if tc.wantMinTier > 0 && report.highestTierAttempted < tc.wantMinTier {
		return got, fmt.Errorf("expected escalation to reach at least tier %d, but highestTierAttempted=%d", tc.wantMinTier, report.highestTierAttempted)
	}
	if tc.wantOldClusterRetained && !report.oldClusterRetained {
		return got, fmt.Errorf("expected a preserve outcome to retain the old PG14 cluster dir, but report.oldClusterRetained=false")
	}

	// Data-preservation invariant (governing decision 2026-06-08): for EVERY failure-ending outcome the original must
	// remain recoverable on disk - the old PG14 data directory present and populated, plus (where a dump was taken) the
	// safety dump directory retained. Asserted per failure case so a regression that deletes the old dir before the
	// migration is 100% complete is caught here. (exec-6d is the dedicated gate test; this is the per-case guard.)
	if isDataTankFailureOutcome(got) {
		if perr := dtAssertOldDataDirPreserved(w.pg14Data()); perr != nil {
			return got, fmt.Errorf("data-preservation invariant violated for %s: %w", got, perr)
		}
		// On a dump failure / disk pre-flight abort / refresh-pause abort the dump directory may be absent by
		// definition (the dump never completed); the old data dir is the surviving copy. When the dump DID run and the
		// restore then failed (DataPreservedOnDisk) or only partially completed (PartialRestoreDataPreserved), the
		// safety dump directory must be retained.
		if got == dtOutcomeDataPreservedOnDisk || got == dtOutcomePartialRestoreDataPreserved {
			if derr := dtAssertDumpDirRetained(filepath.Join(w.backupDir(), "data-tank")); derr != nil {
				return got, fmt.Errorf("data-preservation invariant violated for %s: %w", got, derr)
			}
		}
	}

	// For success cases with a golden, run the assert SQL on both clusters and compare digests (per-schema,
	// per-partition topology + row counts).
	if isDataTankSuccess(got) && tc.assert != "" {
		assertSQL, aerr := dtReadAssert(tc.assert)
		if aerr != nil {
			return got, aerr
		}
		oldRows, oerr := dtQueryChecksum(ctx, src, assertSQL)
		if oerr != nil {
			return got, fmt.Errorf("assert query on PG14: %w", oerr)
		}
		newRows, nerr := dtQueryChecksum(ctx, target, assertSQL)
		if nerr != nil {
			return got, fmt.Errorf("assert query on PG18: %w", nerr)
		}
		if oldRows != newRows {
			return got, fmt.Errorf("assert golden mismatch: PG14 digest %s != PG18 digest %s", oldRows, newRows)
		}
	}

	return got, nil
}

func isDataTankSuccess(o dataTankMigrationOutcome) bool {
	switch o {
	case dtOutcomeAutoRestoreSucceededAtTier1,
		dtOutcomeAutoRestoreSucceededAtTier2,
		dtOutcomeAutoRestoreSucceededAtTier3:
		return true
	}
	return false
}

// isDataTankFailureOutcome reports whether an outcome is a terminal failure where the data-preservation invariant must
// hold (old dir preserved; dump retained where one was taken).
func isDataTankFailureOutcome(o dataTankMigrationOutcome) bool {
	switch o {
	case dtOutcomeDumpFailed,
		dtOutcomeRefreshPauseFailed,
		dtOutcomeDiskPreflightFailed,
		dtOutcomePartialRestoreDataPreserved,
		dtOutcomeDataPreservedOnDisk:
		return true
	}
	return false
}

// dtAssertOldDataDirPreserved confirms the source PG14 data directory still exists and still holds a real cluster (its
// PG_VERSION marker present) after a failed migration - the data-preservation guarantee that the original is
// recoverable.
func dtAssertOldDataDirPreserved(dataDir string) error {
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

// dtAssertDumpDirRetained confirms the directory-format safety dump is still on disk and non-empty after an
// all-tiers-failed outcome (the second independent recovery copy required by the governing decision).
func dtAssertDumpDirRetained(dumpDir string) error {
	info, err := os.Stat(dumpDir)
	if err != nil {
		return fmt.Errorf("safety dump dir %s not retained after failure: %w", dumpDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("safety dump path %s is not a directory", dumpDir)
	}
	entries, err := os.ReadDir(dumpDir)
	if err != nil {
		return fmt.Errorf("reading safety dump dir %s: %w", dumpDir, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("safety dump dir %s retained but empty", dumpDir)
	}
	return nil
}

func dtQueryChecksum(ctx context.Context, c *dtCluster, sqlText string) (string, error) {
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

func dtFixtureDir() string { return "migration_datatank_fixtures" }
func dtAssertDir() string  { return "migration_datatank_assert" }

func dtReadFixture(name string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dtFixtureDir(), name))
	if err != nil {
		return "", fmt.Errorf("read fixture %s: %w", name, err)
	}
	return string(b), nil
}

func dtReadAssert(name string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dtAssertDir(), name))
	if err != nil {
		return "", fmt.Errorf("read assert %s: %w", name, err)
	}
	return string(b), nil
}

// -----------------------------------------------------------------------------
// Load-test metric capture (exec-5c). Each LE-* case appends one JSON line to the metrics file named by
// STEAMPIPE_DT_XMIG_METRICS (default: a file under the test root). The load-test report is built from these measured
// numbers.
// -----------------------------------------------------------------------------

type dtLoadMetric struct {
	Case               string  `json:"case"`
	Outcome            string  `json:"outcome"`
	TierReached        int     `json:"tier_reached"`
	MigrationSecs      float64 `json:"migration_secs"`
	DumpDirBytes       int64   `json:"dump_dir_bytes"`
	PG18DataBytes      int64   `json:"pg18_data_bytes"`
	PG14DataBytes      int64   `json:"pg14_data_bytes"`
	PeakDiskBytes      int64   `json:"peak_disk_bytes"`
	ReservedWordRouted bool    `json:"reserved_word_routed"`
	OldClusterRetained bool    `json:"old_cluster_retained"`
}

var dtMetricsMu sync.Mutex

func dtMetricsPath() string {
	if p := os.Getenv("STEAMPIPE_DT_XMIG_METRICS"); p != "" {
		return p
	}
	return filepath.Join(dtTestRoot(), "load-metrics.jsonl")
}

// dtRecordMetrics appends one metric line. Append-mode + a mutex keeps it safe across the parallel subtests.
func dtRecordMetrics(m dtLoadMetric) {
	dtMetricsMu.Lock()
	defer dtMetricsMu.Unlock()
	line, err := json.Marshal(m)
	if err != nil {
		return
	}
	f, err := os.OpenFile(dtMetricsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(line, '\n'))
}

// dtDirSize returns the total on-disk bytes under dir (best-effort; unreadable entries are skipped). Used to measure
// dump-dir + cluster footprint.
func dtDirSize(dir string) int64 {
	var total int64
	_ = filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

func dtStripComments(sqlText string) string {
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

func TestDataTankMigration(t *testing.T) {
	if os.Getenv("STEAMPIPE_DT_XMIG_TEST") == "off" {
		t.Skip("data-tank migration matrix disabled via STEAMPIPE_DT_XMIG_TEST=off")
	}
	for _, v := range []string{dtPG14Version, dtPG18Version} {
		bin := filepath.Join(dtPGBinDir(v), "postgres")
		if _, err := os.Stat(bin); err != nil {
			t.Skipf("PG%s binary not found at %s - place binaries per the suite header (set STEAMPIPE_DT_XMIG_TEST_ROOT to override)", v, bin)
		}
	}

	cases := dtCases()

	workersBase := filepath.Join(dtTestRoot(), "workers")
	if err := os.MkdirAll(workersBase, 0755); err != nil {
		t.Fatalf("create workers base: %v", err)
	}

	// Each case runs as its own subtest so the standard `-run` filter selects individual cases (e.g. -run
	// TestDataTankMigration/DT-B5 for the timing run). A buffered semaphore bounds concurrency to dtParallelism(); a
	// per-case unique worker dir keeps clusters isolated when run in parallel.
	sem := make(chan struct{}, dtParallelism())
	ctx := context.Background()

	var passCount, failCount int32
	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sem <- struct{}{}
			defer func() { <-sem }()

			w := &dtWorker{id: i, baseDir: filepath.Join(workersBase, fmt.Sprintf("c%02d_%s", i, tc.name))}
			got, err := w.runCase(ctx, tc)
			pass := err == nil && got == tc.expected
			if pass {
				atomic.AddInt32(&passCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
			if err != nil {
				t.Errorf("expected %s: %v", tc.expected, err)
			} else if got != tc.expected {
				t.Errorf("outcome = %s, want %s", got, tc.expected)
			}
		})
	}

	t.Cleanup(func() {
		t.Logf("data-tank migration matrix: %d cases (this run: %d PASS, %d FAIL)", len(cases), atomic.LoadInt32(&passCount), atomic.LoadInt32(&failCount))
	})
}
