package pluginmanager_service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/plugin"
	sdkshared "github.com/turbot/steampipe-plugin-sdk/v6/grpc/shared"
	"github.com/turbot/steampipe/v2/pkg/options"
)

// A plugin's memory limit (memory_max_mb) sets GOMEMLIMIT for the plugin process.
//
// These tests start a plugin process with the real startPluginProcess, so the go-plugin client
// builds the plugin's environment as it does for a real plugin. The plugin is a shell script.
// It writes the variables it was started with to a file, then exits. So startPluginProcess
// returns an error, as the script does not complete the plugin handshake. That is expected.

const testPluginEnvMarker = "STEAMPIPE_TEST_PLUGIN_ENV_MARKER"

// pluginEnv is what the test plugin wrote: each variable's value, or "<not set>"
type pluginEnv struct {
	gomemlimit  string
	marker      string
	magicCookie string
}

// startTestPlugin installs the test plugin in a temporary install directory and starts it
// with startPluginProcess. It returns the variables the plugin was started with.
func startTestPlugin(t *testing.T, pluginConfig *plugin.Plugin) pluginEnv {
	t.Helper()

	installDir := app_specific.InstallDir
	app_specific.InstallDir = t.TempDir()
	t.Cleanup(func() { app_specific.InstallDir = installDir })

	pluginFolder := filepath.Join(filepaths.EnsurePluginDir(), pluginConfig.Plugin)
	require.NoError(t, os.MkdirAll(pluginFolder, 0755))
	outFile := filepath.Join(t.TempDir(), "env.txt")
	script := fmt.Sprintf(`#!/bin/sh
{
  echo "${GOMEMLIMIT-<not set>}"
  echo "${%s-<not set>}"
  echo "${%s-<not set>}"
} > '%s'
`, testPluginEnvMarker, sdkshared.Handshake.MagicCookieKey, outFile)
	require.NoError(t, os.WriteFile(filepath.Join(pluginFolder, "steampipe-plugin-envtest.plugin"), []byte(script), 0700)) //nolint:gosec // G306: the test plugin must be executable

	pm := newTestPluginManager(t)
	pm.plugins[pluginConfig.Instance] = pluginConfig

	client, err := pm.startPluginProcess(pluginConfig.Instance, nil)
	require.Error(t, err, "the test plugin does not complete the handshake, so the start must fail")
	if client != nil {
		client.Kill()
	}

	out, err := os.ReadFile(outFile)
	require.NoError(t, err, "the test plugin did not run")
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	require.Len(t, lines, 3, "test plugin output: %q", out)
	return pluginEnv{gomemlimit: lines[0], marker: lines[1], magicCookie: lines[2]}
}

// setPluginOptions sets the plugin options (options "plugin" { memory_max_mb }) as the CLI does
func setPluginOptions(t *testing.T, memoryMaxMb *int) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	for k, v := range (&options.Plugin{MemoryMaxMb: memoryMaxMb}).ConfigMap() {
		viper.SetDefault(k, v)
	}
}

func intPtr(i int) *int { return &i }

func TestStartPluginProcess_PluginMemoryLimit(t *testing.T) {
	tests := []struct {
		name string
		// GOMEMLIMIT in the plugin manager's environment ("" is not set)
		hostGOMEMLIMIT string
		// memory_max_mb in the plugin block
		pluginMemoryMaxMb *int
		// memory_max_mb in the plugin options
		optionsMemoryMaxMb *int
		wantGOMEMLIMIT     string
	}{
		{
			name:              "plugin memory_max_mb",
			pluginMemoryMaxMb: intPtr(512),
			wantGOMEMLIMIT:    "536870912",
		},
		{
			name:              "plugin memory_max_mb replaces a lower GOMEMLIMIT in the environment",
			hostGOMEMLIMIT:    "64MiB",
			pluginMemoryMaxMb: intPtr(512),
			wantGOMEMLIMIT:    "536870912",
		},
		{
			name:              "plugin memory_max_mb replaces a higher GOMEMLIMIT in the environment",
			hostGOMEMLIMIT:    "4GiB",
			pluginMemoryMaxMb: intPtr(512),
			wantGOMEMLIMIT:    "536870912",
		},
		{
			name:               "plugin options memory_max_mb replaces GOMEMLIMIT in the environment",
			hostGOMEMLIMIT:     "64MiB",
			optionsMemoryMaxMb: intPtr(256),
			wantGOMEMLIMIT:     "268435456",
		},
		{
			name:               "memory_max_mb 0: the plugin keeps GOMEMLIMIT from the environment",
			hostGOMEMLIMIT:     "64MiB",
			pluginMemoryMaxMb:  intPtr(0),
			optionsMemoryMaxMb: intPtr(0),
			wantGOMEMLIMIT:     "64MiB",
		},
		{
			name:               "memory_max_mb 0 and no GOMEMLIMIT in the environment: no limit",
			pluginMemoryMaxMb:  intPtr(0),
			optionsMemoryMaxMb: intPtr(0),
			wantGOMEMLIMIT:     "<not set>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setPluginOptions(t, tt.optionsMemoryMaxMb)
			if tt.hostGOMEMLIMIT == "" {
				// t.Setenv restores the variable after the test
				t.Setenv("GOMEMLIMIT", "")
				require.NoError(t, os.Unsetenv("GOMEMLIMIT"))
			} else {
				t.Setenv("GOMEMLIMIT", tt.hostGOMEMLIMIT)
			}
			t.Setenv(testPluginEnvMarker, "from-the-plugin-manager")

			env := startTestPlugin(t, &plugin.Plugin{
				Instance:    "envtest",
				Alias:       "envtest",
				Plugin:      "hub.steampipe.io/plugins/turbot/envtest@latest",
				MemoryMaxMb: tt.pluginMemoryMaxMb,
			})

			assert.Equal(t, tt.wantGOMEMLIMIT, env.gomemlimit, "plugin GOMEMLIMIT")
			assert.Equal(t, "from-the-plugin-manager", env.marker, "the plugin must get the plugin manager's environment")
			assert.Equal(t, sdkshared.Handshake.MagicCookieValue, env.magicCookie, "the plugin must get go-plugin's handshake variables")
		})
	}
}
