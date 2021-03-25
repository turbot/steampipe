package workspace

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/turbot/steampipe/mod"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type loadWorkspaceTest struct {
	source   string
	expected interface{}
}

var m3alias = "m3"

var testCasesLoadWorkspace = map[string]loadWorkspaceTest{
	"single mod": {
		source: "test_data/w_1",
		expected: mod.ModMap{
			"github.com/turbot/m1": &modconfig.Mod{
				Name:        "m1",
				Title:       "M1",
				Description: "THIS IS M1",
				Version:     "0.0.0",
				ModDepends: []*modconfig.ModVersion{
					{"github.com/turbot/m3", "0.0.0", &m3alias},
				},
				Queries: []*modconfig.Query{
					{
						"q1", "Q1", "THIS IS QUERY 1", "select 1",
					},
				},
			},
			"github.com/turbot/m2": &modconfig.Mod{
				Name:        "m2",
				Title:       "M2",
				Description: "THIS IS M2",
				Version:     "0.0.0",
				Queries: []*modconfig.Query{
					{
						"q1", "Q1", "THIS IS QUERY 1", "select 2",
					},
				},
			},
		},
	},
}

func TestLoadMod(t *testing.T) {
	for name, test := range testCasesLoadWorkspace {
		workspacePath, err := filepath.Abs(test.source)
		if err != nil {
			t.Errorf("failed to build absolute config filepath from %s", test.source)
		}

		workspace, err := Load(workspacePath)

		if err != nil && test.expected != "ERROR" {
			t.Errorf("TestLoadMod failed with unexpected error: %v", err)
		}

		expectedStr := test.expected.(mod.ModMap).String()
		actualString := workspace.Mods.String()
		if expectedStr != actualString {
			fmt.Printf("")
			t.Errorf("Test: '%s'' FAILED : expected:\n\n%s\n\ngot:\n\n%s", name, expectedStr, actualString)
		}
	}
}
