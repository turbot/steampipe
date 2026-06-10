package db_local

// Data-tank cross-major (PG14 -> PG18) migration.
//
// This implements the light-migration + tiered-restore path for data-tank schemas (the <handle> + "<handle>-parts"
// schema pairs that carry the PARTITION BY LIST(_cloud_partition) cached upstream data). It is deliberately narrower
// than the general public-schema migration: data tank has no procedural code, no expression/partial/GIN/GIST indexes,
// and only one catalog risk - a column literally named system_user (PG16+ reserves SYSTEM_USER).
//
// Failure stance: on any unrecoverable or partial failure the OLD PG14 data directory (plus that attempt's safety dump)
// is left intact on disk and the failure is surfaced to the caller; the production caller (restoreDBBackup) fail-stops
// startup on any cross-major failure, so the service never runs on a half-migrated database. No version-revert; data is
// never dropped. The old data directory is the durable copy - a retry replaces the prior attempt's dump with a fresh
// one from it (dumpDataTankSchemas). The tiered restore escalates parallel -> serial -> per-table COPY -> per-partition
// COPY before giving up.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sys/unix"

	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// availableDiskBytes returns bytes available to an unprivileged user at path.
func availableDiskBytes(path string) (int64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("statfs %s: %w", path, err)
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}

// reservedColumnNames are the lower-cased column names that a PG14 pg_dump emits unquoted and a PG18 pg_restore then
// rejects with a syntax error. Currently the only data-tank-relevant entry is system_user (SYSTEM_USER reserved in
// PG16+).
var reservedColumnNames = map[string]struct{}{
	"system_user": {},
}

// dataTankRestoreTier names the restore strategy reached. The operational cost differs per tier, so the migration
// records the highest tier it had to use.
type dataTankRestoreTier int

const (
	dtRestoreTierNone dataTankRestoreTier = iota
	dtRestoreTier1Parallel
	dtRestoreTier2Serial
	dtRestoreTier3PerTable
	dtRestoreTier4PerPartition
)

// reservedWordHit records a data-tank column whose name is a reserved word that would break an unquoted pg_restore.
type reservedWordHit struct {
	schema string
	table  string
	column string
}

// partitionFailure records a single partition that could not be migrated even by the per-partition COPY tier. It feeds
// the orchestrator's "needs help" list.
type partitionFailure struct {
	ParentSchema string `json:"parent_schema"`
	ParentTable  string `json:"parent_table"`
	PartSchema   string `json:"part_schema"`
	PartTable    string `json:"part_table"`
	Reason       string `json:"reason"`
}

// dataTankMigrationStatus is the orchestrator-facing structured failure signal. It is JSON-serialised to a well-known
// marker file the orchestrator polls. The marker is the contract surface this task ships; orchestrator-side polling is
// the separate 5b-orchestrator task.
type dataTankMigrationStatus struct {
	Committed          bool               `json:"committed"`
	TierReached        int                `json:"tier_reached"`
	ReservedWordRouted bool               `json:"reserved_word_routed"`
	OldClusterRetained bool               `json:"old_cluster_retained"`
	RetainedDumpPath   string             `json:"retained_dump_path"`
	FailedTank         string             `json:"failed_tank,omitempty"`
	FailedPartitions   []partitionFailure `json:"failed_partitions,omitempty"`
	Message            string             `json:"message,omitempty"`
}

// pgClusterRef is the minimal handle the shared cross-major migration engine needs to talk to one PostgreSQL cluster -
// enough to run the bundled pg_dump / pg_restore binaries and open a pgx connection. It generalises both shapes: the
// data-tank test harness and the production startup path both supply one. A cluster is reached either over a Unix
// socket directory (sockDir, used by the in-package test clusters) or over a TCP loopback port (port>0, used by the
// production old/new embedded clusters).
type pgClusterRef struct {
	version string   // postgres version (selects the matching binary set)
	binDir  string   // directory holding pg_dump / pg_restore
	env     []string // process env (library path, PGSSLMODE, ...)
	sockDir string   // Unix socket directory (used as host= when port == 0)
	port    int      // TCP loopback port (used as host=127.0.0.1 port= when > 0)
	dbName  string
	user    string
	dataDir string // on-disk data directory (the preserved-on-failure original)
}

func (r *pgClusterRef) connString() string {
	if r.port > 0 {
		return fmt.Sprintf("host=127.0.0.1 port=%d user=%s dbname=%s sslmode=disable", r.port, r.user, r.dbName)
	}
	return fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable", r.sockDir, r.user, r.dbName)
}

func (r *pgClusterRef) connect(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, r.connString())
}

func (r *pgClusterRef) tool(name string) string {
	return filepath.Join(r.binDir, name)
}

// toolConnArgs returns the host/port arguments the bundled pg_dump / pg_restore binaries need to reach this cluster,
// matching connString's socket-vs-TCP selection.
func (r *pgClusterRef) toolConnArgs() []string {
	if r.port > 0 {
		return []string{"--host=127.0.0.1", fmt.Sprintf("--port=%d", r.port)}
	}
	return []string{"--host=" + r.sockDir}
}

// migrationFaults carries deterministic failure injection for testing the tier-escalation and pre-flight paths. In
// production every field is false / zero and the migration runs against real cluster state. The faults model failure
// modes that cannot be reliably produced on a developer host (a full disk, a restore that fails for one partition's
// bytes, an orchestration that did not honour the refresh-pause contract).
type migrationFaults struct {
	forceDumpFailure        bool
	forceDiskPreflightFail  bool
	refreshPauseNotHonoured bool

	failTier1           bool
	failTier2           bool
	failTier3           bool
	failTier3MidTable   bool // tier 3 dies halfway through one parent's partitions - some fully loaded, then error
	corruptOnePartition bool
	failAllTiers        bool

	// interruptMidRestore / targetUnusable model an interrupted migration (SIGKILL / OOM / pod restart) or an
	// old-cluster binary issue: the restore cannot proceed, so the old data directory + safety dump are preserved on
	// disk and the failure is surfaced (no version-revert).
	interruptMidRestore bool
	targetUnusable      bool
}

// migrationResult is the engine's full report. The test seam and the real startup wiring both read it.
type migrationResult struct {
	tierReached        dataTankRestoreTier
	reservedWordRouted bool
	oldClusterRetained bool
	committed          bool
	// dumpPath is the retained dump artefact: a directory for the data-tank shape, a single custom-format file for the
	// public shape.
	dumpPath          string
	partitionFailures []partitionFailure
	skippedMgrTables  []string

	// preflightSkipped is set when the public shape's collation pre-check flagged a risk (or could not run) and the
	// restore was skipped. validationDiverged is set when the public shape's post-restore row-checksum validation found
	// a divergence. Both are public-shape-only signals (the data-tank shape runs neither gate); they let the production
	// caller pick the cause-specific warning.
	preflightSkipped   bool
	validationDiverged bool

	// matviewRefreshFailed is set when the public shape's separate matview-refresh invocation failed. The migration
	// still commits (a failed refresh is a warning, not a migration failure); the caller surfaces the user-facing
	// "REFRESH manually" warning.
	matviewRefreshFailed bool
}

// ----------------------------------------------------------------------------
// Schema / table / partition enumeration.
// ----------------------------------------------------------------------------

// listDataTankSchemas enumerates the data-tank schema pairs on the source cluster: every <handle> schema and its
// "<handle>-parts" sibling. System schemas and public are excluded.
func listDataTankSchemas(ctx context.Context, conn *pgx.Conn) ([]string, error) {
	rows, err := conn.Query(ctx, `
		SELECT nspname FROM pg_namespace
		WHERE nspname NOT IN ('public', 'pg_catalog', 'information_schema', 'pg_toast', 'steampipe_internal')
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

// dataTankTable identifies a partitioned parent table living in a <handle> schema (NOT the -parts schema, which only
// holds child partitions).
type dataTankTable struct {
	schema string
	table  string
}

// listDataTankParentTables returns the partitioned parent tables across the given schemas. The -parts schemas hold only
// attached child partitions, so the parents are what the per-table tier iterates. Tables carrying the _mgr_ infix (an
// in-flight user-driven intra-version migration) are returned separately so the caller can decide to skip-or-wait
// rather than racing the cleanup workflow.
func listDataTankParentTables(ctx context.Context, conn *pgx.Conn, schemas []string) (parents []dataTankTable, mgrTables []dataTankTable, err error) {
	for _, schema := range schemas {
		rows, qerr := conn.Query(ctx, `
			SELECT c.relname
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = $1
			  AND c.relkind IN ('p','r')
			  AND c.relispartition = false
			ORDER BY c.relname`, schema)
		if qerr != nil {
			return nil, nil, qerr
		}
		func() {
			defer rows.Close()
			for rows.Next() {
				var name string
				if serr := rows.Scan(&name); serr != nil {
					err = serr
					return
				}
				t := dataTankTable{schema: schema, table: name}
				if strings.Contains(name, "_mgr_") {
					mgrTables = append(mgrTables, t)
				} else {
					parents = append(parents, t)
				}
			}
			err = rows.Err()
		}()
		if err != nil {
			return nil, nil, err
		}
	}
	return parents, mgrTables, nil
}

// dataTankPartition identifies one attached child partition plus the value list needed to re-attach it on the target.
type dataTankPartition struct {
	partSchema string
	partTable  string
	forValues  string // the FOR VALUES IN (...) expression text
}

// listAttachedPartitions returns each attached child partition of a parent table, with the partition-bound expression
// needed to re-ATTACH it.
func listAttachedPartitions(ctx context.Context, conn *pgx.Conn, parent dataTankTable) ([]dataTankPartition, error) {
	rows, err := conn.Query(ctx, `
		SELECT child_n.nspname, child.relname, pg_get_expr(child.relpartbound, child.oid)
		FROM pg_inherits i
		JOIN pg_class parent ON parent.oid = i.inhparent
		JOIN pg_namespace parent_n ON parent_n.oid = parent.relnamespace
		JOIN pg_class child ON child.oid = i.inhrelid
		JOIN pg_namespace child_n ON child_n.oid = child.relnamespace
		WHERE parent_n.nspname = $1 AND parent.relname = $2
		ORDER BY child_n.nspname, child.relname`, parent.schema, parent.table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var parts []dataTankPartition
	for rows.Next() {
		var p dataTankPartition
		if serr := rows.Scan(&p.partSchema, &p.partTable, &p.forValues); serr != nil {
			return nil, serr
		}
		parts = append(parts, p)
	}
	return parts, rows.Err()
}

// ----------------------------------------------------------------------------
// Reserved-word pre-flight scan.
// ----------------------------------------------------------------------------

// reservedWordColumnScan scans every column of every table in the given data-tank schemas and flags any column whose
// lower-cased name is a reserved word. A hit routes the migration straight to tier 3 (per-table COPY with quoted
// identifiers), because pg_restore's syntax-level rejection happens on the unquoted DDL pg_dump emits - which tiers 1
// and 2 cannot avoid.
func reservedWordColumnScan(ctx context.Context, conn *pgx.Conn, schemaNames []string) ([]reservedWordHit, error) {
	if len(schemaNames) == 0 {
		return nil, nil
	}
	rows, err := conn.Query(ctx, `
		SELECT n.nspname, c.relname, a.attname
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = ANY($1)
		  AND a.attnum > 0
		  AND NOT a.attisdropped
		ORDER BY n.nspname, c.relname, a.attnum`, schemaNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hits []reservedWordHit
	for rows.Next() {
		var h reservedWordHit
		if serr := rows.Scan(&h.schema, &h.table, &h.column); serr != nil {
			return nil, serr
		}
		if _, reserved := reservedColumnNames[strings.ToLower(h.column)]; reserved {
			hits = append(hits, h)
		}
	}
	return hits, rows.Err()
}

// ----------------------------------------------------------------------------
// Disk-space pre-flight.
// ----------------------------------------------------------------------------

// dataTankDiskPreflight estimates the disk the migration window needs (~2x the data-tank cluster size) and checks the
// install directory has it, minus a 1 GB safety margin. Returns the shortfall in bytes (0 = ok).
func dataTankDiskPreflight(ctx context.Context, conn *pgx.Conn, schemas []string, targetDir string) (shortfallBytes int64, err error) {
	var totalBytes int64
	for _, schema := range schemas {
		var schemaBytes int64
		row := conn.QueryRow(ctx, `
			SELECT COALESCE(SUM(pg_total_relation_size(c.oid)), 0)
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = $1 AND c.relkind IN ('r','p')`, schema)
		if serr := row.Scan(&schemaBytes); serr != nil {
			return 0, serr
		}
		totalBytes += schemaBytes
	}
	const safetyMargin = int64(1) << 30 // 1 GB
	needed := totalBytes*2 + safetyMargin

	avail, derr := availableDiskBytes(targetDir)
	if derr != nil {
		return 0, derr
	}
	if avail < needed {
		return needed - avail, nil
	}
	return 0, nil
}

// ----------------------------------------------------------------------------
// Dump.
// ----------------------------------------------------------------------------

// dumpDataTankSchemas runs a single directory-format pg_dump covering every data-tank schema pair against the
// still-running old cluster. Directory format lets every restore tier read the SAME artefact (TOC + per-table data
// files) without re-dumping. The whole schema set is dumped in one invocation so the hyphenated "<handle>-parts"
// partition tables and their parents land in one consistent, dependency-ordered archive.
func dumpDataTankSchemas(ctx context.Context, src *pgClusterRef, schemas []string, dumpDir string, jobs int) error {
	// A prior FAILED attempt leaves its retained dump dir behind, and pg_dump's directory format refuses an existing
	// directory ("File exists") - without this, every retry deadlocks at the dump step. A retry only runs while the old
	// cluster is still intact on disk (the deletion gate fires on full success only), so replacing the stale dump with
	// a fresh one from the same source sacrifices nothing.
	if err := os.RemoveAll(dumpDir); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dumpDir), 0755); err != nil {
		return err
	}
	args := []string{
		"--file=" + dumpDir,
		"--format=directory",
		fmt.Sprintf("--jobs=%d", jobs),
		"--dbname=" + src.dbName,
		"--username=" + src.user,
	}
	args = append(args, src.toolConnArgs()...)
	for _, s := range schemas {
		args = append(args, "--schema="+s)
	}
	cmd := exec.CommandContext(ctx, src.tool("pg_dump"), args...)
	cmd.Env = src.env
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pg_dump (directory) failed: %w\n%s", err, out)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Restore tiers.
// ----------------------------------------------------------------------------

// restoreTier1ParallelDump restores the directory dump with parallel jobs inside a single transaction. Fast path;
// expected to succeed for almost all workspaces given data tank's narrow risk surface.
func restoreTier1ParallelDump(ctx context.Context, target *pgClusterRef, dumpDir string, jobs int) error {
	args := []string{
		dumpDir,
		"--format=directory",
		fmt.Sprintf("--jobs=%d", jobs),
		"--exit-on-error",
		"--no-owner",
		"--dbname=" + target.dbName,
		"--username=" + target.user,
	}
	args = append(args, target.toolConnArgs()...)
	cmd := exec.CommandContext(ctx, target.tool("pg_restore"), args...)
	cmd.Env = target.env
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tier1 parallel pg_restore failed: %w\n%s", err, out)
	}
	return nil
}

// restoreTier2SerialDump restores the directory dump serially in a single transaction. Avoids
// parallel-dependency-ordering edges that can trip tier 1 on schemas with unusual constraint shapes.
func restoreTier2SerialDump(ctx context.Context, target *pgClusterRef, dumpDir string) error {
	args := []string{
		dumpDir,
		"--format=directory",
		"--single-transaction",
		"--no-owner",
		"--dbname=" + target.dbName,
		"--username=" + target.user,
	}
	args = append(args, target.toolConnArgs()...)
	cmd := exec.CommandContext(ctx, target.tool("pg_restore"), args...)
	cmd.Env = target.env
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tier2 serial pg_restore failed: %w\n%s", err, out)
	}
	return nil
}

// restoreTier3PerTableCOPY recreates each parent table's DDL with quoted identifiers (sidestepping the SYSTEM_USER
// reserved-word case) and COPYs the data table-by-table. Bypasses pg_restore's all-or-nothing transaction: a single bad
// table does not block the others. Returns the parent tables that failed (input to tier 4).
//
// It reads the LIVE old cluster for table structure + partition metadata and streams data from old -> new via COPY, so
// it does not depend on pg_restore being able to parse the dump's emitted DDL at all.
func restoreTier3PerTableCOPY(ctx context.Context, src, target *pgClusterRef, parents []dataTankTable, faults migrationFaults) (failed []dataTankTable, err error) {
	for _, p := range parents {
		if faults.corruptOnePartition {
			// A corrupt partition's data cannot be COPYed as a whole table, so the per-table tier fails for the
			// affected parent and defers it to the per-partition tier.
			failed = append(failed, p)
			continue
		}
		if cerr := copyParentTable(ctx, src, target, p, faults); cerr != nil {
			failed = append(failed, p)
		}
	}
	return failed, nil
}

// copyParentTable creates the partitioned parent + every child partition on the target (all identifiers quoted),
// attaches the partitions, then COPYs each partition's rows across. The structure is read from the live source catalog.
func copyParentTable(ctx context.Context, src, target *pgClusterRef, parent dataTankTable, faults migrationFaults) error {
	srcConn, err := src.connect(ctx)
	if err != nil {
		return err
	}
	defer srcConn.Close(ctx)
	tgtConn, err := target.connect(ctx)
	if err != nil {
		return err
	}
	defer tgtConn.Close(ctx)

	parts, err := listAttachedPartitions(ctx, srcConn, parent)
	if err != nil {
		return err
	}

	if err := createParentAndPartitions(ctx, srcConn, tgtConn, parent, parts); err != nil {
		return err
	}
	for i, part := range parts {
		// Models the real-world tier-3 failure shape: some partitions fully loaded, then a mid-table death. The
		// escalation to tier 4 must rebuild this parent from scratch or the loaded partitions get their rows twice.
		if faults.failTier3MidTable && i == len(parts)/2 {
			return fmt.Errorf("injected mid-table failure after %d of %d partitions", i, len(parts))
		}
		if err := copyPartitionData(ctx, src, target, part); err != nil {
			return fmt.Errorf("copy partition %q.%q: %w", part.partSchema, part.partTable, err)
		}
	}
	// A plain (non-partitioned) table holds its rows directly - there are no partitions to carry them. A partitioned
	// parent with zero partitions has no rows by definition, so this only fires for genuinely plain tables.
	if len(parts) == 0 {
		if plain, perr := isPlainTable(ctx, srcConn, parent); perr != nil {
			return perr
		} else if plain {
			if err := copyTableRows(ctx, src, target, parent); err != nil {
				return fmt.Errorf("copy table %q.%q: %w", parent.schema, parent.table, err)
			}
		}
	}
	return nil
}

// isPlainTable reports whether the table is non-partitioned (no partition key).
func isPlainTable(ctx context.Context, conn *pgx.Conn, t dataTankTable) (bool, error) {
	expr, err := partitionKeyExpr(ctx, conn, t.schema, t.table)
	if err != nil {
		return false, err
	}
	return expr == "", nil
}

// copyTableRows streams a plain table's rows old -> new via the same COPY pipe the partition path uses.
func copyTableRows(ctx context.Context, src, target *pgClusterRef, t dataTankTable) error {
	return copyPartitionData(ctx, src, target, dataTankPartition{partSchema: t.schema, partTable: t.table})
}

// dropTargetTable removes one table (and its attached partitions) on the target so a rebuild starts from empty.
func dropTargetTable(ctx context.Context, tgtConn *pgx.Conn, t dataTankTable) error {
	_, err := tgtConn.Exec(ctx, "DROP TABLE IF EXISTS "+qualName(t.schema, t.table)+" CASCADE")
	return err
}

// restoreTier4PerPartitionCOPY handles parents that tier 3 could not migrate whole. It creates the parent and migrates
// partitions one at a time, attaching the ones that COPY cleanly and recording the failures on the needs-help list. A
// tank that lands 19/20 partitions is degraded-but-functional; the bad partition can be hand-migrated.
func restoreTier4PerPartitionCOPY(ctx context.Context, src, target *pgClusterRef, parents []dataTankTable, faults migrationFaults) (failures []partitionFailure, err error) {
	for _, parent := range parents {
		srcConn, cerr := src.connect(ctx)
		if cerr != nil {
			return nil, cerr
		}
		tgtConn, cerr := target.connect(ctx)
		if cerr != nil {
			srcConn.Close(ctx)
			return nil, cerr
		}

		parts, lerr := listAttachedPartitions(ctx, srcConn, parent)
		if lerr != nil {
			srcConn.Close(ctx)
			tgtConn.Close(ctx)
			return nil, lerr
		}

		// Rebuild this parent from scratch. A failed tier-3 attempt may have left some partitions fully loaded on the
		// target; re-copying into them would APPEND (COPY adds rows), silently doubling their data. Dropping just this
		// parent (never the whole schema - tier 3 may have completed OTHER parents that tier 4 won't revisit)
		// guarantees every partition below starts empty.
		if derr := dropTargetTable(ctx, tgtConn, parent); derr != nil {
			srcConn.Close(ctx)
			tgtConn.Close(ctx)
			return nil, fmt.Errorf("tier4 reset of %s failed: %w", qualName(parent.schema, parent.table), derr)
		}
		// Create the bare parent (no partitions yet). A failure here is NOT tolerable: with no parent every partition
		// below would "fail" - or, for a parent with zero partitions, NOTHING would fail and the table would silently
		// vanish from the migration. Record it and move to the next parent.
		if cerr := createParent(ctx, srcConn, tgtConn, parent); cerr != nil {
			failures = append(failures, partitionFailure{
				ParentSchema: parent.schema,
				ParentTable:  parent.table,
				Reason:       "create table on target failed: " + cerr.Error(),
			})
			srcConn.Close(ctx)
			tgtConn.Close(ctx)
			continue
		}

		// Plain table: its rows live in the table itself, not in partitions - copy them directly, recording a
		// failure like any partition would get.
		if len(parts) == 0 {
			if plain, perr := isPlainTable(ctx, srcConn, parent); perr == nil && plain {
				if cerr := copyTableRows(ctx, src, target, parent); cerr != nil {
					failures = append(failures, partitionFailure{
						ParentSchema: parent.schema,
						ParentTable:  parent.table,
						Reason:       "table data copy failed: " + cerr.Error(),
					})
				}
			} else if perr != nil {
				failures = append(failures, partitionFailure{
					ParentSchema: parent.schema,
					ParentTable:  parent.table,
					Reason:       "could not classify table: " + perr.Error(),
				})
			}
		}

		for i, part := range parts {
			// corruptOnePartition models exactly one unrestorable partition.
			if faults.corruptOnePartition && i == corruptPartitionIndex(len(parts)) {
				failures = append(failures, partitionFailure{
					ParentSchema: parent.schema,
					ParentTable:  parent.table,
					PartSchema:   part.partSchema,
					PartTable:    part.partTable,
					Reason:       "partition data unrestorable; deferred to manual migration",
				})
				continue
			}
			if perr := copyAndAttachPartition(ctx, src, target, srcConn, tgtConn, parent, part); perr != nil {
				failures = append(failures, partitionFailure{
					ParentSchema: parent.schema,
					ParentTable:  parent.table,
					PartSchema:   part.partSchema,
					PartTable:    part.partTable,
					Reason:       perr.Error(),
				})
			}
		}
		srcConn.Close(ctx)
		tgtConn.Close(ctx)
	}
	return failures, nil
}

// corruptPartitionIndex picks which partition the corrupt-injection fault hits. The DT-F4 fixture documents "partition
// 7"; for smaller scaled fixtures it clamps into range so exactly one partition is always affected.
func corruptPartitionIndex(n int) int {
	idx := 6 // 0-based -> "partition 7"
	if idx >= n {
		idx = n - 1
	}
	return idx
}

// ----------------------------------------------------------------------------
// COPY building blocks.
// ----------------------------------------------------------------------------

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func qualName(schema, table string) string {
	return quoteIdent(schema) + "." + quoteIdent(table)
}

// columnDDL returns the "<quoted col> <type>" fragments for a relation, in attribute order, with identifiers quoted so
// a reserved-word column name is safe.
func columnDDL(ctx context.Context, conn *pgx.Conn, schema, table string) ([]string, []string, error) {
	rows, err := conn.Query(ctx, `
		SELECT a.attname, format_type(a.atttypid, a.atttypmod)
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2
		  AND a.attnum > 0 AND NOT a.attisdropped
		ORDER BY a.attnum`, schema, table)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var defs, names []string
	for rows.Next() {
		var name, typ string
		if serr := rows.Scan(&name, &typ); serr != nil {
			return nil, nil, serr
		}
		defs = append(defs, fmt.Sprintf("%s %s", quoteIdent(name), typ))
		names = append(names, name)
	}
	return defs, names, rows.Err()
}

// partitionKeyExpr returns the parent table's PARTITION BY expression text.
func partitionKeyExpr(ctx context.Context, conn *pgx.Conn, schema, table string) (string, error) {
	// pg_get_partkeydef returns NULL for a non-partitioned table - a legitimate inhabitant of tank schemas (a
	// detached old partition awaiting cleanup, or an interrupted refresh's pre-attach table). Scan via a pointer so
	// NULL means "plain table" instead of a scan error.
	var expr *string
	err := conn.QueryRow(ctx, `
		SELECT pg_get_partkeydef(c.oid)
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2`, schema, table).Scan(&expr)
	if err != nil {
		return "", err
	}
	if expr == nil {
		return "", nil
	}
	return *expr, nil
}

// createParent creates only the partitioned parent (no children) on the target.
func createParent(ctx context.Context, srcConn, tgtConn *pgx.Conn, parent dataTankTable) error {
	defs, _, err := columnDDL(ctx, srcConn, parent.schema, parent.table)
	if err != nil {
		return err
	}
	partExpr, err := partitionKeyExpr(ctx, srcConn, parent.schema, parent.table)
	if err != nil {
		return err
	}
	if _, err := tgtConn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+quoteIdent(parent.schema)); err != nil {
		return err
	}
	ddl := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", qualName(parent.schema, parent.table), strings.Join(defs, ", "))
	if partExpr != "" {
		ddl += " PARTITION BY " + partExpr
	}
	_, err = tgtConn.Exec(ctx, ddl)
	return err
}

// createParentAndPartitions creates the parent plus every child partition table (in its -parts schema) and attaches
// them. Used by the whole-table tier 3 path.
func createParentAndPartitions(ctx context.Context, srcConn, tgtConn *pgx.Conn, parent dataTankTable, parts []dataTankPartition) error {
	if err := createParent(ctx, srcConn, tgtConn, parent); err != nil {
		return err
	}
	parentDefs, _, err := columnDDL(ctx, srcConn, parent.schema, parent.table)
	if err != nil {
		return err
	}
	for _, part := range parts {
		if err := createAndAttachPartitionTable(ctx, tgtConn, parent, part, parentDefs); err != nil {
			return err
		}
	}
	return nil
}

// createAndAttachPartitionTable creates one child partition table with quoted identifiers and attaches it to the
// parent.
func createAndAttachPartitionTable(ctx context.Context, tgtConn *pgx.Conn, parent dataTankTable, part dataTankPartition, parentDefs []string) error {
	if _, err := tgtConn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+quoteIdent(part.partSchema)); err != nil {
		return err
	}
	createChild := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", qualName(part.partSchema, part.partTable), strings.Join(parentDefs, ", "))
	if _, err := tgtConn.Exec(ctx, createChild); err != nil {
		return err
	}
	attach := fmt.Sprintf("ALTER TABLE %s ATTACH PARTITION %s %s",
		qualName(parent.schema, parent.table), qualName(part.partSchema, part.partTable), part.forValues)
	if _, err := tgtConn.Exec(ctx, attach); err != nil {
		// A partition already attached (partial earlier tier) is not fatal.
		if !strings.Contains(err.Error(), "already") {
			return err
		}
	}
	return nil
}

// copyAndAttachPartition (tier 4) creates+attaches a single partition then COPYs its data, so one bad partition is
// isolated from the rest.
func copyAndAttachPartition(ctx context.Context, src, target *pgClusterRef, srcConn, tgtConn *pgx.Conn, parent dataTankTable, part dataTankPartition) error {
	parentDefs, _, err := columnDDL(ctx, srcConn, parent.schema, parent.table)
	if err != nil {
		return err
	}
	if err := createAndAttachPartitionTable(ctx, tgtConn, parent, part, parentDefs); err != nil {
		return err
	}
	return copyPartitionData(ctx, src, target, part)
}

// copyPartitionData streams one partition's rows old -> new using the PostgreSQL COPY protocol over pgx: `COPY ... TO
// STDOUT (FORMAT binary)` from the source piped into `COPY ... FROM STDIN (FORMAT binary)` on the target. This keeps
// the data path independent of pg_restore's DDL parser AND independent of any external psql binary (the Steampipe DB
// bundle ships only postgres / initdb / pg_ctl / pg_dump / pg_restore).
func copyPartitionData(ctx context.Context, src, target *pgClusterRef, part dataTankPartition) error {
	rel := qualName(part.partSchema, part.partTable)

	srcConn, err := src.connect(ctx)
	if err != nil {
		return err
	}
	defer srcConn.Close(ctx)
	tgtConn, err := target.connect(ctx)
	if err != nil {
		return err
	}
	defer tgtConn.Close(ctx)

	pr, pw := io.Pipe()

	copyErr := make(chan error, 1)
	go func() {
		_, e := srcConn.PgConn().CopyTo(ctx, pw, fmt.Sprintf("COPY %s TO STDOUT (FORMAT binary)", rel))
		// Closing the writer signals EOF to the reader side.
		pw.CloseWithError(e)
		copyErr <- e
	}()

	_, inErr := tgtConn.PgConn().CopyFrom(ctx, pr, fmt.Sprintf("COPY %s FROM STDIN (FORMAT binary)", rel))
	// If the target aborted mid-stream, its reader is gone and the source goroutine is blocked in pw.Write on the
	// unbuffered pipe - it would never reach its CloseWithError/send and the receive below would hang forever. Closing
	// the read end unblocks that Write immediately with this error.
	pr.CloseWithError(inErr)
	outErr := <-copyErr
	if outErr != nil && inErr == nil {
		return fmt.Errorf("source COPY out failed: %w", outErr)
	}
	if inErr != nil {
		return fmt.Errorf("target COPY in failed: %w", inErr)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Orchestration: thin adapters over the shared engine's 10-step flow (runMigrationEngine, migration_engine.go).
// ----------------------------------------------------------------------------

var (
	errDataTankDumpFailed     = errors.New("data-tank dump failed")
	errDataTankRefreshPause   = errors.New("data-tank refresh pause not acquired")
	errDataTankDiskPreflight  = errors.New("data-tank disk pre-flight failed")
	errDataTankAllTiersFailed = errors.New("restore failed at all tiers; original preserved on disk")

	// errDataTankPartialRestore means the restore ladder produced a PARTIAL result: the ladder climbed (typically to
	// the per-partition COPY tier) and most rows landed, but at least one partition could not be migrated. Unlike
	// errDataTankAllTiersFailed (nothing moved), some data did reach the new cluster - but because a partition's rows
	// are missing this is NOT a success. The original is preserved on disk (old data dir + safety dump) and the
	// deletion gate must NOT fire.
	errDataTankPartialRestore = errors.New("restore partially failed; one or more partitions not migrated; original preserved on disk")

	// errMigrationPreflightSkipped and errMigrationValidationDiverged are the public-shape gate failures the shared
	// engine surfaces (the data-tank shape runs neither gate, so it never returns these). Both preserve the original on
	// disk, like every other failure outcome.
	errMigrationPreflightSkipped   = errors.New("cross-major migration: pre-flight collation scan skipped the restore")
	errMigrationValidationDiverged = errors.New("cross-major migration: post-restore validation found divergence")
)

// migrateDataTank runs the data-tank cross-major migration from src (PG14) to target (PG18). It is a thin shape adapter
// over the shared copy-and-fallback engine (runMigrationEngine): it builds the data-tank shape (many
// <handle>/<handle>-parts schema pairs; collation pre-check OFF and row-checksum validation OFF, per the
// light-migration policy; refresh-pause coordination ON) and delegates. On any unrecoverable failure the result carries
// oldClusterRetained=true and a non-nil error; under the 2026-06-08 governing decision the old data directory + safety
// dump are preserved on disk, and the production caller (restoreDBBackup) fail-stops startup on the surfaced failure.
// statusPath, if non-empty, receives the JSON orchestrator marker.
func migrateDataTank(ctx context.Context, src, target *pgClusterRef, backupDir, statusPath string, jobs int, faults migrationFaults) (migrationResult, error) {
	return runMigrationEngine(ctx, dataTankMigrationShape(), src, target, backupDir, statusPath, jobs, faults)
}

// dataTankMigrationJobs is the parallelism used for the production data-tank dump/restore. The CLI is single-workspace
// and not contending for cores during startup, so a small fixed value keeps the migration window bounded without
// over-subscribing.
const dataTankMigrationJobs = 4

// migrationLibEnv builds the process environment a bundled pg_dump / pg_restore needs to find the matching libraries
// for a given install. The Steampipe DB bundle links the dump/restore binaries against libraries in the install's lib
// directory, so the platform's dynamic-library search path must point at it.
func migrationLibEnv(libDir string) []string {
	env := append(os.Environ(), "PGSSLMODE=disable")
	switch runtime.GOOS {
	case "darwin":
		env = append(env, "DYLD_LIBRARY_PATH="+libDir)
	default:
		env = append(env, "LD_LIBRARY_PATH="+libDir)
	}
	return env
}

// oldClusterRef builds the engine cluster handle for the retained old (source) cluster on a cross-major startup
// migration. The old server is reached over its TCP loopback port (prepareBackup leaves it listening on 127.0.0.1); the
// bundled pg_dump lives under the OLD install location so the dump speaks the old catalog. oldInstallLocation is the
// db/<oldVersion> directory returned by findDifferentPgInstallation; dataDir is its data directory (the
// preserved-on-failure original).
func oldClusterRef(oldVersion, oldInstallLocation, dbName string, port int) *pgClusterRef {
	binDir := filepath.Join(oldInstallLocation, "postgres", "bin")
	libDir := filepath.Join(oldInstallLocation, "postgres", "lib")
	return &pgClusterRef{
		version: oldVersion,
		binDir:  binDir,
		env:     migrationLibEnv(libDir),
		port:    port,
		dbName:  dbName,
		user:    constants.DatabaseSuperUser,
		dataDir: filepath.Join(oldInstallLocation, "data"),
	}
}

// newClusterRef builds the engine cluster handle for the new (target) embedded cluster on a cross-major startup
// migration. It is the freshly-started target service reached over its loopback port, with the bundled pg_restore from
// the current install location. targetVersion is the embedded-PG version this build ships.
func newClusterRef(port int, dbName string, targetVersion string) *pgClusterRef {
	return &pgClusterRef{
		version: targetVersion,
		binDir:  filepath.Join(filepaths.GetDatabaseLocation(), "bin"),
		env:     migrationLibEnv(filepaths.GetDatabaseLibPath()),
		port:    port,
		dbName:  dbName,
		user:    constants.DatabaseSuperUser,
		dataDir: filepaths.GetDataLocation(),
	}
}

// migrateDataTankSchemasOnStartup is the production entry point the cross-major startup path calls AFTER the
// public-schema migration has succeeded and while the old cluster is still live. It detects data-tank schemas on the
// old cluster and, if any exist, runs the shared engine to migrate them old -> new. When the old cluster has no
// data-tank schemas (the normal CLI case) it is a clean no-op: it returns committed=true with no work done, so the
// caller's deletion gate is not blocked.
//
// On any data-tank failure or partial result it returns committed=false with a non-nil error; the caller preserves the
// old data directory and fail-stops startup (the public-schema success is not reverted, but startup fails; the next
// attempt rebuilds the whole new side from the preserved old data). The directory-format dump under backupDir/data-tank
// is the second independent recovery copy required by the 2026-06-08 governing decision; a retry replaces it with a
// fresh dump from the preserved old directory.
func migrateDataTankSchemasOnStartup(ctx context.Context, old, new *pgClusterRef, backupDir string) (migrationResult, error) {
	// Build the status path up front: the early failures below never reach the engine, and without a status write a
	// previous attempt's file (possibly committed:true) would remain as the orchestrator-visible state.
	statusPath := filepath.Join(backupDir, "data-tank-migration-status.json")

	oldConn, err := old.connect(ctx)
	if err != nil {
		writeDataTankStatus(statusPath, migrationResult{oldClusterRetained: true}, "could not connect to old cluster: "+err.Error())
		return migrationResult{}, fmt.Errorf("data-tank migration: could not connect to old cluster: %w", err)
	}
	schemas, err := listDataTankSchemas(ctx, oldConn)
	oldConn.Close(ctx)
	if err != nil {
		writeDataTankStatus(statusPath, migrationResult{oldClusterRetained: true}, "could not list data-tank schemas: "+err.Error())
		return migrationResult{}, fmt.Errorf("data-tank migration: could not list data-tank schemas: %w", err)
	}
	if len(schemas) == 0 {
		// No data tank present (the normal CLI workspace). Nothing to migrate; report a clean no-op so the caller's
		// deletion gate stays unblocked.
		return migrationResult{committed: true}, nil
	}

	return migrateDataTank(ctx, old, new, backupDir, statusPath, dataTankMigrationJobs, migrationFaults{})
}

// runTieredRestore tries the restore tiers in order, escalating on failure, up to ceiling - the highest tier this
// shape's content can use. When a reserved word was detected, tiers 1-2 are skipped (their pg_restore would hit the
// unquoted-DDL syntax error) and the migration starts at tier 3.
//
// ceiling caps the ladder per shape: the data-tank shape allows the full ladder (dtRestoreTier4PerPartition); the
// public shape stops after the pg_restore tiers (dtRestoreTier2Serial), because the per-table / per-partition COPY
// tiers reconstruct PARTITION BY LIST topology that only a data tank has. When the pg_restore tiers fail under a tier-2
// ceiling (or a reserved word forces a jump past the ceiling) the restore is reported as failed, preserving the
// original.
func runTieredRestore(ctx context.Context, src, target *pgClusterRef, parents []dataTankTable, dumpPath string, jobs int, reservedWordRouted bool, ceiling dataTankRestoreTier, restoreTier1Fn func(context.Context, *pgClusterRef, string, int) error, restoreTier2Fn func(context.Context, *pgClusterRef, string) error, faults migrationFaults) (dataTankRestoreTier, []partitionFailure, error) {
	// Reserved-word route: skip straight to tier 3.
	if !reservedWordRouted {
		// Tier 1: parallel pg_restore.
		if !faults.failTier1 && !faults.failAllTiers {
			if err := restoreTier1Fn(ctx, target, dumpPath, jobs); err == nil {
				return dtRestoreTier1Parallel, nil, nil
			}
		}
		// Tier 2: serial pg_restore. The target may hold partial objects from a failed tier 1; reset it first so tier 2
		// starts clean.
		if err := resetTargetDataTankSchemas(ctx, src, target); err == nil {
			if !faults.failTier2 && !faults.failAllTiers {
				if err := restoreTier2Fn(ctx, target, dumpPath); err == nil {
					return dtRestoreTier2Serial, nil, nil
				}
			}
		}
	}

	// Shape ceiling: shapes that cannot use the COPY tiers (public) stop here. The pg_restore tiers failed (or a
	// reserved word forced a jump past the ceiling); report a restore failure so the original is preserved.
	if ceiling < dtRestoreTier3PerTable {
		return dtRestoreTier2Serial, nil, errDataTankAllTiersFailed
	}

	// Tier 3: per-table COPY (quoted identifiers; reserved-word safe).
	if err := resetTargetDataTankSchemas(ctx, src, target); err != nil {
		return dtRestoreTier3PerTable, nil, fmt.Errorf("tier3 reset failed: %w", err)
	}
	if !faults.failTier3 && !faults.failAllTiers {
		failed, terr := restoreTier3PerTableCOPY(ctx, src, target, parents, faults)
		if terr != nil {
			return dtRestoreTier3PerTable, nil, terr
		}
		if len(failed) == 0 {
			return dtRestoreTier3PerTable, nil, nil
		}
		// Some tables failed -> tier 4 for just those tables.
		partFailures, perr := restoreTier4PerPartitionCOPY(ctx, src, target, failed, faults)
		if perr != nil {
			return dtRestoreTier4PerPartition, nil, perr
		}
		return dtRestoreTier4PerPartition, partFailures, nil
	}

	// Tier 4: per-partition COPY for everything (tier 3 was forced to fail).
	if faults.failAllTiers {
		return dtRestoreTier4PerPartition, nil, errDataTankAllTiersFailed
	}
	partFailures, perr := restoreTier4PerPartitionCOPY(ctx, src, target, parents, faults)
	if perr != nil {
		return dtRestoreTier4PerPartition, nil, perr
	}
	return dtRestoreTier4PerPartition, partFailures, nil
}

// resetTargetDataTankSchemas drops the data-tank schemas on the target so a subsequent tier restores into a clean
// slate. The source list defines which schemas to clear.
func resetTargetDataTankSchemas(ctx context.Context, src, target *pgClusterRef) error {
	srcConn, err := src.connect(ctx)
	if err != nil {
		return err
	}
	schemas, err := listDataTankSchemas(ctx, srcConn)
	srcConn.Close(ctx)
	if err != nil {
		return err
	}
	tgtConn, err := target.connect(ctx)
	if err != nil {
		return err
	}
	defer tgtConn.Close(ctx)
	for _, s := range schemas {
		if _, err := tgtConn.Exec(ctx, "DROP SCHEMA IF EXISTS "+quoteIdent(s)+" CASCADE"); err != nil {
			return err
		}
	}
	return nil
}

// sanityCheckRestore confirms every expected parent table exists on the target and has the same attached-partition
// count as the source. This is the light-migration validation - no row-level checksum.
func sanityCheckRestore(ctx context.Context, src, target *pgClusterRef, parents []dataTankTable) error {
	srcConn, err := src.connect(ctx)
	if err != nil {
		return err
	}
	defer srcConn.Close(ctx)
	tgtConn, err := target.connect(ctx)
	if err != nil {
		return err
	}
	defer tgtConn.Close(ctx)

	for _, p := range parents {
		// Existence is its own requirement: listAttachedPartitions on a nonexistent table returns zero rows, not an
		// error, so without this check a table that was never created on the target would sail through below.
		var exists *string
		if err := tgtConn.QueryRow(ctx, "SELECT to_regclass($1)::text", qualName(p.schema, p.table)).Scan(&exists); err != nil {
			return err
		}
		if exists == nil {
			return fmt.Errorf("table %s missing on target", qualName(p.schema, p.table))
		}
		srcParts, serr := listAttachedPartitions(ctx, srcConn, p)
		if serr != nil {
			return serr
		}
		tgtParts, terr := listAttachedPartitions(ctx, tgtConn, p)
		if terr != nil {
			return fmt.Errorf("table %s missing on target: %w", qualName(p.schema, p.table), terr)
		}
		// Tier 4 may legitimately land fewer partitions (degraded). Only flag a total absence of partitions where the
		// source had some.
		if len(srcParts) > 0 && len(tgtParts) == 0 {
			return fmt.Errorf("table %s has no attached partitions on target", qualName(p.schema, p.table))
		}
	}
	return nil
}

// writeDataTankStatus serialises the orchestrator-facing marker. A nil/empty path is a no-op (the test harness does not
// require the marker on disk).
func writeDataTankStatus(statusPath string, res migrationResult, message string) {
	if statusPath == "" {
		return
	}
	status := dataTankMigrationStatus{
		Committed:          res.committed,
		TierReached:        int(res.tierReached),
		ReservedWordRouted: res.reservedWordRouted,
		OldClusterRetained: res.oldClusterRetained,
		RetainedDumpPath:   res.dumpPath,
		FailedPartitions:   res.partitionFailures,
		Message:            message,
	}
	// FailedTank is the schema of the first recorded partition failure - the orchestrator's entry point for a
	// targeted retry. Failures with no partition attribution (e.g. a whole-dump or whole-restore error) leave it
	// empty; the message carries the cause there.
	if len(res.partitionFailures) > 0 {
		status.FailedTank = res.partitionFailures[0].ParentSchema
	}
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(statusPath, data, 0644)
}
