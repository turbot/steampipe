package cmdconfig

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
)

// workspaceLoaderEnv points the install dir at a scratch folder, with viper and the app settings
// the loader reads put back afterwards. It returns the config folder.
func workspaceLoaderEnv(t *testing.T) string {
	t.Helper()
	installDir := t.TempDir()
	t.Setenv("STEAMPIPE_INSTALL_DIR", installDir)
	t.Setenv("STEAMPIPE_WORKSPACE_PROFILES_LOCATION", "")
	os.Unsetenv("STEAMPIPE_WORKSPACE_PROFILES_LOCATION")
	t.Setenv("STEAMPIPE_WORKSPACE", "")
	os.Unsetenv("STEAMPIPE_WORKSPACE")

	prevExt := app_specific.ConfigExtension
	app_specific.ConfigExtension = ".spc"
	t.Cleanup(func() {
		app_specific.ConfigExtension = prevExt
		viper.Reset()
	})
	viper.Reset()

	configDir := filepath.Join(installDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	return configDir
}

const goodWorkspace = `
workspace "default" {
  snapshot_location = "mine"
}
`

func TestGetWorkspaceProfileLoader_MalformedFileDoesNotStopLoad(t *testing.T) {
	configDir := workspaceLoaderEnv(t)
	for name, content := range map[string]string{
		"workspaces.spc": goodWorkspace,
		"broken.spc":     "connection \"x\" {\n",
	} {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	loader, warnings, err := getWorkspaceProfileLoader(context.Background())
	if err != nil {
		t.Fatalf("a malformed file should not fail the load: %v", err)
	}
	// the valid file is still used
	if got := loader.GetActiveWorkspaceProfile().SnapshotLocation; got == nil || *got != "mine" {
		t.Errorf("the valid file was not used, snapshot_location = %v", got)
	}
	if len(warnings) == 0 || !strings.Contains(strings.Join(warnings, "\n"), filepath.Join(configDir, "broken.spc")) {
		t.Errorf("want a warning naming broken.spc, got %v", warnings)
	}
	// the bad file is not touched
	if _, err := os.Stat(filepath.Join(configDir, "broken.spc")); err != nil {
		t.Errorf("broken.spc should remain: %v", err)
	}
}

func TestGetWorkspaceProfileLoader_MalformedFileInOtherFolderIsNamed(t *testing.T) {
	workspaceLoaderEnv(t)
	otherDir := t.TempDir()
	t.Setenv("STEAMPIPE_WORKSPACE_PROFILES_LOCATION", otherDir)
	for name, content := range map[string]string{
		"workspaces.spc": goodWorkspace,
		"broken.spc":     "workspace \"x\" {\n",
	} {
		if err := os.WriteFile(filepath.Join(otherDir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	loader, warnings, err := getWorkspaceProfileLoader(context.Background())
	if err != nil {
		t.Fatalf("a malformed file should not fail the load: %v", err)
	}
	if got := loader.GetActiveWorkspaceProfile().SnapshotLocation; got == nil || *got != "mine" {
		t.Errorf("the valid file was not used, snapshot_location = %v", got)
	}
	if len(warnings) == 0 || !strings.Contains(strings.Join(warnings, "\n"), filepath.Join(otherDir, "broken.spc")) {
		t.Errorf("want a warning naming broken.spc, got %v", warnings)
	}
}

func TestGetWorkspaceProfileLoader_ErrorInValidFileNamesRealFolder(t *testing.T) {
	configDir := workspaceLoaderEnv(t)
	files := map[string]string{
		"broken.spc": "connection \"x\" {\n",
		// parses, but is not a valid workspace profile
		"bad_workspace.spc": "workspace \"default\" {\n  not_an_argument = 1\n}\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	_, _, err := getWorkspaceProfileLoader(context.Background())
	if err == nil {
		t.Fatal("a workspace profile which does not decode should still fail the load")
	}
	if !strings.Contains(err.Error(), filepath.Join(configDir, "bad_workspace.spc")) || strings.Contains(err.Error(), "steampipe-config-") {
		t.Errorf("the error should name the file in the real folder, got: %v", err)
	}
}

// a typo in the file holding the selected workspace must still name the file: the error alone says only
// that the workspace is not found
func TestGetWorkspaceProfileLoader_SelectedWorkspaceInMalformedFileIsNamed(t *testing.T) {
	configDir := workspaceLoaderEnv(t)
	t.Setenv("STEAMPIPE_WORKSPACE", "prod")
	if err := os.WriteFile(filepath.Join(configDir, "prod.spc"), []byte("workspace \"prod\" {\n  snapshot_location = \"mine\"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, warnings, err := getWorkspaceProfileLoader(context.Background())
	if err == nil {
		t.Fatal("a workspace which cannot be found should still fail the load")
	}
	if !strings.Contains(strings.Join(warnings, "\n"), filepath.Join(configDir, "prod.spc")) {
		t.Errorf("want a warning naming prod.spc beside the error %q, got %v", err, warnings)
	}
}

func TestGetWorkspaceProfileLoader_NoBadFileLoadsAsBefore(t *testing.T) {
	configDir := workspaceLoaderEnv(t)
	if err := os.WriteFile(filepath.Join(configDir, "workspaces.spc"), []byte(goodWorkspace), 0600); err != nil {
		t.Fatal(err)
	}
	loader, warnings, err := getWorkspaceProfileLoader(context.Background())
	if err != nil || len(warnings) != 0 {
		t.Fatalf("want a clean load, got %v, %v", warnings, err)
	}
	if got := loader.GetActiveWorkspaceProfile().SnapshotLocation; got == nil || *got != "mine" {
		t.Errorf("snapshot_location = %v", got)
	}
}
