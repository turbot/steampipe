package db_local

// PostgreSQL 15 removed the default CREATE grant on the public schema (on PG14 it was world-creatable). Steampipe
// users create tables/views in public over their own connections, so setupInternal grants USAGE, CREATE on public to
// the users role at every service start. This test pins both halves of that fact against a REAL PG18 cluster:
// without the grant a non-superuser member of the role is denied CREATE in public (the PG15 behaviour change that
// forced the fix), and after the exact grant statement setupInternal issues, CREATE succeeds. The wired path - a real
// service start followed by CREATE TABLE as the steampipe user - is covered by the acceptance suite
// (service.bats "test crosstab function").

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestPublicSchemaCreateGrant_PG18(t *testing.T) {
	skipIfNoBinaries(t)
	ctx := context.Background()

	base := filepath.Join(dtTestRoot(), "wire", shortTestKey(t.Name()))
	dataDir := filepath.Join(base, "pg18-data")
	sockDir := filepath.Join(base, "s18")
	if err := os.RemoveAll(base); err != nil {
		t.Fatalf("clean base: %v", err)
	}

	if err := dtInitCluster(ctx, dtPG18Version, dataDir); err != nil {
		t.Fatalf("init pg18: %v", err)
	}
	cluster, err := dtStartCluster(ctx, dtPG18Version, dataDir, sockDir)
	if err != nil {
		t.Fatalf("start pg18: %v", err)
	}
	t.Cleanup(cluster.stop)
	if err := cluster.ensureFixtureDB(ctx); err != nil {
		t.Fatalf("ensure db: %v", err)
	}

	conn, err := cluster.connect(ctx, cluster.dbName)
	if err != nil {
		t.Fatalf("connect as superuser: %v", err)
	}
	defer conn.Close(ctx)

	// The role + user shape setupInternal's grant targets, as installDatabaseWithPermissions creates them.
	for _, stmt := range []string{
		fmt.Sprintf("create role %s", constants.DatabaseUsersRole),
		"create user grant_probe login",
		fmt.Sprintf("grant %s to grant_probe", constants.DatabaseUsersRole),
		fmt.Sprintf("grant connect on database %s to %s", cluster.dbName, constants.DatabaseUsersRole),
	} {
		if _, err := conn.Exec(ctx, stmt); err != nil {
			t.Fatalf("setup %q: %v", stmt, err)
		}
	}

	probe, err := pgx.Connect(ctx, fmt.Sprintf("host=%s user=grant_probe dbname=%s sslmode=disable", cluster.sockDir, cluster.dbName))
	if err != nil {
		t.Fatalf("connect as probe user: %v", err)
	}
	defer probe.Close(ctx)

	// The PG15+ behaviour change this guards against: without our grant, CREATE in public is DENIED.
	_, err = probe.Exec(ctx, "create table public.grant_probe_t(id int)")
	if err == nil {
		t.Fatal("PG18 allowed CREATE in public without a grant - the PG15 lockdown this test pins has changed; revisit whether setupInternal's grant is still needed")
	}
	if !strings.Contains(err.Error(), "42501") && !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission-denied creating in public pre-grant, got: %v", err)
	}

	// The exact statement setupInternal issues restores the PG14 behaviour.
	if _, err := conn.Exec(ctx, fmt.Sprintf(`GRANT USAGE, CREATE ON SCHEMA public TO %s;`, constants.DatabaseUsersRole)); err != nil {
		t.Fatalf("grant: %v", err)
	}
	if _, err := probe.Exec(ctx, "create table public.grant_probe_t(id int)"); err != nil {
		t.Fatalf("CREATE in public still denied after setupInternal's grant: %v", err)
	}
}
