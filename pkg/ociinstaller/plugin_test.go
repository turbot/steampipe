package ociinstaller

import (
	"bytes"
	"fmt"
	"github.com/turbot/pipe-fittings/ociinstaller"
	"github.com/turbot/steampipe/pkg/filepaths"
	"os"
	"path/filepath"
	"testing"
)

type transformTest struct {
	ref                                  *ociinstaller.ImageRef
	pluginLineContent                    []byte
	expectedTransformedPluginLineContent []byte
}

var transformTests = map[string]transformTest{
	"empty": {
		ref:                                  NewSteampipeImageRef("chaos"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos"`),
	},
	"latest": {
		ref:                                  NewSteampipeImageRef("chaos@latest"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos"`),
	},
	"0": {
		ref:                                  NewSteampipeImageRef("chaos@0"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos@0"`),
	},
	"0.2": {
		ref:                                  NewSteampipeImageRef("chaos@0.2"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos@0.2"`),
	},
	"0.2.0": {
		ref:                                  NewSteampipeImageRef("chaos@0.2.0"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos@0.2.0"`),
	},
	"^0.2": {
		ref:                                  NewSteampipeImageRef("chaos@^0.2"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos@^0.2"`),
	},
	">=0.2": {
		ref:                                  NewSteampipeImageRef("chaos@>=0.2"),
		pluginLineContent:                    []byte(`plugin = "chaos"`),
		expectedTransformedPluginLineContent: []byte(`plugin = "chaos@>=0.2"`),
	},
}

func TestAddPluginName(t *testing.T) {
	for name, test := range transformTests {
		sourcebytes := test.pluginLineContent
		expectedBytes := test.expectedTransformedPluginLineContent
		_, _, constraint := test.ref.GetOrgNameAndConstraint(constants.SteampipeHubOCIBase)
		transformed := bytes.TrimSpace(addPluginConstraintToConfig(sourcebytes, constraint))

		if !bytes.Equal(transformed, expectedBytes) {
			t.Fatalf("%s failed - expected(%s) - got(%s)", name, test.expectedTransformedPluginLineContent, transformed)
		}
	}
}

func TestConstraintBasedFilePathsReadWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	filepaths.SteampipeDir = tmpDir

	cases := make(map[string][]string)
	fileContent := "test string"

	cases["basic_checks"] = []string{
		"latest",
		"1.2.3",
		"1.0",
		"1",
	}
	cases["operators"] = []string{
		"!=1.2.3",
		">1.2.0",
		">1.2",
		">1",
		">=1.2.0",
		"<1.2.0",
		"<=1.2.0",
	}
	cases["hyphen_range"] = []string{
		"1.1-1.2.3",
		"1.2.1-1.2.3",
	}
	cases["wild_cards"] = []string{
		"*",
		"1.x",
		"1.*",
		"1.1.x",
		"1.1.*",
		">=1.2.x",
		"<=1.1.x",
	}
	cases["tilde_range"] = []string{
		"~1",
		"~1.1",
		"~1.x",
		"~1.1.1",
		"~1.1.x",
	}
	cases["caret_range"] = []string{
		"^1",
		"^1.1",
		"^1.x",
		"^1.1.1",
		"^1.1.*",
	}

	for category, testCases := range cases {
		for _, testCase := range testCases {
			constraintedDir := filepaths.EnsurePluginInstallDir(fmt.Sprintf("constraint-test:%s", testCase))
			filePath := filepath.Join(constraintedDir, "test.txt")

			// Write Test
			err := os.WriteFile(filePath, []byte(fileContent), 0644)
			if err != nil {
				t.Fatalf("Write failed for constraint %s %s", category, testCase)
			}
			// Read Test
			b, err := os.ReadFile(filePath)
			if err != nil || string(b) != fileContent {
				t.Fatalf("Read failed for constraint %s %s", category, testCase)
			}
			// tidy up
			if err := os.RemoveAll(constraintedDir); err != nil {
				t.Logf("Failed to remove test folder and contents: %s", constraintedDir)
			}
		}
	}
}
