package steampipeconfig

import (
	utils "github.com/turbot/steampipe/pkg/utils"
	"golang.org/x/exp/maps"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
)

// TODO KAI add plugin block tests

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
var databaseQueryTimeout int64 = 240

var terminalMulti = false
var terminalOutput = "table"
var terminalHeader = true
var terminalSeparator = ","
var terminalTiming = false
var terminalSearchPath = "aws,gcp"
var generalUpdateCheck = "true"
var terminalAutoComplete = true

var workspaceMulti = true
var workspaceAutoComplete = true
var workspaceOutput = "json"
var workspaceSearchPath = "bar,aws,gcp"
var workspaceSearchPathPrefix = "foobar"

var testCasesLoadConfig = map[string]loadConfigTest{
	"multiple_connections": {
		steampipeDir: "testdata/hcl_helpers/multiple_connections",
		expected: &SteampipeConfig{
			Connections: map[string]*modconfig.Connection{
				"aws_dmi_001": {
					Name:           "aws_dmi_001",
					PluginAlias:    "aws",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					PluginInstance: utils.ToStringPointer("hub.steampipe.io/plugins/turbot/aws@latest"),
					Type:           "",
					ImportSchema:   "enabled",
					Config:         "access_key = \"aws_dmi_001_access_key\"\nregions    = \"- us-east-1\\n-us-west-\"\nsecret_key = \"aws_dmi_001_secret_key\"\n",
					DeclRange: modconfig.Range{
						Filename: "$$test_pwd$$/testdata/hcl_helpers/multiple_connections/config/connection1.spc",
						Start: modconfig.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: modconfig.Pos{
							Line:   1,
							Column: 11,
							Byte:   10,
						},
					},
				},
				"aws_dmi_002": {
					Name:           "aws_dmi_002",
					PluginAlias:    "aws",
					Plugin:         "hub.steampipe.io/plugins/turbot/aws@latest",
					PluginInstance: utils.ToStringPointer("hub.steampipe.io/plugins/turbot/aws@latest"),
					Type:           "",
					ImportSchema:   "enabled",
					Config:         "access_key = \"aws_dmi_002_access_key\"\nregions    = \"- us-east-1\\n-us-west-\"\nsecret_key = \"aws_dmi_002_secret_key\"\n",
					DeclRange: modconfig.Range{
						Filename: "$$test_pwd$$/testdata/hcl_helpers/multiple_connections/config/connection2.spc",
						Start: modconfig.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: modconfig.Pos{
							Line:   1,
							Column: 11,
							Byte:   10,
						},
					},
				},
			},
			DefaultConnectionOptions: &options.Connection{
				Cache:    &trueVal,
				CacheTTL: &ttlVal,
			},
		},
	},
	//"single_connection": {
	//	steampipeDir: "testdata/hcl_helpers/single_connection",
	//	expected: &SteampipeConfig{
	//		Connections: map[string]*modconfig.Connection{
	//			"a": {
	//				Name:           "a",
	//				PluginAlias:    "test_data/connection-test-1",
	//				Plugin:         "hub.steampipe.io/plugins/test_data/connection-test-1@latest",
	//				PluginInstance: utils.ToStringPointer("hub.steampipe.io/plugins/test_data/connection-test-1@latest"),
	//				Type:           "",
	//				ImportSchema:   "enabled",
	//				Config:         "",
	//				DeclRange: modconfig.Range{
	//					Filename: "$$test_pwd$$/testdata/hcl_helpers/single_connection/config/connection1.spc",
	//					Start: modconfig.Pos{
	//						Line:   1,
	//						Column: 1,
	//						Byte:   0,
	//					},
	//					End: modconfig.Pos{
	//						Line:   1,
	//						Column: 11,
	//						Byte:   10,
	//					},
	//				},
	//			},
	//		},
	//		DefaultConnectionOptions: &options.Connection{
	//			Cache:    &trueVal,
	//			CacheTTL: &ttlVal,
	//		},
	//	},
	//},
	//"single_connection_with_default_options": { // fixed
	//	steampipeDir: "testdata/hcl_helpers/single_connection_with_default_options",
	//	expected: &SteampipeConfig{
	//		Connections: map[string]*modconfig.Connection{
	//			"a": {
	//				Name:           "a",
	//				PluginAlias:    "test_data/connection-test-1",
	//				Plugin:         "hub.steampipe.io/plugins/test_data/connection-test-1@latest",
	//				PluginInstance: utils.ToStringPointer("hub.steampipe.io/plugins/test_data/connection-test-1@latest"),
	//				Type:           "",
	//				ImportSchema:   "enabled",
	//				Config:         "",
	//				DeclRange: modconfig.Range{
	//					Filename: "$$test_pwd$$/testdata/hcl_helpers/single_connection_with_default_options/config/connection1.spc",
	//					Start: modconfig.Pos{
	//						Line:   1,
	//						Column: 1,
	//						Byte:   0,
	//					},
	//					End: modconfig.Pos{
	//						Line:   1,
	//						Column: 11,
	//						Byte:   10,
	//					},
	//				},
	//			},
	//		},
	//		DefaultConnectionOptions: &options.Connection{
	//			Cache:    &trueVal,
	//			CacheTTL: &ttlVal,
	//		},
	//		DatabaseOptions: &options.Database{
	//			Port:       &databasePort,
	//			Listen:     &databaseListen,
	//			SearchPath: &databaseSearchPath,
	//		},
	//		GeneralOptions: &options.General{
	//			UpdateCheck: &generalUpdateCheck,
	//		},
	//	},
	//},
	//"single_connection_with_default_options_and_workspace_invalid_options_block": { // fixed
	//	steampipeDir: "testdata/hcl_helpers/single_connection_with_default_options",
	//	workspaceDir: "testdata/load_config_test/invalid_options_block",
	//	expected:     "ERROR",
	//},
	//"single_connection_with_default_options_and_workspace_search_path_prefix": { // fixed
	//	steampipeDir: "testdata/hcl_helpers/single_connection_with_default_options",
	//	workspaceDir: "testdata/load_config_test/search_path_prefix",
	//	expected: &SteampipeConfig{
	//		Connections: map[string]*modconfig.Connection{
	//			"a": {
	//				Name:           "a",
	//				PluginAlias:    "test_data/connection-test-1",
	//				Plugin:         "hub.steampipe.io/plugins/test_data/connection-test-1@latest",
	//				PluginInstance: utils.ToStringPointer("hub.steampipe.io/plugins/test_data/connection-test-1@latest"),
	//				Type:           "",
	//				ImportSchema:   "enabled",
	//				Config:         "",
	//				DeclRange: modconfig.Range{
	//					Filename: "$$test_pwd$$/testdata/hcl_helpers/single_connection_with_default_options/config/connection1.spc",
	//					Start: modconfig.Pos{
	//						Line:   1,
	//						Column: 1,
	//						Byte:   0,
	//					},
	//					End: modconfig.Pos{
	//						Line:   1,
	//						Column: 11,
	//						Byte:   10,
	//					},
	//				},
	//			},
	//		},
	//		DefaultConnectionOptions: &options.Connection{
	//			Cache:    &trueVal,
	//			CacheTTL: &ttlVal,
	//		},
	//		DatabaseOptions: &options.Database{
	//			Port:       &databasePort,
	//			Listen:     &databaseListen,
	//			SearchPath: &databaseSearchPath,
	//		},
	//		GeneralOptions: &options.General{
	//			UpdateCheck: &generalUpdateCheck,
	//		},
	//	},
	//},
	//"single_connection_with_default_options_and_workspace_override_terminal_config": { // fixed
	//	steampipeDir: "testdata/hcl_helpers/single_connection_with_default_options",
	//	workspaceDir: "testdata/load_config_test/override_terminal_config",
	//	expected: &SteampipeConfig{
	//		Connections: map[string]*modconfig.Connection{
	//			"a": {
	//				Name:           "a",
	//				PluginAlias:    "test_data/connection-test-1",
	//				Plugin:         "hub.steampipe.io/plugins/test_data/connection-test-1@latest",
	//				PluginInstance: utils.ToStringPointer("hub.steampipe.io/plugins/test_data/connection-test-1@latest"),
	//				Type:           "",
	//				ImportSchema:   "enabled",
	//				Config:         "",
	//				DeclRange: modconfig.Range{
	//					Filename: "$$test_pwd$$/testdata/hcl_helpers/single_connection_with_default_options/config/connection1.spc",
	//					Start: modconfig.Pos{
	//						Line:   1,
	//						Column: 1,
	//						Byte:   0,
	//					},
	//					End: modconfig.Pos{
	//						Line:   1,
	//						Column: 11,
	//						Byte:   10,
	//					},
	//				},
	//			},
	//		},
	//		DefaultConnectionOptions: &options.Connection{
	//			Cache:    &trueVal,
	//			CacheTTL: &ttlVal,
	//		},
	//		DatabaseOptions: &options.Database{
	//			Port:       &databasePort,
	//			Listen:     &databaseListen,
	//			SearchPath: &databaseSearchPath,
	//		},
	//		GeneralOptions: &options.General{
	//			UpdateCheck: &generalUpdateCheck,
	//		},
	//	},
	//},
	//"single_connection_with_default_and_connection_options": {
	//	steampipeDir: "testdata/hcl_helpers/single_connection_with_default_and_connection_options",
	//	expected: &SteampipeConfig{
	//		Connections: map[string]*modconfig.Connection{
	//			"a": {
	//				Name:           "a",
	//				ImportSchema:   "enabled",
	//				PluginAlias:    "test_data/connection-test-1",
	//				Plugin:         "hub.steampipe.io/plugins/test_data/connection-test-1@latest",
	//				PluginInstance: utils.ToStringPointer("hub.steampipe.io/plugins/test_data/connection-test-1@latest"),
	//				Config:         "",
	//				Options: &options.Connection{
	//					Cache:    &trueVal,
	//					CacheTTL: &ttlVal,
	//				},
	//				DeclRange: modconfig.Range{
	//					Filename: "$$test_pwd$$/testdata/hcl_helpers/single_connection_with_default_and_connection_options/config/connection1.spc",
	//					Start: modconfig.Pos{
	//						Line:   1,
	//						Column: 1,
	//						Byte:   0,
	//					},
	//					End: modconfig.Pos{
	//						Line:   1,
	//						Column: 11,
	//						Byte:   10,
	//					},
	//				},
	//			},
	//		},
	//		DefaultConnectionOptions: &options.Connection{
	//			Cache:    &trueVal,
	//			CacheTTL: &ttlVal,
	//		},
	//		DatabaseOptions: &options.Database{
	//			Port:       &databasePort,
	//			Listen:     &databaseListen,
	//			SearchPath: &databaseSearchPath,
	//		},
	//		GeneralOptions: &options.General{
	//			UpdateCheck: &generalUpdateCheck,
	//		},
	//	},
	//},
	//"options_only": { // fixed
	//	steampipeDir: "testdata/hcl_helpers/options_only",
	//	expected: &SteampipeConfig{
	//		Connections: map[string]*modconfig.Connection{},
	//		DefaultConnectionOptions: &options.Connection{
	//			Cache:    &trueVal,
	//			CacheTTL: &ttlVal,
	//		},
	//		DatabaseOptions: &options.Database{
	//			Port:       &databasePort,
	//			Listen:     &databaseListen,
	//			SearchPath: &databaseSearchPath,
	//		},
	//		GeneralOptions: &options.General{
	//			UpdateCheck: &generalUpdateCheck,
	//		},
	//	},
	//},
	//"options_duplicate_block": {
	//	steampipeDir: "testdata/hcl_helpers/options_duplicate_block",
	//	expected:     "ERROR",
	//},
}

func TestLoadConfig(t *testing.T) {
	// TODO KAI update these
	t.Skip("needs updating")
	// get the current working directory of the test(used to build the DeclRange.Filename property)
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current working directory")
	}

	for name, test := range testCasesLoadConfig {
		// default workspoace to empty dir
		workspaceDir := test.workspaceDir
		if workspaceDir == "" {
			workspaceDir = "testdata/load_config_test/empty"
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
		config, errorsAndWarnings := loadSteampipeConfig(workspaceDir, "")
		if errorsAndWarnings.GetError() != nil {
			if test.expected != "ERROR" {
				t.Errorf("Test: '%s'' FAILED with unexpected error: %v", name, errorsAndWarnings.GetError())
			}
			continue
		}

		if test.expected == "ERROR" {
			t.Errorf("Test: '%s'' FAILED - expected error", name)
			continue
		}

		expectedConfig := test.expected.(*SteampipeConfig)
		for _, c := range expectedConfig.Connections {
			c.DeclRange.Filename = strings.Replace(c.DeclRange.Filename, "$$test_pwd$$", pwd, 1)
		}
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

	if !maps.EqualFunc(left.Connections, right.Connections,
		func(c1, c2 *modconfig.Connection) bool { return c1.Equals(c2) }) {
		return false
	}
	if !reflect.DeepEqual(left.DefaultConnectionOptions, right.DefaultConnectionOptions) {
		return false
	}
	if !reflect.DeepEqual(left.DatabaseOptions, right.DatabaseOptions) {
		return false
	}
	if !reflect.DeepEqual(left.TerminalOptions, right.TerminalOptions) {
		return false
	}
	if !reflect.DeepEqual(left.GeneralOptions, right.GeneralOptions) {
		return false
	}
	return true
}
