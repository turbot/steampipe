package steampipeconfig

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/versionfile"
)

const unreadablePluginConfig = `
connection "conn_bad" {
  plugin = "chaos"
}
connection "conn_good" {
  plugin = "chaosdynamic"
}
`

// installFakePlugin writes an empty plugin binary where steampipe expects to find it
func installFakePlugin(t *testing.T, alias string) (imageRef, binaryPath string) {
	t.Helper()
	imageRef = plugin.ResolvePluginImageRef(alias)
	binaryPath = filepaths.PluginBinaryPath(imageRef, alias)
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0755); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}
	if err := os.WriteFile(binaryPath, []byte("fake"), 0700); err != nil {
		t.Fatalf("failed to write plugin binary: %v", err)
	}
	return imageRef, binaryPath
}

func loadUnreadablePluginConfig(t *testing.T, versions map[string]*versionfile.InstalledVersion) *SteampipeConfig {
	t.Helper()
	dir := t.TempDir()
	writeConfigFile(t, dir, "a.spc", unreadablePluginConfig)
	config := NewSteampipeConfig("")
	config.PluginVersions = versions
	res := loadConfig(context.TODO(), dir, config, &loadConfigOptions{include: []string{"*.spc"}})
	if err := res.GetError(); err != nil {
		t.Fatalf("a plugin that cannot be read must not fail the whole config load, got: %v", err)
	}
	return config
}

func assertGoodConnectionLoads(t *testing.T, config *SteampipeConfig) {
	t.Helper()
	good := config.Connections["conn_good"]
	if good == nil {
		t.Fatal("conn_good should have loaded")
	}
	if good.Error != nil {
		t.Errorf("conn_good uses another plugin and must still load, got error: %v", good.Error)
	}
	if good.PluginPath == nil {
		t.Error("conn_good should have a plugin path")
	}
}

func assertUnreadableError(t *testing.T, config *SteampipeConfig) {
	t.Helper()
	bad := config.Connections["conn_bad"]
	if bad == nil {
		t.Fatal("conn_bad should have loaded")
	}
	if bad.Error == nil {
		t.Fatal("conn_bad should carry an error")
	}
	msg := bad.Error.Error()
	if msg == pconstants.ConnectionErrorPluginNotInstalled || strings.Contains(msg, "not installed") {
		t.Errorf("an unreadable plugin must not be reported as not installed, got: %s", msg)
	}
	if !strings.Contains(msg, "chaos") {
		t.Errorf("the error should name the plugin, got: %s", msg)
	}
	if !strings.Contains(msg, "cannot be read") {
		t.Errorf("the error should say the plugin cannot be read, got: %s", msg)
	}
}

// A plugin binary that is not in the version file (a locally built plugin) and cannot be
// stat'ed for a reason other than being absent is an error naming the plugin and the reason.
func TestInitializePlugins_UnreadableLocalPluginBinaryIsAnError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("needs symlinks")
	}
	setInstallDir(t)
	chaosRef, chaosBinary := installFakePlugin(t, "chaos")
	// a symlink pointing at itself makes os.Stat fail with "too many levels of symbolic links"
	if err := os.Remove(chaosBinary); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(chaosBinary, chaosBinary); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	goodRef, _ := installFakePlugin(t, "chaosdynamic")

	config := loadUnreadablePluginConfig(t, map[string]*versionfile.InstalledVersion{
		goodRef: {Version: "1.0.0"},
	})

	assertUnreadableError(t, config)
	if msg := config.Connections["conn_bad"].Error.Error(); !strings.Contains(msg, "too many levels of symbolic links") && !strings.Contains(msg, "symbolic link") {
		t.Errorf("the error should give the reason, got: %s", msg)
	}
	if _, ok := config.PluginVersions[chaosRef]; ok {
		t.Error("an unreadable plugin must not be added to the plugin versions as a local plugin")
	}
	assertGoodConnectionLoads(t, config)
}

// An installed plugin (listed in the version file) whose folder cannot be read must be
// reported as unreadable, not as not installed.
func TestInitializePlugins_UnreadableInstalledPluginFolderIsAnError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permissions are not enforced the same way on windows")
	}
	setInstallDir(t)
	chaosRef, chaosBinary := installFakePlugin(t, "chaos")
	goodRef, _ := installFakePlugin(t, "chaosdynamic")

	pluginDir := filepath.Dir(chaosBinary)
	if err := os.Chmod(pluginDir, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(pluginDir, 0755) })
	if _, err := os.Stat(chaosBinary); err == nil {
		t.Skip("permissions are not enforced for this user (running as root?)")
	}

	config := loadUnreadablePluginConfig(t, map[string]*versionfile.InstalledVersion{
		chaosRef: {Version: "1.0.0"},
		goodRef:  {Version: "1.0.0"},
	})

	assertUnreadableError(t, config)
	if config.Connections["conn_bad"].PluginPath != nil {
		t.Error("an unreadable plugin must not have a plugin path")
	}
	assertGoodConnectionLoads(t, config)
}

// A plugin that really is not installed is still reported as not installed.
func TestInitializePlugins_MissingPluginIsStillNotInstalled(t *testing.T) {
	tests := []struct {
		name string
		// whether chaos is listed in the version file although its binary is gone
		inVersionFile bool
	}{
		{name: "not_in_version_file"},
		{name: "in_version_file_but_binary_missing", inVersionFile: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setInstallDir(t)
			goodRef, _ := installFakePlugin(t, "chaosdynamic")
			versions := map[string]*versionfile.InstalledVersion{goodRef: {Version: "1.0.0"}}
			if tt.inVersionFile {
				versions[plugin.ResolvePluginImageRef("chaos")] = &versionfile.InstalledVersion{Version: "1.0.0"}
			}

			config := loadUnreadablePluginConfig(t, versions)

			bad := config.Connections["conn_bad"]
			if bad == nil || bad.Error == nil {
				t.Fatalf("conn_bad should carry a not installed error, got %+v", bad)
			}
			if got := bad.Error.Error(); got != pconstants.ConnectionErrorPluginNotInstalled {
				t.Errorf("error = %q, want %q", got, pconstants.ConnectionErrorPluginNotInstalled)
			}
			assertGoodConnectionLoads(t, config)
		})
	}
}
