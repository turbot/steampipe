package steampipeconfig

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/versionfile"
	"github.com/turbot/steampipe/v2/pkg/constants"
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

func TestLoadConfig_PerFileAndPerBlockFailuresDoNotFailWholeLoad(t *testing.T) {
	tests := []struct {
		name    string
		badFile string
		// addPlugin only checks for duplicates among installed plugins
		chaosInstalled bool
	}{
		{name: "malformed_hcl", badFile: `connection "broken" {
  plugin = "chaos"
`},
		{name: "unknown_block_type", badFile: `widget "w" {}`},
		{name: "connection_missing_plugin", badFile: `connection "no_plugin" {}`},
		{name: "duplicate_plugin_label", badFile: `plugin "chaos_one" {
  source = "chaos"
}
plugin "chaos_one" {
  source = "chaos"
}`, chaosInstalled: true},
		{name: "duplicate_options_block", badFile: `options "database" {
  port = 9193
}
options "database" {
  port = 9194
}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setInstallDir(t)
			dir := t.TempDir()
			writeConfigFile(t, dir, "good.spc", connectionConfig("conn_good"))
			writeConfigFile(t, dir, "bad.spc", tt.badFile)

			config := NewSteampipeConfig("")
			if tt.chaosInstalled {
				config.PluginVersions = map[string]*versionfile.InstalledVersion{
					plugin.ResolvePluginImageRef("chaos"): {},
				}
			}
			res := loadConfig(context.TODO(), dir, config, &loadConfigOptions{include: []string{"*.spc"}})

			if err := res.GetError(); err != nil {
				t.Fatalf("a failure in bad.spc must not fail the whole config load, got error: %v", err)
			}
			if len(res.Warnings) == 0 || !strings.Contains(strings.Join(res.Warnings, "\n"), "bad.spc") {
				t.Errorf("expected a warning naming bad.spc, got %v", res.Warnings)
			}
			if _, ok := config.Connections["conn_good"]; !ok {
				t.Error("the connection in good.spc should still have loaded")
			}
		})
	}
}

// default.spc always exists and holds only comments, so it must not count as loaded config.
func TestLoadConfig_NothingLoadedBesidesBadFileFails(t *testing.T) {
	setInstallDir(t)
	dir := t.TempDir()
	writeConfigFile(t, dir, "default.spc", constants.DefaultConnectionConfigContent)
	writeConfigFile(t, dir, "bad.spc", `connection "broken" {`)

	config := NewSteampipeConfig("")
	res := loadConfig(context.TODO(), dir, config, &loadConfigOptions{include: []string{"*.spc"}})

	if res.GetError() == nil {
		t.Fatal("expected an error when nothing besides the bad file defines any config")
	}
}

func TestLoadConfig_UndecodableOptionsBlockDoesNotBlockLaterOne(t *testing.T) {
	setInstallDir(t)
	dir := t.TempDir()
	writeConfigFile(t, dir, "a.spc", `options "database" {
  port = "not a number"
}`)
	writeConfigFile(t, dir, "b.spc", `options "database" {
  port = 9194
}`)

	config := NewSteampipeConfig("")
	res := loadConfig(context.TODO(), dir, config, &loadConfigOptions{include: []string{"*.spc"}})

	if err := res.GetError(); err != nil {
		t.Fatalf("an undecodable options block must not fail the whole config load, got error: %v", err)
	}
	if config.DatabaseOptions == nil || config.DatabaseOptions.Port == nil || *config.DatabaseOptions.Port != 9194 {
		t.Errorf("the valid database options block should have been applied, got %+v", config.DatabaseOptions)
	}
}
