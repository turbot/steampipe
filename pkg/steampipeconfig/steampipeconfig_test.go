package steampipeconfig

import (
	"testing"

	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/versionfile"
	"github.com/turbot/steampipe/v2/pkg/options"
)

// TestNewSteampipeConfig removed - trivial constructor test that only verified map initialization
// Documented in cleanup report as providing no regression value

func TestSteampipeConfig_Validate(t *testing.T) {
	tests := []struct {
		name              string
		config            *SteampipeConfig
		expectedWarnings  int
		expectedErrors    int
		expectConnRemoved bool
	}{
		{
			name: "valid_connections",
			config: &SteampipeConfig{
				Connections: map[string]*modconfig.SteampipeConnection{
					"test": {
						Name:         "test",
						PluginAlias:  "test_plugin",
						Type:         "plugin",
						ImportSchema: "enabled",
					},
				},
			},
			expectedWarnings:  0,
			expectedErrors:    0,
			expectConnRemoved: false,
		},
		{
			name: "empty_config",
			config: &SteampipeConfig{
				Connections: map[string]*modconfig.SteampipeConnection{},
			},
			expectedWarnings:  0,
			expectedErrors:    0,
			expectConnRemoved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialConnCount := len(tt.config.Connections)
			warnings, errors := tt.config.Validate()

			if len(warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d: %v", tt.expectedWarnings, len(warnings), warnings)
			}
			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectedErrors, len(errors), errors)
			}

			if tt.expectConnRemoved && len(tt.config.Connections) >= initialConnCount {
				t.Errorf("Expected connection to be removed")
			}
		})
	}
}

func TestSteampipeConfig_ConfigMap(t *testing.T) {
	config := &SteampipeConfig{
		DatabaseOptions: &options.Database{
			Port: intPtr(9193),
		},
		GeneralOptions: &options.General{
			UpdateCheck: stringPtr("false"),
		},
	}

	configMap := config.ConfigMap()
	if configMap == nil {
		t.Fatal("ConfigMap returned nil")
	}

	// The config map should contain values from both DatabaseOptions and GeneralOptions
	// GeneralOptions are set first, so DatabaseOptions should override any conflicts
	if len(configMap) == 0 {
		t.Error("Expected non-empty config map")
	}
}

func TestSteampipeConfig_SetOptions(t *testing.T) {
	tests := []struct {
		name    string
		config  *SteampipeConfig
		opts    interface{}
		checkFn func(*testing.T, *SteampipeConfig)
	}{
		{
			name:   "set_database_options",
			config: NewSteampipeConfig("test"),
			opts: &options.Database{
				Port: intPtr(9193),
			},
			checkFn: func(t *testing.T, c *SteampipeConfig) {
				if c.DatabaseOptions == nil {
					t.Error("DatabaseOptions should not be nil")
				}
				if c.DatabaseOptions.Port == nil || *c.DatabaseOptions.Port != 9193 {
					t.Errorf("Expected Port to be 9193, got %v", c.DatabaseOptions.Port)
				}
			},
		},
		{
			name:   "set_general_options",
			config: NewSteampipeConfig("test"),
			opts: &options.General{
				UpdateCheck: stringPtr("false"),
			},
			checkFn: func(t *testing.T, c *SteampipeConfig) {
				if c.GeneralOptions == nil {
					t.Error("GeneralOptions should not be nil")
				}
				if c.GeneralOptions.UpdateCheck == nil || *c.GeneralOptions.UpdateCheck != "false" {
					t.Errorf("Expected UpdateCheck to be 'false', got %v", c.GeneralOptions.UpdateCheck)
				}
			},
		},
		{
			name: "merge_database_options",
			config: &SteampipeConfig{
				Connections:      make(map[string]*modconfig.SteampipeConnection),
				Plugins:          make(map[string][]*plugin.Plugin),
				PluginsInstances: make(map[string]*plugin.Plugin),
				DatabaseOptions: &options.Database{
					Port: intPtr(9193),
				},
			},
			opts: &options.Database{
				StartTimeout: intPtr(30),
			},
			checkFn: func(t *testing.T, c *SteampipeConfig) {
				if c.DatabaseOptions == nil {
					t.Fatal("DatabaseOptions should not be nil")
				}
				if c.DatabaseOptions.Port == nil || *c.DatabaseOptions.Port != 9193 {
					t.Errorf("Expected Port to remain 9193, got %v", c.DatabaseOptions.Port)
				}
				if c.DatabaseOptions.StartTimeout == nil || *c.DatabaseOptions.StartTimeout != 30 {
					t.Errorf("Expected StartTimeout to be 30, got %v", c.DatabaseOptions.StartTimeout)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cast to the proper interface type based on the concrete type
			var err error
			switch opts := tt.opts.(type) {
			case *options.Database:
				ew := tt.config.SetOptions(opts)
				err = ew.GetError()
			case *options.General:
				ew := tt.config.SetOptions(opts)
				err = ew.GetError()
			case *options.Plugin:
				ew := tt.config.SetOptions(opts)
				err = ew.GetError()
			default:
				t.Fatalf("Unknown options type: %T", tt.opts)
			}

			if err != nil {
				t.Errorf("SetOptions returned unexpected error: %v", err)
			}
			tt.checkFn(t, tt.config)
		})
	}
}

// TestSteampipeConfig_ConnectionNames removed - trivial test that only verified map keys extraction
// Documented in cleanup report

// TestSteampipeConfig_ConnectionList removed - trivial test that only verified map to slice conversion
// Documented in cleanup report

func TestSteampipeConfig_AddPlugin(t *testing.T) {
	tests := []struct {
		name        string
		config      *SteampipeConfig
		plugin      *plugin.Plugin
		expectError bool
	}{
		{
			name: "add_new_plugin",
			config: &SteampipeConfig{
				Connections:      make(map[string]*modconfig.SteampipeConnection),
				Plugins:          make(map[string][]*plugin.Plugin),
				PluginsInstances: make(map[string]*plugin.Plugin),
				PluginVersions: map[string]*versionfile.InstalledVersion{
					"test/plugin": {Version: "1.0.0"},
				},
			},
			plugin: &plugin.Plugin{
				Instance: "test-instance",
				Plugin:   "test/plugin",
				Alias:    "test",
			},
			expectError: false,
		},
		{
			name: "add_duplicate_plugin",
			config: &SteampipeConfig{
				Connections: make(map[string]*modconfig.SteampipeConnection),
				Plugins:     make(map[string][]*plugin.Plugin),
				PluginsInstances: map[string]*plugin.Plugin{
					"test-instance": {
						Instance:        "test-instance",
						FileName:        stringPtr("test.spc"),
						StartLineNumber: intPtr(1),
					},
				},
				PluginVersions: map[string]*versionfile.InstalledVersion{
					"test/plugin": {Version: "1.0.0"},
				},
			},
			plugin: &plugin.Plugin{
				Instance:        "test-instance",
				Plugin:          "test/plugin",
				Alias:           "test",
				FileName:        stringPtr("test2.spc"),
				StartLineNumber: intPtr(5),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.addPlugin(tt.plugin)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSteampipeConfig_GetNonSearchPathConnections(t *testing.T) {
	config := &SteampipeConfig{
		Connections: map[string]*modconfig.SteampipeConnection{
			"conn1": {Name: "conn1"},
			"conn2": {Name: "conn2"},
			"conn3": {Name: "conn3"},
			"conn4": {Name: "conn4"},
		},
	}

	searchPath := []string{"conn1", "conn3"}
	nonSearchPath := config.GetNonSearchPathConnections(searchPath)

	if len(nonSearchPath) != 2 {
		t.Errorf("Expected 2 non-search-path connections, got %d", len(nonSearchPath))
	}

	// Verify the correct connections are returned
	found := make(map[string]bool)
	for _, name := range nonSearchPath {
		found[name] = true
	}

	if found["conn1"] || found["conn3"] {
		t.Error("Search path connections should not be in result")
	}
	if !found["conn2"] || !found["conn4"] {
		t.Error("Non-search-path connections should be in result")
	}
}

// TestSteampipeConfig_String removed - trivial test that only checked string is not empty
// Documented in cleanup report as providing no meaningful validation

// TestSteampipeConfig_ConnectionsForPlugin removed - empty test with no assertions
// Comment explicitly stated "just test that the function doesn't crash" - not a valid test goal
// Documented in cleanup report

func TestSteampipeConfig_ResolvePluginInstance(t *testing.T) {
	config := &SteampipeConfig{
		Connections:      make(map[string]*modconfig.SteampipeConnection),
		Plugins:          make(map[string][]*plugin.Plugin),
		PluginsInstances: make(map[string]*plugin.Plugin),
		PluginVersions: map[string]*versionfile.InstalledVersion{
			"/plugins/test/plugin@latest": {Version: "1.0.0"},
		},
	}

	// Test resolving plugin for a connection
	conn := &modconfig.SteampipeConnection{
		Name:        "test",
		PluginAlias: "test/plugin",
	}

	plugin, err := config.resolvePluginInstanceForConnection(conn)

	// Should create an implicit plugin when none exists
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if plugin == nil {
		t.Error("Expected plugin to be created")
	}
}

func TestSteampipeConfig_ResolvePluginInstance_WithReference(t *testing.T) {
	instanceName := "my-plugin-instance"
	config := &SteampipeConfig{
		Connections: make(map[string]*modconfig.SteampipeConnection),
		Plugins:     make(map[string][]*plugin.Plugin),
		PluginsInstances: map[string]*plugin.Plugin{
			instanceName: {
				Instance: instanceName,
				Plugin:   "test/plugin",
			},
		},
		PluginVersions: map[string]*versionfile.InstalledVersion{},
	}

	// Test with explicit plugin instance reference
	conn := &modconfig.SteampipeConnection{
		Name:           "test",
		PluginAlias:    "test/plugin",
		PluginInstance: &instanceName,
	}

	plugin, err := config.resolvePluginInstanceForConnection(conn)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if plugin == nil {
		t.Error("Expected plugin to be found")
	}
	if plugin.Instance != instanceName {
		t.Errorf("Expected instance %s, got %s", instanceName, plugin.Instance)
	}
}

func TestSteampipeConfig_InitializePlugins(t *testing.T) {
	config := &SteampipeConfig{
		Connections: map[string]*modconfig.SteampipeConnection{
			"test": {
				Name:        "test",
				PluginAlias: "test/plugin",
			},
		},
		Plugins:          make(map[string][]*plugin.Plugin),
		PluginsInstances: make(map[string]*plugin.Plugin),
		PluginVersions: map[string]*versionfile.InstalledVersion{
			"/plugins/test/plugin@latest": {Version: "1.0.0"},
		},
	}

	// Should not panic
	config.initializePlugins()

	// Check that plugin was set
	conn := config.Connections["test"]
	if conn.Plugin == "" {
		t.Error("Expected Plugin to be set")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}
