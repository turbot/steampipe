package steampipeconfig

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/steampipeconfig/options"
)

type loadConfigTest struct {
	steampipeDir string
	workspaceDir string
	expected     interface{}
}

var trueVal = true
var ttlVal = 300

var cache_ttl = 300
var databasePort = 9193
var databaseListen = "local"
var databaseSearchPath = []string{"aws", "gcp", "foo"}

var terminalMulti = false
var terminalOutput = "table"
var terminalHeader = true
var terminalSeparator = ","
var terminalTiming = false
var terminalSearchPath = []string{"aws", "gcp"}
var generalUpdateCheck = "true"

var workspaceSearchPath = []string{"bar", "aws", "gcp"}
var workspaceSearchPathPrefix = []string{"aws"}

var testCasesLoadConfig = map[string]loadConfigTest{
	"multiple_connections": {
		steampipeDir: "test_data/connection_config/multiple_connections",
		workspaceDir: "test_data/workspaces/empty",
		expected: &SteampipeConfig{
			Connections: map[string]*Connection{
				"aws_dmi_001": {
					Name:   "aws_dmi_001",
					Plugin: "hub.steampipe.io/plugins/turbot/aws@latest",
					Config: `access_key            = "aws_dmi_001_access_key"
	regions               = "- us-east-1\n-us-west-"
	secret_key            = "aws_dmi_001_secret_key"`,
				},
				"aws_dmi_002": {
					Name:   "aws_dmi_002",
					Plugin: "hub.steampipe.io/plugins/turbot/aws@latest",
					Config: `access_key            = "aws_dmi_002_access_key"
	regions               = "- us-east-1\n-us-west-"
	secret_key            = "aws_dmi_002_secret_key"`,
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			}},
	},
	"single_connection": {
		steampipeDir: "test_data/connection_config/single_connection",
		expected: &SteampipeConfig{
			Connections: map[string]*Connection{
				// todo normalise plugin names here?
				"a": {
					Name:   "a",
					Plugin: "hub.steampipe.io/plugins/test_data/connection-test-1@latest",
					//Config: map[string]string{},
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			}},
	},
	"single_connection_with_default_options": {
		steampipeDir: "test_data/connection_config/single_connection_with_default_options",
		expected: &SteampipeConfig{
			Connections: map[string]*Connection{
				// todo normalise plugin names here?
				"a": {
					Name:   "a",
					Plugin: "hub.steampipe.io/plugins/test_data/connection-test-1@latest",
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
	"single_connection_with_default_and_connection_options": {
		steampipeDir: "test_data/connection_config/single_connection_with_default_and_connection_options",
		expected: &SteampipeConfig{
			Connections: map[string]*Connection{
				"a": {
					Name:   "a",
					Plugin: "hub.steampipe.io/plugins/test_data/connection-test-1@latest",
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
		steampipeDir: "test_data/connection_config/options_only",
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
		steampipeDir: "test_data/connection_config/options_duplicate_block",
		expected:     "ERROR",
	},
}

func TestLoadConfig(t *testing.T) {
	for name, test := range testCasesLoadConfig {
		steampipeDir, err := filepath.Abs(test.steampipeDir)
		workspaceDir, err := filepath.Abs(test.workspaceDir)
		if err != nil {
			t.Errorf("failed to build absolute config filepath from %s", test.steampipeDir)
		}

		// set SteampipeDir
		constants.SteampipeDir = steampipeDir

		// now load config
		config, err := newSteampipeConfig(workspaceDir)
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
			fmt.Printf("")
			t.Errorf("Test: '%s'' FAILED : expected:\n%s\n\ngot:\n%s", name, expectedConfig, config)
		}
	}
}

// helpers
func SteampipeConfigEquals(l, r *SteampipeConfig) bool {
	if l == nil || r == nil {
		return l == nil && r == nil
	}

	for k, lConn := range l.Connections {
		rConn, ok := r.Connections[k]
		if !ok {
			return false
		}
		if lConn.String() != rConn.String() {
			fmt.Printf("Connections different: l:\n%s\nr:\n%s\n", lConn.String(), r.Connections[k].String())
			return false
		}
	}
	for k := range r.Connections {
		if _, ok := l.Connections[k]; !ok {
			return false
		}
	}
	return l.DefaultConnectionOptions.String() == r.DefaultConnectionOptions.String() &&
		l.DatabaseOptions.String() == r.DatabaseOptions.String() &&
		l.TerminalOptions.String() == r.TerminalOptions.String() &&
		l.GeneralOptions.String() == r.GeneralOptions.String()
}
