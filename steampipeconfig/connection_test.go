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
	"github.com/turbot/steampipe/utils"
)

type getConnectionsToUpdateTest struct {
	// hcl connection config(s)
	required []string
	// current connection state
	current  ConnectionMap
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
		current: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}, RequiredConnections: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
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
		current: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "b",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}, RequiredConnections: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "b",
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
		current: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:       connectionTest2Checksum,
				ConnectionName: "b",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}, RequiredConnections: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:       connectionTest2Checksum,
				ConnectionName: "b",
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
		current: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "b",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}, RequiredConnections: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "b",
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
		current: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:       connectionTest2Checksum,
				ConnectionName: "b",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}, RequiredConnections: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:       connectionTest2Checksum,
				ConnectionName: "b",
			},
		}},
	},
	"update": {
		required: []string{
			`connection "a" {
	 plugin = "connection-test-1"
	}
	`},
		current: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       "xxxxxx",
				ConnectionName: "a",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
		}, Delete: ConnectionMap{}, RequiredConnections: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
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
		current: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       "xxxxxx",
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       "xxxxxx",
				ConnectionName: "a",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "b",
			},
		},
			Delete: ConnectionMap{},
			RequiredConnections: ConnectionMap{
				"a": {
					Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:       connectionTest1Checksum,
					ConnectionName: "a",
				},
				"b": {
					Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:       connectionTest1Checksum,
					ConnectionName: "b",
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
		current: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       "xxxxxx",
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:       "xxxxxx",
				ConnectionName: "a",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:       connectionTest2Checksum,
				ConnectionName: "b",
			},
		}, Delete: ConnectionMap{},
			RequiredConnections: ConnectionMap{
				"a": {
					Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:       connectionTest1Checksum,
					ConnectionName: "a",
				},
				"b": {
					Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
					CheckSum:       connectionTest2Checksum,
					ConnectionName: "b",
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
		current: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       "xxxxxx",
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       "xxxxxx",
				ConnectionName: "a",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       connectionTest1Checksum,
				ConnectionName: "b",
			},
		},
			Delete: ConnectionMap{},
			RequiredConnections: ConnectionMap{
				"a": {
					Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:       connectionTest1Checksum,
					ConnectionName: "a",
				},
				"b": {
					Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:       connectionTest1Checksum,
					ConnectionName: "b",
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
		current: ConnectionMap{
			"a": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
				CheckSum:       "xxxxxx",
				ConnectionName: "a",
			},
			"b": {
				Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
				CheckSum:       "xxxxxx",
				ConnectionName: "a",
			},
		},
		expected: &ConnectionUpdates{
			Update: ConnectionMap{
				"a": {
					Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:       connectionTest1Checksum,
					ConnectionName: "a",
				},
				"b": {
					Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
					CheckSum:       connectionTest2Checksum,
					ConnectionName: "b",
				},
			},
			Delete: ConnectionMap{},
			RequiredConnections: ConnectionMap{
				"a": {
					Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-1@latest",
					CheckSum:       connectionTest1Checksum,
					ConnectionName: "a",
				},
				"b": {
					Plugin:         "hub.steampipe.io/plugins/turbot/connection-test-2@latest",
					CheckSum:       connectionTest2Checksum,
					ConnectionName: "b",
				},
			}},
	},

	"not installed": {
		required: []string{
			`connection "a" {
	 plugin = "not-installed"
	}
	`},
		current:  ConnectionMap{},
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

		config, err := LoadSteampipeConfig(wd)
		if config == nil {
			t.Fatalf("Could not load config")
		}
		requiredConnections := config.Connections
		// all tests assume connections a, b
		res, err := GetConnectionsToUpdate([]string{"a", "b"}, requiredConnections)

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
