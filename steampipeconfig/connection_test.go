package steampipeconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

type getConnectionsToUpdateTest struct {
	// hcl connection config(s)
	required []string
	// current connection state
	current  ConnectionDataMap
	expected interface{}
}

var connectionTest1Checksum = getTestFileCheckSum("test_data/connections_to_update/plugins_src/hub.steampipe.io/plugins/turbot/connection-test-1@latest/connection-test-1.plugin")
var connectionTest2Checksum = getTestFileCheckSum("test_data/connections_to_update/plugins_src/hub.steampipe.io/plugins/turbot/connection-test-2@latest/connection-test-2.plugin")

var testCasesGetConnectionsToUpdate = map[string]getConnectionsToUpdateTest{
	"no changes": {
		required: []string{
			`connection "a" {
  plugin = "connection-test-1"
}
`},
		current: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionDataMap{}, Delete: ConnectionDataMap{}, RequiredConnections: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
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
		current: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "b"},
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionDataMap{}, Delete: ConnectionDataMap{}, RequiredConnections: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "b"},
			},
		}},
	},
	"no changes multiple in same file": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}
	
	connection "b" {
	 plugin = "connection-test-2"
	}
	`},
		current: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:   connectionTest2Checksum,
				Connection: &modconfig.Connection{Name: "b"},
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionDataMap{}, Delete: ConnectionDataMap{}, RequiredConnections: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:   connectionTest2Checksum,
				Connection: &modconfig.Connection{Name: "b"},
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
		current: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "b"},
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionDataMap{}, Delete: ConnectionDataMap{}, RequiredConnections: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "b"},
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
		current: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:   connectionTest2Checksum,
				Connection: &modconfig.Connection{Name: "b"},
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionDataMap{}, Delete: ConnectionDataMap{}, RequiredConnections: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:   connectionTest2Checksum,
				Connection: &modconfig.Connection{Name: "b"},
			},
		}},
	},
	"update": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}
	`},
		current: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   "xxxxxx",
				Connection: &modconfig.Connection{Name: "a"},
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
		}, Delete: ConnectionDataMap{}, RequiredConnections: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
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
		current: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   "xxxxxx",
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   "xxxxxx",
				Connection: &modconfig.Connection{Name: "a"},
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "b"},
			},
		},
			Delete: ConnectionDataMap{},
			RequiredConnections: ConnectionDataMap{
				"a": {
					Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:   connectionTest1Checksum,
					Connection: &modconfig.Connection{Name: "a"},
				},
				"b": {
					Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:   connectionTest1Checksum,
					Connection: &modconfig.Connection{Name: "b"},
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
		current: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   "xxxxxx",
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:   "xxxxxx",
				Connection: &modconfig.Connection{Name: "a"},
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:   connectionTest2Checksum,
				Connection: &modconfig.Connection{Name: "b"},
			},
		}, Delete: ConnectionDataMap{},
			RequiredConnections: ConnectionDataMap{
				"a": {
					Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:   connectionTest1Checksum,
					Connection: &modconfig.Connection{Name: "a"},
				},
				"b": {
					Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
					CheckSum:   connectionTest2Checksum,
					Connection: &modconfig.Connection{Name: "b"},
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
		current: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   "xxxxxx",
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   "xxxxxx",
				Connection: &modconfig.Connection{Name: "a"},
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   connectionTest1Checksum,
				Connection: &modconfig.Connection{Name: "b"},
			},
		},
			Delete: ConnectionDataMap{},
			RequiredConnections: ConnectionDataMap{
				"a": {
					Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:   connectionTest1Checksum,
					Connection: &modconfig.Connection{Name: "a"},
				},
				"b": {
					Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:   connectionTest1Checksum,
					Connection: &modconfig.Connection{Name: "b"},
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
		current: ConnectionDataMap{
			"a": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:   "xxxxxx",
				Connection: &modconfig.Connection{Name: "a"},
			},
			"b": {
				Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:   "xxxxxx",
				Connection: &modconfig.Connection{Name: "a"},
			},
		},
		expected: &ConnectionUpdates{
			Update: ConnectionDataMap{
				"a": {
					Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:   connectionTest1Checksum,
					Connection: &modconfig.Connection{Name: "a"},
				},
				"b": {
					Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
					CheckSum:   connectionTest2Checksum,
					Connection: &modconfig.Connection{Name: "b"},
				},
			},
			Delete: ConnectionDataMap{},
			RequiredConnections: ConnectionDataMap{
				"a": {
					Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:   connectionTest1Checksum,
					Connection: &modconfig.Connection{Name: "a"},
				},
				"b": {
					Plugin:     "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
					CheckSum:   connectionTest2Checksum,
					Connection: &modconfig.Connection{Name: "b"},
				},
			}},
	},

	"not installed": {
		required: []string{
			`connection "a" {
	 plugin = "not-installed"
	}
	`},
		current:  ConnectionDataMap{},
		expected: "SHOULD NOT BE ERROR?",
	},
}

func TestGetConnectionsToUpdate(t *testing.T) {
	// set steampipe dir
	os.Chdir("./test_data/connections_to_update")
	wd, _ := os.Getwd()
	constants.SteampipeDir = wd

	for name, test := range testCasesGetConnectionsToUpdate {
		// setup connection config
		setup(test)

		config, err := LoadSteampipeConfig(wd, "")
		if config == nil {
			t.Fatalf("Could not load config")
		}
		requiredConnections := config.Connections
		// all tests assume connections a, b
		res, err := GetConnectionsToUpdate([]string{"a", "b"}, requiredConnections, nil, nil)

		if err != nil && test.expected != "ERROR" {
			continue
			t.Fatalf("GetConnectionsToUpdate failed with unexpected error: %v", err)
		}

		expectedUpdates := test.expected.(*ConnectionUpdates)
		if !res.RequiredConnections.Equals(expectedUpdates.RequiredConnections) ||
			!res.Update.Equals(expectedUpdates.Update) ||
			!res.Delete.Equals(expectedUpdates.Delete) {
			t.Errorf(`Test: '%s'' FAILED`, name)

		}

		fmt.Printf("\n\n'Test: %s' PASSED\n\n", name)
		resetConfig(test)
	}
}

type connectionDataEqual struct {
	data1       *ConnectionData
	data2       *ConnectionData
	expectation bool
}

var data1 ConnectionData = ConnectionData{
	Plugin:     "plugin",
	CheckSum:   "checksum",
	Connection: &modconfig.Connection{Name: "a"},
}
var data1_duplicate ConnectionData = ConnectionData{
	Plugin:     "plugin",
	CheckSum:   "checksum",
	Connection: &modconfig.Connection{Name: "a"},
}
var data2 ConnectionData = ConnectionData{
	Plugin:     "plugin2",
	CheckSum:   "checksum2",
	Connection: &modconfig.Connection{Name: "b"},
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

	os.RemoveAll(constants.PluginDir())
	os.RemoveAll(constants.ConfigDir())
	os.RemoveAll(constants.InternalDir())

	os.MkdirAll(constants.PluginDir(), os.ModePerm)
	os.MkdirAll(constants.ConfigDir(), os.ModePerm)
	os.MkdirAll(constants.InternalDir(), os.ModePerm)

	for _, plugin := range test.current {
		copyPlugin(plugin.Plugin)
	}
	setupTestConfig(test)
}

func setupTestConfig(test getConnectionsToUpdateTest) {
	for i, config := range test.required {
		ioutil.WriteFile(connectionConfigPath(i), []byte(config), 0644)
	}
	os.MkdirAll(constants.InternalDir(), os.ModePerm)
	writeJson(test.current, constants.ConnectionStatePath())
}

func resetConfig(test getConnectionsToUpdateTest) {
	connectionStatePath := constants.ConnectionStatePath()

	os.Remove(connectionStatePath)
	for i, _ := range test.required {
		os.Remove(connectionConfigPath(i))
	}
}

func connectionConfigPath(i int) string {
	fileName := fmt.Sprintf("test%d%s", i, constants.ConfigExtension)
	path := filepath.Join(constants.ConfigDir(), fileName)
	return path
}

func copyPlugin(plugin string) {
	source, err := filepath.Abs(filepath.Join("plugins_src", plugin))
	if err != nil {
		log.Fatal(err)
	}
	dest, err := filepath.Abs(filepath.Join(constants.PluginDir(), plugin))
	if err != nil {
		log.Fatal(err)
	}

	err = copy.Copy(source, dest)
	if err != nil {
		log.Fatal(err)
	}
}

func getTestFileCheckSum(file string) string {
	p, err := filepath.Abs(file)
	if err != nil {
		log.Fatal(err)
	}
	sha, err := utils.FileHash(p)
	if err != nil {
		log.Fatal(err)
	}
	return sha
}
