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

var alias = "_m2"

var testCasesLoadMod = map[string]loadModTest{
	"single_mod_no_query": {
		source: "test_data/single_mod_no_query",
		expected: &modconfig.Mod{
			Name:        "m1",
			Title:       "M1",
			Description: "THIS IS M1",
			ModDepends: []*modconfig.ModVersion{
				{"github.com/turbot/m2", "0.0.0", &alias},
			},
		},
	},
	"single_mod_one_query": {
		source: "test_data/single_mod_one_query",
		expected: &modconfig.Mod{
			Name:        "m1",
			Title:       "M1",
			Description: "THIS IS M1",
			ModDepends: []*modconfig.ModVersion{
				{"github.com/turbot/m2", "0.0.0", &alias},
			},
			Queries: []*modconfig.Query{
				{
					"q1", "Q1", "THIS IS QUERY 1", "select 1",
				},
			},
		},
	},
	"single_mod_two_queries_diff_files": {
		source: "test_data/single_mod_two_queries_diff_files",
		expected: &modconfig.Mod{
			Name:        "m1",
			Title:       "M1",
			Description: "THIS IS M1",
			ModDepends: []*modconfig.ModVersion{
				{"github.com/turbot/m2", "0.0.0", &alias},
			},
			Queries: []*modconfig.Query{
				{
					"q1", "Q1", "THIS IS QUERY 1", "select 1",
				},
				{
					"q2", "Q2", "THIS IS QUERY 2", "select 2",
				},
			},
		},
	},
	"single_mod_two_queries_same_file": {
		source: "test_data/single_mod_two_queries_same_file",
		expected: &modconfig.Mod{
			Name:        "m1",
			Title:       "M1",
			Description: "THIS IS M1",
			ModDepends: []*modconfig.ModVersion{
				{"github.com/turbot/m2", "0.0.0", &alias},
			},
			Queries: []*modconfig.Query{
				{
					"q1", "Q1", "THIS IS QUERY 1", "select 1",
				},
				{
					"q2", "Q2", "THIS IS QUERY 2", "select 2",
				},
			},
		},
	},
	"single_mod_duplicate_query": {
		source:   "test_data/single_mod_duplicate_query",
		expected: "ERROR",
	},
	"no_mod": {
		source:   "test_data/no_mod",
		expected: "ERROR",
	},
	"two_mods": {
		source:   "test_data/two_mods",
		expected: "ERROR",
	},
}

func TestLoadMod(t *testing.T) {
	for name, test := range testCasesLoadMod {
		modPath, err := filepath.Abs(test.source)
		if err != nil {
			t.Errorf("failed to build absolute config filepath from %s", test.source)
		}

		mod, err := LoadMod(modPath)

		if err != nil {
			if test.expected != "ERROR" {
				t.Errorf("Test: '%s'' FAILED with unexpected error: %v", name, err)
			}
			return
		}

		if test.expected == "ERROR" {
			t.Errorf("Test: '%s'' FAILED - expected error", name)
		}

		expectedStr := test.expected.(*modconfig.Mod).String()
		actualString := mod.String()

		if expectedStr != actualString {
			fmt.Printf("")
			t.Errorf("Test: '%s'' FAILED : expected:\n\n%s\n\ngot:\n\n%s", name, expectedStr, actualString)
		}
	}
}
