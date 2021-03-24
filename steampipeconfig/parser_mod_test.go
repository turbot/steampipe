package steampipeconfig

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type loadModTest struct {
	source   string
	expected interface{}
}

var testCasesLoadMod = map[string]loadModTest{
	"single mod": {
		source: "test_data/single_mod",
		expected: &modconfig.Mod{
			Name:        "m1",
			Title:       "M1",
			Description: "THIS IS M1",
			Version:     "0.0.0",
			ModDepends: []*modconfig.ModVersion{
				{"github.com/turbot/m2", "0.0.0"},
			},
			Queries: []*modconfig.Query{
				{
					"q1", "Q1", "THIS IS QUERY 1", "select 1",
				},
			},
		},
	},
}

func TestLoadMod(t *testing.T) {
	for name, test := range testCasesLoadMod {
		modPath, err := filepath.Abs(test.source)
		if err != nil {
			t.Errorf("failed to build absolute config filepath from %s", test.source)
		}

		mod, err := loadMod(modPath)

		if err != nil && test.expected != "ERROR" {
			t.Errorf("TestLoadMod failed with unexpected error: %v", err)
		}

		expectedStr := test.expected.(*modconfig.Mod).String()
		actualString := mod.String()
		if expectedStr != actualString {
			fmt.Printf("")
			t.Errorf("Test: '%s'' FAILED : expected:\n\n%s\n\ngot:\n\n%s", name, expectedStr, actualString)
		}
	}
}
