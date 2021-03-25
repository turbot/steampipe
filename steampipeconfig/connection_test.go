package steampipeconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
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

var testCasesGetConnectionsToUpdate = map[string]getConnectionsToUpdateTest{
	"no changes": {
		required: []string{
			`connection "a" {
  plugin = "test_data/connection-test-1"
}
`},
		current: ConnectionMap{
			"a": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: getTestFileCheckSum("test_data/connection-test-1"),
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}},
	},
	"no changes multiple in same file same plugin": {
		required: []string{
			`connection "a" {
  plugin = "test_data/connection-test-1"
}

connection "b" {
  plugin = "test_data/connection-test-1"
}
`},
		current: ConnectionMap{
			"a": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: getTestFileCheckSum("test_data/connection-test-1"),
			},
			"b": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: getTestFileCheckSum("test_data/connection-test-1"),
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}},
	},
	"no changes multiple in same file": {
		required: []string{
			`connection "a" {
  plugin = "test_data/connection-test-1"
}

connection "b" {
  plugin = "test_data/connection-test-2"
}
`},
		current: ConnectionMap{
			"a": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: getTestFileCheckSum("test_data/connection-test-1"),
			},
			"b": {
				Plugin:   "test_data/connection-test-2",
				CheckSum: getTestFileCheckSum("test_data/connection-test-2"),
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}},
	},
	"no changes multiple in different files same plugin": {
		required: []string{
			`connection "a" {
  plugin = "test_data/connection-test-1"
}`,
			`connection "b" {
  plugin = "test_data/connection-test-1"
}
`},
		current: ConnectionMap{
			"a": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: getTestFileCheckSum("test_data/connection-test-1"),
			},
			"b": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: getTestFileCheckSum("test_data/connection-test-1"),
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}},
	},
	"no changes multiple in different files": {
		required: []string{
			`connection "a" {
  plugin = "test_data/connection-test-1"
}`,
			`connection "b" {
  plugin = "test_data/connection-test-2"
}
`},
		current: ConnectionMap{
			"a": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: getTestFileCheckSum("test_data/connection-test-1"),
			},
			"b": {
				Plugin:   "test_data/connection-test-2",
				CheckSum: getTestFileCheckSum("test_data/connection-test-2"),
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}},
	},
	"update": {
		required: []string{
			`connection "a" {
  plugin = "test_data/connection-test-1"
}
`},
		current: ConnectionMap{
			"a": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: "xxxxxx",
			},
		},
		expected: &ConnectionUpdates{

			Update: ConnectionMap{
				"a": {
					Plugin:   "test_data/connection-test-1",
					CheckSum: getTestFileCheckSum("test_data/connection-test-1"),
				},
			},
			Delete: ConnectionMap{},
		},
	},

	"update multiple in same file same plugin": {
		required: []string{
			`connection "a" {
  plugin = "test_data/connection-test-1"
}

connection "b" {
  plugin = "test_data/connection-test-1"
}
`},
		current: ConnectionMap{
			"a": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: "xxxxxx",
			},
			"b": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: "xxxxxx",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}},
	},
	"update multiple in same file": {
		required: []string{
			`connection "a" {
  plugin = "test_data/connection-test-1"
}

connection "b" {
  plugin = "test_data/connection-test-2"
}
`},
		current: ConnectionMap{
			"a": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: "xxxxxx",
			},
			"b": {
				Plugin:   "test_data/connection-test-2",
				CheckSum: "xxxxxx",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}},
	},
	"update multiple in different files same plugin": {
		required: []string{
			`connection "a" {
  plugin = "test_data/connection-test-1"
}`,
			`connection "b" {
  plugin = "test_data/connection-test-1"
}
`},
		current: ConnectionMap{
			"a": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: "xxxxxx",
			},
			"b": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: "xxxxxx",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}},
	},
	"update multiple in different files": {
		required: []string{
			`connection "a" {
  plugin = "test_data/connection-test-1"
}`,
			`connection "b" {
  plugin = "test_data/connection-test-2"
}
`},
		current: ConnectionMap{
			"a": {
				Plugin:   "test_data/connection-test-1",
				CheckSum: "xxxxxx",
			},
			"b": {
				Plugin:   "test_data/connection-test-2",
				CheckSum: "xxxxxx",
			},
		},
		expected: &ConnectionUpdates{Update: ConnectionMap{}, Delete: ConnectionMap{}},
	},

	"not installed": {
		required: []string{
			`connection "a" {
  plugin = "test_data/not-installed"
}
`},
		current: ConnectionMap{},
		expected: &ConnectionUpdates{
			MissingPlugins: []string{"test_data/not-installed"},
			Update:         ConnectionMap{},
			Delete:         ConnectionMap{},
		},
	},
}

func TestGetConnectionsToUpdate(t *testing.T) {
	for name, test := range testCasesGetConnectionsToUpdate {
		setup(test)

		res, err := GetConnectionsToUpdate(nil, nil)

		if err != nil && test.expected != "ERROR" {
			t.Errorf("GetConnectionsToUpdate failed with unexpected error: %v", err)
		}

		if !reflect.DeepEqual(res, test.expected) {
			t.Errorf(`Test: '%s'' FAILED : expected %v, got %v`, name, test.expected, res)
		}
		resetConfig(test)
	}
}

func setup(test getConnectionsToUpdateTest) {
	clearPluginFolder()

	for _, plugin := range test.current {
		copyPlugin(plugin.Plugin)
	}
	setupTestConfig(test)
}

func clearPluginFolder() {
	// all plugins are put into the test_data plugin directory
	targetFolder, err := filepath.Abs(filepath.Join(constants.PluginDir(), "test_data"))
	if err != nil {
		log.Fatal(err)
	}
	os.RemoveAll(targetFolder)
	os.MkdirAll(targetFolder, 0777)
}

func setupTestConfig(test getConnectionsToUpdateTest) {
	// move real config
	connectionStatePath := constants.ConnectionStatePath()

	for i, config := range test.required {
		ioutil.WriteFile(connectionConfigPath(i), []byte(config), 0644)
	}
	os.Rename(connectionStatePath, connectionStatePath+"___")
	writeJson(test.current, constants.ConnectionStatePath())
}

func resetConfig(test getConnectionsToUpdateTest) {
	connectionStatePath := constants.ConnectionStatePath()

	os.Remove(connectionStatePath)
	for i, _ := range test.required {
		os.Remove(connectionConfigPath(i))
	}

	os.Rename(connectionStatePath+"___", connectionStatePath)
}

func connectionConfigPath(i int) string {
	fileName := fmt.Sprintf("test%d%s", i, constants.ConfigExtension)
	path := filepath.Join(constants.ConfigDir(), fileName)
	return path
}

func copyPlugin(plugin string) {

	source, err := filepath.Abs(plugin)
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
	path := filepath.Join(file, filepath.Base(file)+constants.PluginExtension)
	p, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	sha, err := utils.FileHash(p)
	if err != nil {
		log.Fatal(err)
	}
	return sha
}

func TestGetPluginCheckSum(t *testing.T) {
	sha := getTestFileCheckSum("connection-test-1")
	fmt.Println(sha)
}
