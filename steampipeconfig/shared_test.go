package steampipeconfig

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/pluginmanager"
)

type findPluginFolderTest struct {
	schema   string
	expected string
}

var testCasesFindPluginFolderTest map[string]findPluginFolderTest

func init() {
	log.Printf("[TRACE] BEginning of init")
	filepaths.SteampipeDir, _ = helpers.Tildefy("~/.steampipe")

	testCasesFindPluginFolderTest = map[string]findPluginFolderTest{
		"truncated 1": {
			"hub.steampipe.io/plugins/test/test@sha256-a5ec85d93329-32c3ed1c",
			filepath.Join(filepaths.EnsurePluginDir(), "hub.steampipe.io/plugins/test/test@sha256-a5ec85d9332910f42a2a9dd44d646eba95f77a0236289a1a14a14abbbdea7a42"),
		},
		// "truncated 2 - 2 folders with same prefix": {
		// 	"hub.steampipe.io/plugins/test/test@sha256-5f77a0236289-94a0eea6",
		// 	filepath.Join(filepaths.EnsurePluginDir(), "hub.steampipe.io/plugins/test/test@sha256-5f77a0236289a1a14a14abbbdea7a42a5ec85d9332910f42a2a9dd44d646eba9"),
		// },
		// "no truncation needed": {
		// 	"hub.steampipe.io/plugins/test/test@latest",
		// 	filepath.Join(filepaths.EnsurePluginDir(), "hub.steampipe.io/plugins/test/test@latest"),
		// },
	}
	log.Printf("[TRACE] End of init, SteampipeDir: %s", filepaths.SteampipeDir)
}

func TestFindPluginFolderTest(t *testing.T) {
	log.Printf("[TRACE] BEginning of test, SteampipeDir: %s", filepaths.SteampipeDir)

	directories := []string{
		"hub.steampipe.io/plugins/test/test@sha256-a5ec85d9332910f42a2a9dd44d646eba95f77a0236289a1a14a14abbbdea7a42",
		"hub.steampipe.io/plugins/test/test@sha256-5f77a0236289a1a14a14abbbdea7a42a5ec85d9332910f42a2a9dd44d646eb00",
		"hub.steampipe.io/plugins/test/test@sha256-5f77a0236289a1a14a14abbbdea7a42a5ec85d9332910f42a2a9dd44d646eba9",
		"hub.steampipe.io/plugins/test/test@latest",
	}
	log.Printf("[TRACE] About to set up folders, SteampipeDir: %s", filepaths.SteampipeDir)

	setupFindPluginFolderTest(directories)
	log.Printf("[TRACE] After setting up folders, SteampipeDir: %s", filepaths.SteampipeDir)
	for name, test := range testCasesFindPluginFolderTest {

		path, err := pluginmanager.FindPluginFolder(test.schema)
		log.Printf("[TRACE] path: %s\n", path)
		if err != nil {
			if test.expected != "ERROR" {
				t.Errorf(`Test: '%s'' FAILED : unexpected error %v`, name, err)
			}
			continue
		}

		if path != test.expected {
			t.Errorf(`Test: '%s'' FAILED : expected %s, got %s`, name, test.expected, path)
		}
	}
	cleanupFindPluginFolderTest(directories)

}

func setupFindPluginFolderTest(directories []string) {
	for _, dir := range directories {
		pluginFolder := filepath.Join(filepaths.EnsurePluginDir(), dir)
		log.Printf("[TRACE] Adding pluginFolder: %s, steampipeDir: %s\n", pluginFolder, filepaths.SteampipeDir)
		if err := os.MkdirAll(pluginFolder, 0755); err != nil {
			panic(err)
		}
	}
}

func cleanupFindPluginFolderTest(directories []string) {
	pluginFolder := filepath.Join(filepaths.EnsurePluginDir(), "hub.steampipe.io/plugins/test")
	os.RemoveAll(pluginFolder)
}
