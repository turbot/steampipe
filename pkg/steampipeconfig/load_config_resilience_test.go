package steampipeconfig

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
)

// setInstallDir points the app at a scratch install dir; several filepath
// helpers panic if it is unset.
func setInstallDir(t *testing.T) {
	t.Helper()
	prev := app_specific.InstallDir
	app_specific.InstallDir = t.TempDir()
	t.Cleanup(func() { app_specific.InstallDir = prev })
}

// writeConfigFile writes a .spc file into dir.
func writeConfigFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}

func connectionConfig(name string) string {
	return `
connection "` + name + `" {
  plugin = "chaos"
}
`
}

// TestLoadConfig_DuplicateConnectionDoesNotFailWholeLoad asserts that a single
// duplicate connection name is reported as a warning and skipped, while every
// other connection still loads.
//
// This matters because loadConfig re-parses the WHOLE config folder and the
// connection watcher aborts the refresh when it returns an error: failing the
// entire load for one bad block stops all connections from being synced (no
// schemas created, no connection config updates applied) for as long as the
// offending block exists, while the file watcher keeps running.
func TestLoadConfig_DuplicateConnectionDoesNotFailWholeLoad(t *testing.T) {
	setInstallDir(t)
	dir := t.TempDir()
	writeConfigFile(t, dir, "a.spc", connectionConfig("conn_a"))
	writeConfigFile(t, dir, "b.spc", connectionConfig("conn_b"))
	// duplicates conn_a, declared in a second file
	writeConfigFile(t, dir, "c.spc", connectionConfig("conn_a"))

	config := NewSteampipeConfig("")
	res := loadConfig(context.TODO(), dir, config, &loadConfigOptions{include: []string{"*.spc"}})

	if err := res.GetError(); err != nil {
		t.Fatalf("a duplicate connection must not fail the whole config load, got error: %v", err)
	}
	if len(res.Warnings) == 0 {
		t.Error("expected a warning describing the duplicate connection")
	}
	for _, name := range []string{"conn_a", "conn_b"} {
		if _, ok := config.Connections[name]; !ok {
			t.Errorf("connection %q should still have loaded despite the duplicate", name)
		}
	}
}

// TestLoadConfig_InvalidConnectionNameDoesNotFailWholeLoad asserts the same
// contract for a connection whose name is not a valid schema name.
func TestLoadConfig_InvalidConnectionNameDoesNotFailWholeLoad(t *testing.T) {
	setInstallDir(t)
	dir := t.TempDir()
	writeConfigFile(t, dir, "good.spc", connectionConfig("conn_good"))
	// "pg_catalog" is reserved, so this connection name is invalid
	writeConfigFile(t, dir, "bad.spc", connectionConfig("pg_catalog"))

	config := NewSteampipeConfig("")
	res := loadConfig(context.TODO(), dir, config, &loadConfigOptions{include: []string{"*.spc"}})

	if err := res.GetError(); err != nil {
		t.Fatalf("an invalid connection name must not fail the whole config load, got error: %v", err)
	}
	if _, ok := config.Connections["conn_good"]; !ok {
		t.Error("the valid connection should still have loaded")
	}
	if _, ok := config.Connections["pg_catalog"]; ok {
		t.Error("the invalid connection must be skipped")
	}
}
