package db_local

// Shared engine, public shape.
//
// migrateDataTank exercises runMigrationEngine with the DATA-TANK shape
// (collation pre-check OFF, row-checksum validation OFF) across the whole
// TestDataTankMigration matrix. This suite exercises the SAME shared engine with
// the PUBLIC shape (collation pre-check ON, row-checksum validation ON, COPY-tier
// ceiling at the pg_restore tiers), proving the two shapes really do run on one
// engine and that the collation pre-check / row-checksum validation are genuine
// parameters that fire for public and are absent for data tank.
//
// It reuses the TestDataTankMigration harness (dtInitCluster / dtStartCluster /
// dtClusterRef) and the same pre-placed PG14 + PG18 binaries.

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// runPublicShapeEngine boots a PG14 source + PG18 target, applies publicSQL to
// the source, and runs the shared engine with the public shape. It returns the
// engine result, the source data dir (for the data-preservation assertion), and
// any error. verify, if non-nil, runs against the target cluster after the
// engine finishes and before the clusters are torn down.
func runPublicShapeEngine(t *testing.T, publicSQL string, verify func(ctx context.Context, target *pgClusterRef) error) (migrationResult, string, error) {
	t.Helper()
	ctx := context.Background()

	// Keep the base path short: the Unix socket path (base/s14/.s.PGSQL.NNNN) must
	// stay under the ~104-char sockaddr limit, so the full test name cannot be in
	// it. A short stable per-test directory under the test root suffices.
	base := filepath.Join(dtTestRoot(), "ep", shortTestKey(t.Name()))
	if err := os.RemoveAll(base); err != nil {
		t.Fatalf("clean base: %v", err)
	}
	pg14Data := filepath.Join(base, "pg14-data")
	pg18Data := filepath.Join(base, "pg18-data")
	pg14Sock := filepath.Join(base, "s14")
	pg18Sock := filepath.Join(base, "s18")
	backupDir := filepath.Join(base, "backups")
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
	defer src.stop()
	if err := src.ensureFixtureDB(ctx); err != nil {
		t.Fatalf("ensure pg14 db: %v", err)
	}
	if publicSQL != "" {
		if err := src.applyFixtureSQL(ctx, publicSQL); err != nil {
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
	defer target.stop()
	if err := target.ensureFixtureDB(ctx); err != nil {
		t.Fatalf("ensure pg18 db: %v", err)
	}

	statusPath := filepath.Join(backupDir, "public-migration-status.json")
	res, merr := runMigrationEngine(ctx, publicMigrationShape(), dtClusterRef(src), dtClusterRef(target), backupDir, statusPath, dtParallelism(), migrationFaults{})
	if verify != nil {
		if verr := verify(ctx, dtClusterRef(target)); verr != nil {
			t.Errorf("post-migration verification on the target cluster failed: %v", verr)
		}
	}
	return res, pg14Data, merr
}

// shortTestKey maps a test name to a short, stable, filesystem-safe directory
// key so the Unix socket path stays under the sockaddr length limit.
func shortTestKey(name string) string {
	sum := md5.Sum([]byte(name))
	return fmt.Sprintf("%x", sum[:4])
}

func skipIfNoBinaries(t *testing.T) {
	t.Helper()
	if os.Getenv("STEAMPIPE_DT_XMIG_TEST") == "off" {
		t.Skip("engine public-shape suite disabled via STEAMPIPE_DT_XMIG_TEST=off")
	}
	for _, v := range []string{dtPG14Version, dtPG18Version} {
		bin := filepath.Join(dtPGBinDir(v), "postgres")
		if _, err := os.Stat(bin); err != nil {
			t.Skipf("PG%s binary not found at %s", v, bin)
		}
	}
}

// TestSharedEnginePublicShape_CleanASCII: a clean ASCII public schema migrates to
// a committed restore through the shared engine's public shape. The collation
// pre-check (ON for public) runs and finds nothing; the row-checksum validation
// (ON for public) runs and matches.
func TestSharedEnginePublicShape_CleanASCII(t *testing.T) {
	skipIfNoBinaries(t)

	const publicSQL = `
create table public.things (id int primary key, name text);
insert into public.things values (1, 'alpha'), (2, 'bravo'), (3, 'charlie');
create index things_name_idx on public.things (name);`

	res, _, err := runPublicShapeEngine(t, publicSQL, nil)
	if err != nil {
		t.Fatalf("public-shape clean migration returned error: %v", err)
	}
	if !res.committed {
		t.Errorf("expected committed=true for a clean ASCII public schema, got result %+v", res)
	}
	if res.preflightSkipped {
		t.Errorf("collation pre-check should NOT flag a clean ASCII schema, but preflightSkipped=true")
	}
	if res.validationDiverged {
		t.Errorf("row-checksum validation should match on a clean restore, but validationDiverged=true")
	}
	if res.matviewRefreshFailed {
		t.Errorf("the matview-refresh invocation must succeed cleanly on a schema with no matviews (empty refresh list), but matviewRefreshFailed=true")
	}
}

// TestSharedEnginePublicShape_CollationPreCheckFires: the public shape's
// collation pre-check (a PARAMETER the data-tank shape does NOT run) must flag a
// btree text index over genuinely non-ASCII data and skip the restore, leaving
// the original preserved on disk. This is the public-vs-data-tank parameter made
// concrete on the shared engine.
func TestSharedEnginePublicShape_CollationPreCheckFires(t *testing.T) {
	skipIfNoBinaries(t)

	// Non-ASCII text under a btree index: the collation provider changes across a
	// major upgrade, so the pre-flight scan must flag this and skip the restore.
	const publicSQL = `
create table public.names (id int primary key, label text);
insert into public.names values (1, 'Zürich'), (2, 'Köln'), (3, 'Genève');
create index names_label_idx on public.names (label);`

	res, pg14Data, err := runPublicShapeEngine(t, publicSQL, nil)
	if !errors.Is(err, errMigrationPreflightSkipped) {
		t.Fatalf("expected errMigrationPreflightSkipped from the public-shape collation pre-check, got err=%v result=%+v", err, res)
	}
	if !res.preflightSkipped {
		t.Errorf("expected preflightSkipped=true when the collation pre-check fires")
	}
	if res.committed {
		t.Errorf("a pre-flight-skipped migration must NOT be committed")
	}
	if !res.oldClusterRetained {
		t.Errorf("a pre-flight-skipped migration must preserve the original (oldClusterRetained=true)")
	}
	// Data-preservation invariant: the source data dir survives untouched.
	if err := dtAssertOldDataDirPreserved(pg14Data); err != nil {
		t.Errorf("data-preservation invariant violated after pre-flight skip: %v", err)
	}
}

// TestSharedEnginePublicShape_MatviewRefreshFailureTolerated proves the matview
// tolerance the production flow promises: when the objects+data restore
// succeeds but the separate REFRESH MATERIALIZED VIEW invocation fails, the
// migration still COMMITS with a warning (matviewRefreshFailed=true) and the
// rest of the data is present on the target - NOT a rollback (a failed matview
// refresh is a warning, not a migration failure).
//
// The fixture is the documented production failure mode: a matview over a
// plpgsql function with a transitive UNQUALIFIED table reference. At creation
// time the default search_path resolves it; during pg_restore the search_path
// is blank, so REFRESH fails ("relation does not exist") while every other
// object restores cleanly. plpgsql is used (not SQL) so the body is neither
// validated at CREATE FUNCTION time nor inlined when the matview is created
// WITH NO DATA - the failure fires only at REFRESH execution.
func TestSharedEnginePublicShape_MatviewRefreshFailureTolerated(t *testing.T) {
	skipIfNoBinaries(t)

	const publicSQL = `
create table public.base (id int primary key);
insert into public.base values (1),(2),(3);
create function public.base_ids() returns setof int language plpgsql as
$$ begin return query select id from base; end $$;
create materialized view public.mv_unqualified as select * from public.base_ids() as id;`

	verify := func(ctx context.Context, target *pgClusterRef) error {
		conn, err := target.connect(ctx)
		if err != nil {
			return err
		}
		defer conn.Close(ctx)
		// the committed objects+data restore carried the table rows
		var rows int
		if err := conn.QueryRow(ctx, "select count(*) from public.base").Scan(&rows); err != nil {
			return fmt.Errorf("base table missing on target after tolerated refresh failure: %w", err)
		}
		if rows != 3 {
			return fmt.Errorf("base table has %d rows on target, want 3", rows)
		}
		// the matview exists but is unpopulated - the refresh failed, nothing rolled back
		var populated bool
		if err := conn.QueryRow(ctx, "select relispopulated from pg_class where relname = 'mv_unqualified'").Scan(&populated); err != nil {
			return fmt.Errorf("matview missing on target: %w", err)
		}
		if populated {
			return fmt.Errorf("matview is populated - the REFRESH unexpectedly succeeded, so this case no longer exercises the failure path")
		}
		return nil
	}

	res, _, err := runPublicShapeEngine(t, publicSQL, verify)
	if err != nil {
		t.Fatalf("a failed matview refresh must NOT fail the migration, got error: %v (result %+v)", err, res)
	}
	if !res.committed {
		t.Errorf("a failed matview refresh must still COMMIT the migration, got result %+v", res)
	}
	if !res.matviewRefreshFailed {
		t.Errorf("expected matviewRefreshFailed=true so the caller can surface the REFRESH-manually warning")
	}
	if res.oldClusterRetained {
		t.Errorf("a committed migration must not report oldClusterRetained=true")
	}
}

// TestSharedEnginePublicShape_MatviewRefreshSucceeds: the success half of the
// matview split - a well-qualified matview is refreshed by the second
// invocation and arrives populated, with no warning recorded.
func TestSharedEnginePublicShape_MatviewRefreshSucceeds(t *testing.T) {
	skipIfNoBinaries(t)

	const publicSQL = `
create table public.base (id int primary key, val int);
insert into public.base values (1,10),(2,20),(3,30);
create materialized view public.mv_ok as select id, val*2 as doubled from public.base;`

	verify := func(ctx context.Context, target *pgClusterRef) error {
		conn, err := target.connect(ctx)
		if err != nil {
			return err
		}
		defer conn.Close(ctx)
		var total int
		if err := conn.QueryRow(ctx, "select sum(doubled)::int from public.mv_ok").Scan(&total); err != nil {
			return fmt.Errorf("matview not queryable on target (not refreshed?): %w", err)
		}
		if total != 120 {
			return fmt.Errorf("matview content wrong on target: sum(doubled)=%d, want 120", total)
		}
		return nil
	}

	res, _, err := runPublicShapeEngine(t, publicSQL, verify)
	if err != nil {
		t.Fatalf("clean matview migration returned error: %v", err)
	}
	if !res.committed {
		t.Errorf("expected committed=true, got result %+v", res)
	}
	if res.matviewRefreshFailed {
		t.Errorf("matviewRefreshFailed=true on a clean refresh")
	}
}

// TestSharedEngineShapesDiffer documents, at the type level, the exact parameter
// contrast the governing decision mandates: collation pre-check + row-checksum
// validation ON for public, OFF for data tank; refresh-pause coordination ON for
// data tank, OFF for public; the COPY tiers reachable only for data tank.
func TestSharedEngineShapesDiffer(t *testing.T) {
	pub := publicMigrationShape()
	dt := dataTankMigrationShape()

	checks := []struct {
		name      string
		pub, dtnk bool
		wantPub   bool
		wantDt    bool
	}{
		{"collationPreCheck", pub.collationPreCheck, dt.collationPreCheck, true, false},
		{"rowChecksumValidation", pub.rowChecksumValidation, dt.rowChecksumValidation, true, false},
		{"refreshPauseHook", pub.refreshPauseHook, dt.refreshPauseHook, false, true},
		{"matviewRefresh", pub.matviewRefresh, dt.matviewRefresh, true, false},
	}
	for _, c := range checks {
		if c.pub != c.wantPub {
			t.Errorf("public shape %s = %v, want %v", c.name, c.pub, c.wantPub)
		}
		if c.dtnk != c.wantDt {
			t.Errorf("data-tank shape %s = %v, want %v", c.name, c.dtnk, c.wantDt)
		}
	}
	if pub.copyTierCeiling >= dtRestoreTier3PerTable {
		t.Errorf("public shape must NOT reach the COPY tiers; ceiling=%v", pub.copyTierCeiling)
	}
	if dt.copyTierCeiling != dtRestoreTier4PerPartition {
		t.Errorf("data-tank shape must reach the full ladder; ceiling=%v want %v", dt.copyTierCeiling, dtRestoreTier4PerPartition)
	}
}
