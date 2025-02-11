package steampipeconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/otiai10/copy"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/pkg/constants"
)

type getConnectionsToUpdateTest struct {
	// hcl connection config(s)
	required []string
	// current connection state
	current  ConnectionStateMap
	expected interface{}
}

var connectionTest1ModTime = getTestFileModTime("test_data/connections_to_update/plugins_src/hub.steampipe.io/plugins/turbot/connection-test-1@latest/connection-test-1.plugin")
var connectionTest2ModTime = getTestFileModTime("test_data/connections_to_update/plugins_src/hub.steampipe.io/plugins/turbot/connection-test-2@latest/connection-test-2.plugin")

var testCasesGetConnectionsToUpdate = map[string]getConnectionsToUpdateTest{
	"no changes": {
		required: []string{
			`connection "a" {
  plugin = "connection-test-1"
}
`},
		current: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionStateMap{}, Delete: nil, FinalConnectionState: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
		}},
	},
	"no changes multiple in same file same plugin": {
		required: []string{
			`connection "a" {
  plugin = "connection-test-1"
}

connection "b" {
  plugin = "connection-test-1"
}
`},
		current: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionStateMap{}, Delete: nil, FinalConnectionState: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
		}},
	},
	"no changes multiple in same file": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}—
	
	connection "b" {
	 plugin = "connection-test-2"
	}
	`},
		current: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				PluginModTime: connectionTest2ModTime,
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionStateMap{}, Delete: nil, FinalConnectionState: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				PluginModTime: connectionTest2ModTime,
			},
		}},
	},
	"no changes multiple in different files same plugin": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}`,
			`connection "b" {
	 plugin = "connection-test-1"
	}
	`},
		current: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionStateMap{}, Delete: nil, FinalConnectionState: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
		}},
	},
	"no changes multiple in different files": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}`,
			`connection "b" {
	 plugin = "connection-test-2"
	}
	`},
		current: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				PluginModTime: connectionTest2ModTime,
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionStateMap{}, Delete: nil, FinalConnectionState: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				PluginModTime: connectionTest2ModTime,
			},
		}},
	},
	"update": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}
	`},
		current: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: time.Now(),
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
		}, Delete: nil, FinalConnectionState: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
		}},
	},

	"update multiple in same file same plugin": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}
	
	connection "b" {
	 plugin = "connection-test-1"
	}
	`},
		current: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: time.Now(),
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: time.Now(),
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
		},
			Delete: nil,
			FinalConnectionState: ConnectionStateMap{
				"a": {
					Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					PluginModTime: connectionTest1ModTime,
				},
				"b": {
					Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					PluginModTime: connectionTest1ModTime,
				},
			}},
	},
	"update multiple in same file": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}
	
	connection "b" {
	 plugin = "connection-test-2"
	}
	`},
		current: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: time.Now(),
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				PluginModTime: time.Now(),
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				PluginModTime: connectionTest2ModTime,
			},
		}, Delete: nil,
			FinalConnectionState: ConnectionStateMap{
				"a": {
					Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					PluginModTime: connectionTest1ModTime,
				},
				"b": {
					Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
					PluginModTime: connectionTest2ModTime,
				},
			}},
	},
	"update multiple in different files same plugin": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}`,
			`connection "b" {
	 plugin = "connection-test-1"
	}
	`},
		current: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: time.Now(),
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: time.Now(),
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: connectionTest1ModTime,
			},
		},
			Delete: nil,
			FinalConnectionState: ConnectionStateMap{
				"a": {
					Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					PluginModTime: connectionTest1ModTime,
				},
				"b": {
					Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					PluginModTime: connectionTest1ModTime,
				},
			}},
	},
	"update multiple in different files": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}`,
			`connection "b" {
	 plugin = "connection-test-2"
	}
	`},
		current: ConnectionStateMap{
			"a": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				PluginModTime: time.Now(),
			},
			"b": {
				Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				PluginModTime: time.Now(),
			},
		},
		expected: &ConnectionUpdates{
			Update: ConnectionStateMap{
				"a": {
					Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					PluginModTime: connectionTest1ModTime,
				},
				"b": {
					Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
					PluginModTime: connectionTest2ModTime,
				},
			},
			Delete: nil,
			FinalConnectionState: ConnectionStateMap{
				"a": {
					Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					PluginModTime: connectionTest1ModTime,
				},
				"b": {
					Plugin:        "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
					PluginModTime: connectionTest2ModTime,
				},
			}},
	},

	"not installed": {
		required: []string{
			`connection "a" {
	 plugin = "not-installed"
	}
	`},
		current:  ConnectionStateMap{},
		expected: "SHOULD NOT BE ERROR?",
	},
}

// This test is disabled since the code this one tests also starts up the plugin manager
// process. We need to find a lower denominator to test the functionalities that this one covers
//
// func TestGetConnectionsToUpdate(t *testing.T) {
// 	// set steampipe dir
// 	os.Chdir("./test_data/connections_to_update")
// 	wd, _ := os.Getwd()
// 	app_specific.InstallDir = wd

// 	for name, test := range testCasesGetConnectionsToUpdate {
// 		// setup connection config
// 		setup(test)
// 		defer func(t getConnectionsToUpdateTest) {
// 			teardown(t)
// 		}(test)

// 		config, err := LoadSteampipeConfig(wd, "")
// 		if err != nil {
// 			t.Fatalf("LoadSteampipeConfig failed with unexpected error: %v", err)
// 		}
// 		if config == nil {
// 			t.Fatalf("Could not load config")
// 		}
// 		GlobalConfig = config
// 		// all tests assume connections a, b
// 		updates, res := NewConnectionUpdates([]string{"a", "b"})

// 		if res.Error != nil && test.expected != "ERROR" {
// 			t.Fatalf("NewConnectionUpdates failed with unexpected error for \"%s\": %v", name, res.Error)
// 			continue
// 		}

// 		expectedUpdates := test.expected.(*ConnectionUpdates)
// 		if !updates.RequiredConnectionState.Equals(expectedUpdates.RequiredConnectionState) ||
// 			!updates.Update.Equals(expectedUpdates.Update) ||
// 			!updates.Delete.Equals(expectedUpdates.Delete) {
// 			t.Errorf(`Test: '%s'' FAILED`, name)

// 		}

// 		fmt.Printf("\n\n'Test: %s' PASSED\n\n", name)
// 	}
// }

type connectionDataEqual struct {
	data1       *ConnectionState
	data2       *ConnectionState
	expectation bool
}

var data1 = ConnectionState{
	Plugin:        "plugin",
	PluginModTime: time.Now().Round(time.Second),
}
var data1_duplicate = ConnectionState{
	Plugin:        "plugin",
	PluginModTime: time.Now().Round(time.Second),
}
var data2 = ConnectionState{
	Plugin:        "plugin2",
	PluginModTime: time.Now().Add(-1 * time.Hour),
}

var connectionDataEqualCases map[string]connectionDataEqual = map[string]connectionDataEqual{
	"expected_equal":     {data1: &data1, data2: &data1_duplicate, expectation: true},
	"not_expected_equal": {data1: &data1, data2: &data2, expectation: false},
}

func TestConnectionsUpdateEqual(t *testing.T) {
	for caseName, caseData := range connectionDataEqualCases {
		isEqual := caseData.data1.Equals(caseData.data2)
		if caseData.expectation != isEqual {
			t.Errorf(`Test: '%s' FAILED: expected: %v, actual: %v`, caseName, caseData.expectation, isEqual)
		}
	}
}

func setup(test getConnectionsToUpdateTest) {

	os.RemoveAll(filepaths.EnsurePluginDir())
	os.RemoveAll(filepaths.EnsureConfigDir())
	os.RemoveAll(filepaths.EnsureInternalDir())

	err := os.MkdirAll(filepaths.EnsurePluginDir(), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(filepaths.EnsureConfigDir(), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(filepaths.EnsureInternalDir(), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	for _, plugin := range test.current {
		copyPlugin(plugin.Plugin)
	}
	setupTestConfig(test)
}
func teardown(test getConnectionsToUpdateTest) {
	os.RemoveAll(filepaths.EnsurePluginDir())
	os.RemoveAll(filepaths.EnsureConfigDir())
	os.RemoveAll(filepaths.EnsureInternalDir())

	for _, plugin := range test.current {
		deletePlugin(plugin.Plugin)
	}
	resetConfig(test)
}

func setupTestConfig(test getConnectionsToUpdateTest) {
	for i, config := range test.required {
		if err := os.WriteFile(connectionConfigPath(i), []byte(config), 0644); err != nil {
			log.Fatal(err)
		}
	}
	err := os.MkdirAll(filepaths.EnsureInternalDir(), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	test.current.Save()
}

func resetConfig(test getConnectionsToUpdateTest) {
	connectionStatePath := filepaths.ConnectionStatePath()

	os.Remove(connectionStatePath)
	for i := range test.required {
		os.Remove(connectionConfigPath(i))
	}
}

func connectionConfigPath(i int) string {
	fileName := fmt.Sprintf("test%d%s", i, constants.ConfigExtension)
	path := filepath.Join(filepaths.EnsureConfigDir(), fileName)
	return path
}

func copyPlugin(plugin string) {
	source, err := filepath.Abs(filepath.Join("testdata", "connections_to_update", "plugins_src", plugin))

	if err != nil {
		log.Fatal(err)
	}
	dest, err := filepath.Abs(filepath.Join(filepaths.EnsurePluginDir(), plugin))
	if err != nil {
		log.Fatal(err)
	}

	err = copy.Copy(source, dest)
	if err != nil {
		log.Fatal(err)
	}
}
func deletePlugin(plugin string) {
	dest, err := filepath.Abs(filepath.Join(filepaths.EnsurePluginDir(), plugin))
	if err != nil {
		log.Fatal(err)
	}
	os.RemoveAll(dest)
}

func getTestFileModTime(file string) time.Time {
	modTime, _ := utils.FileModTime(file)
	return modTime
}
