package steampipeconfig

import (
	"os"
	"path/filepath"
	"testing"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
)

type findPluginFolderTest struct {
	schema   string
	expected string
}

var testCasesFindPluginFolderTest map[string]findPluginFolderTest

func setupTestData() {

	testCasesFindPluginFolderTest = map[string]findPluginFolderTest{
		"truncated 1": {
			"hub.steampipe.io/plugins/test/test@sha256-a5ec85d93329-32c3ed1c",
			filepath.Join(filepaths.EnsurePluginDir(), "hub.steampipe.io/plugins/test/test@sha256-a5ec85d9332910f42a2a9dd44d646eba95f77a0236289a1a14a14abbbdea7a42"),
		},
		"truncated 2 - 2 folders with same prefix": {
			"hub.steampipe.io/plugins/test/test@sha256-5f77a0236289-94a0eea6",
			filepath.Join(filepaths.EnsurePluginDir(), "hub.steampipe.io/plugins/test/test@sha256-5f77a0236289a1a14a14abbbdea7a42a5ec85d9332910f42a2a9dd44d646eba9"),
		},
		"no truncation needed": {
			"hub.steampipe.io/plugins/test/test@latest",
			filepath.Join(filepaths.EnsurePluginDir(), "hub.steampipe.io/plugins/test/test@latest"),
		},
	}
}

func TestFindPluginFolderTest(t *testing.T) {

	filepaths.SteampipeDir, _ = filehelpers.Tildefy("~/.steampipe")
	setupTestData()

	directories := []string{
		"hub.steampipe.io/plugins/test/test@sha256-a5ec85d9332910f42a2a9dd44d646eba95f77a0236289a1a14a14abbbdea7a42",
		"hub.steampipe.io/plugins/test/test@sha256-5f77a0236289a1a14a14abbbdea7a42a5ec85d9332910f42a2a9dd44d646eb00",
		"hub.steampipe.io/plugins/test/test@sha256-5f77a0236289a1a14a14abbbdea7a42a5ec85d9332910f42a2a9dd44d646eba9",
		"hub.steampipe.io/plugins/test/test@latest",
	}

	setupFindPluginFolderTest(directories)
	for name, test := range testCasesFindPluginFolderTest {
		path, err := filepaths.FindPluginFolder(test.schema)
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
		if err := os.MkdirAll(pluginFolder, 0755); err != nil {
			panic(err)
		}
	}
}

func cleanupFindPluginFolderTest(directories []string) {
	pluginFolder := filepath.Join(filepaths.EnsurePluginDir(), "hub.steampipe.io/plugins/test")
	os.RemoveAll(pluginFolder)
}
