package steampipeconfig

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/turbot/steampipe/steampipeconfig/options"
)

type loadConfigTest struct {
	steampipeDir string
	workspaceDir string
	expected     interface{}
}

var trueVal = true
var ttlVal = 300

var databasePort = 9193
var databaseListen = "local"
var databaseSearchPath = "aws,gcp,foo"

var terminalMulti = false
var terminalOutput = "table"
var terminalHeader = true
var terminalSeparator = ","
var terminalTiming = false
var terminalSearchPath = "aws,gcp"
var generalUpdateCheck = "true"

var workspaceMulti = true
var workspaceOutput = "json"
var workspaceSearchPath = "bar,aws,gcp"
var workspaceSearchPathPrefix = "foobar"

var testCasesLoadConfig = map[string]loadConfigTest{
	"multiple_connections": {
		steampipeDir: "testdata/connection_config/multiple_connections",
		expected: &SteampipeConfig{
			Connections: map[string]*modconfig.Connection{
				"aws_dmi_001": {
					Name:   "aws_dmi_001",
					Plugin: "hub.steampipe.io/plugins/turbot/aws@latest",
					Config: `access_key = "aws_dmi_001_access_key"
regions    = "- us-east-1\n-us-west-"
secret_key = "aws_dmi_001_secret_key"`,
				},
				"aws_dmi_002": {
					Name:   "aws_dmi_002",
					Plugin: "hub.steampipe.io/plugins/turbot/aws@latest",
					Config: `access_key = "aws_dmi_002_access_key"
regions    = "- us-east-1\n-us-west-"
secret_key = "aws_dmi_002_secret_key"`,
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			}},
	},
	"single_connection": {
		steampipeDir: "testdata/connection_config/single_connection",
		expected: &SteampipeConfig{
			Connections: map[string]*modconfig.Connection{
				"a": {
					Name:   "a",
					Plugin: "hub.steampipe.io/plugins/testdata/connection-test-1@latest",
					//Config: map[string]string{},
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			}},
	},
	"single_connection_with_default_options": {
		steampipeDir: "testdata/connection_config/single_connection_with_default_options",
		expected: &SteampipeConfig{
			Connections: map[string]*modconfig.Connection{
				"a": {
					Name:   "a",
					Plugin: "hub.steampipe.io/plugins/testdata/connection-test-1@latest",
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			},
			DatabaseOptions: &options.Database{
				Port:       &databasePort,
				Listen:     &databaseListen,
				SearchPath: &databaseSearchPath,
			},
			TerminalOptions: &options.Terminal{
				Output:     &terminalOutput,
				Separator:  &terminalSeparator,
				Header:     &terminalHeader,
				Multi:      &terminalMulti,
				Timing:     &terminalTiming,
				SearchPath: &terminalSearchPath,
			},
			GeneralOptions: &options.General{
				UpdateCheck: &generalUpdateCheck,
			},
		},
	},
	"single_connection_with_default_options_and_workspace_invalid_options_block": {
		steampipeDir: "testdata/connection_config/single_connection_with_default_options",
		workspaceDir: "testdata/workspaces/invalid_options_block",
		expected:     "ERROR",
	},
	"single_connection_with_default_options_and_workspace_search_path_prefix": {
		steampipeDir: "testdata/connection_config/single_connection_with_default_options",
		workspaceDir: "testdata/workspaces/search_path_prefix",
		expected: &SteampipeConfig{
			Connections: map[string]*modconfig.Connection{
				"a": {
					Name:   "a",
					Plugin: "hub.steampipe.io/plugins/testdata/connection-test-1@latest",
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			},
			DatabaseOptions: &options.Database{
				Port:       &databasePort,
				Listen:     &databaseListen,
				SearchPath: &databaseSearchPath,
			},
			TerminalOptions: &options.Terminal{
				Output:           &terminalOutput,
				Separator:        &terminalSeparator,
				Header:           &terminalHeader,
				Multi:            &terminalMulti,
				Timing:           &terminalTiming,
				SearchPath:       &terminalSearchPath,
				SearchPathPrefix: &workspaceSearchPathPrefix,
			},
			GeneralOptions: &options.General{
				UpdateCheck: &generalUpdateCheck,
			},
		},
	},
	"single_connection_with_default_options_and_workspace_override_terminal_config": {
		steampipeDir: "testdata/connection_config/single_connection_with_default_options",
		workspaceDir: "testdata/workspaces/override_terminal_config",
		expected: &SteampipeConfig{
			Connections: map[string]*modconfig.Connection{
				"a": {
					Name:   "a",
					Plugin: "hub.steampipe.io/plugins/testdata/connection-test-1@latest",
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			},
			DatabaseOptions: &options.Database{
				Port:       &databasePort,
				Listen:     &databaseListen,
				SearchPath: &databaseSearchPath,
			},
			TerminalOptions: &options.Terminal{
				Output:           &workspaceOutput,
				Separator:        &terminalSeparator,
				Header:           &terminalHeader,
				Multi:            &workspaceMulti,
				Timing:           &terminalTiming,
				SearchPath:       &workspaceSearchPath,
				SearchPathPrefix: &workspaceSearchPathPrefix,
			},
			GeneralOptions: &options.General{
				UpdateCheck: &generalUpdateCheck,
			},
		},
	},
	"single_connection_with_default_and_connection_options": {
		steampipeDir: "testdata/connection_config/single_connection_with_default_and_connection_options",
		expected: &SteampipeConfig{
			Connections: map[string]*modconfig.Connection{
				"a": {
					Name:   "a",
					Plugin: "hub.steampipe.io/plugins/testdata/connection-test-1@latest",
					Options: &options.Connection{
						Cache:    &trueVal,
						CacheTTL: &ttlVal,
					},
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			},
			DatabaseOptions: &options.Database{
				Port:       &databasePort,
				Listen:     &databaseListen,
				SearchPath: &databaseSearchPath,
			},
			TerminalOptions: &options.Terminal{
				Output:     &terminalOutput,
				Separator:  &terminalSeparator,
				Header:     &terminalHeader,
				Multi:      &terminalMulti,
				Timing:     &terminalTiming,
				SearchPath: &terminalSearchPath,
			},
			GeneralOptions: &options.General{
				UpdateCheck: &generalUpdateCheck,
			},
		},
	},
	"options_only": {
		steampipeDir: "testdata/connection_config/options_only",
		expected: &SteampipeConfig{
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			},
			DatabaseOptions: &options.Database{
				Port:       &databasePort,
				Listen:     &databaseListen,
				SearchPath: &databaseSearchPath,
			},
			TerminalOptions: &options.Terminal{
				Output:     &terminalOutput,
				Separator:  &terminalSeparator,
				Header:     &terminalHeader,
				Multi:      &terminalMulti,
				Timing:     &terminalTiming,
				SearchPath: &terminalSearchPath,
			},
			GeneralOptions: &options.General{
				UpdateCheck: &generalUpdateCheck,
			},
		},
	},
	"options_duplicate_block": {
		steampipeDir: "testdata/connection_config/options_duplicate_block",
		expected:     "ERROR",
	},
}

func TestLoadConfig(t *testing.T) {
	for name, test := range testCasesLoadConfig {
		// default workspoace to empty dir
		workspaceDir := test.workspaceDir
		if workspaceDir == "" {
			workspaceDir = "testdata/workspaces/empty"
		}
		steampipeDir, err := filepath.Abs(test.steampipeDir)
		if err != nil {
			t.Errorf("failed to build absolute config filepath from %s", test.steampipeDir)
		}

		workspaceDir, err = filepath.Abs(workspaceDir)
		if err != nil {
			t.Errorf("failed to build absolute config filepath from %s", workspaceDir)
		}

		// set SteampipeDir
		filepaths.SteampipeDir = steampipeDir

		// now load config
		config, err := loadSteampipeConfig(workspaceDir, "")
		if err != nil {
			if test.expected != "ERROR" {
				t.Errorf("Test: '%s'' FAILED with unexpected error: %v", name, err)
			}
			continue
		}

		if test.expected == "ERROR" {
			t.Errorf("Test: '%s'' FAILED - expected error", name)
			continue
		}

		expectedConfig := test.expected.(*SteampipeConfig)
		if !SteampipeConfigEquals(config, expectedConfig) {
			t.Errorf("Test: '%s'' FAILED : expected:\n%s\n\ngot:\n%s", name, expectedConfig, config)
		}
	}
}

// helpers
func SteampipeConfigEquals(left, right *SteampipeConfig) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	if reflect.DeepEqual(left.Connections, right.Connections) {
		return false
	}

	return reflect.DeepEqual(left.DefaultConnectionOptions, right.DefaultConnectionOptions) &&
		reflect.DeepEqual(left.DatabaseOptions, right.DatabaseOptions) &&
		reflect.DeepEqual(left.TerminalOptions, right.TerminalOptions) &&
		reflect.DeepEqual(left.GeneralOptions, right.GeneralOptions)
}
