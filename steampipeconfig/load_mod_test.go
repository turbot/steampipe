package steampipeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type loadModTest struct {
	source   string
	expected interface{}
}

var alias = "_m2"

// TODO these are really workspace loading tests - maybe movve there and have simpler mod loading tests here?
var loadWorkspaceOptions = &LoadModOptions{
	Exclude: []string{fmt.Sprintf("**/%s*", constants.WorkspaceDataDir)},
	Flags:   CreatePseudoResources | CreateDefaultMod,
}

var testCasesLoadMod = map[string]loadModTest{
	"no_mod_hcl_queries": {
		source: "test_data/mods/no_mod_hcl_queries",
		expected: &modconfig.Mod{
			Name: "local",
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
	"no_mod_nested_sql_files": {
		source: "test_data/mods/no_mod_nested_sql_files",
		expected: &modconfig.Mod{
			Name: "local",
			Queries: []*modconfig.Query{
				{
					Name: "queries_a_aa_q1", SQL: "select 1",
				},
				{
					Name: "queries_a_q1", SQL: "select 1",
				},
				{
					Name: "queries_b_bb_q2", SQL: "select 2",
				},
				{
					Name: "queries_b_q2", SQL: "select 2",
				},
			},
		},
	},
	"no_mod_sql_files": {
		source: "test_data/mods/no_mod_sql_files",
		expected: &modconfig.Mod{
			Name: "local",
			Queries: []*modconfig.Query{
				{
					Name: "q1", SQL: "select 1",
				},
				{
					Name: "q2", SQL: "select 2",
				},
			}},
	},
	"single_mod_nested_sql_files": {
		source: "test_data/mods/single_mod_nested_sql_files",
		expected: &modconfig.Mod{
			Name:        "m1",
			Title:       "M1",
			Description: "THIS IS M1",
			Queries: []*modconfig.Query{
				{
					Name: "queries_a_aa_q1", SQL: "select 1",
				},
				{
					Name: "queries_a_q1", SQL: "select 1",
				},
				{
					Name: "queries_b_bb_q2", SQL: "select 2",
				},
				{
					Name: "queries_b_q2", SQL: "select 2",
				},
			},
		},
	},
	"single_mod_no_query": {
		source: "test_data/mods/single_mod_no_query",
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
		source: "test_data/mods/single_mod_one_query",
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
	"single_mod_one_sql_file": {
		source: "test_data/mods/single_mod_one_sql_file",
		expected: &modconfig.Mod{
			Name:        "m1",
			Title:       "M1",
			Description: "THIS IS M1",
			Queries:     []*modconfig.Query{{Name: "q1", SQL: "select 1"}},
		},
	},
	"single_mod_sql_file_and_hcl_query": {
		source: "test_data/mods/single_mod_sql_file_and_hcl_query",
		expected: &modconfig.Mod{
			Name:        "m1",
			Title:       "M1",
			Description: "THIS IS M1",
			Queries: []*modconfig.Query{
				{
					"q1", "Q1", "THIS IS QUERY 1", "select 1",
				},
				{
					Name: "q2", SQL: "select 2",
				},
			},
		},
	},
	"single_mod_two_queries_diff_files": {
		source: "test_data/mods/single_mod_two_queries_diff_files",
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
		source: "test_data/mods/single_mod_two_queries_same_file",
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
		source:   "test_data/mods/single_mod_duplicate_query",
		expected: "ERROR",
	},
	"two_mods": {
		source:   "test_data/mods/two_mods",
		expected: "ERROR",
	},
}

func TestLoadMod(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
		return
	}
	for name, test := range testCasesLoadMod {
		executeLoadTest(t, name, test, wd)
	}
}

func executeLoadTest(t *testing.T, name string, test loadModTest, wd string) {
	modPath, err := filepath.Abs(test.source)
	if err != nil {
		t.Errorf("failed to build absolute config filepath from %s", test.source)
	}

	// set working directory to the mod path
	os.Chdir(modPath)
	// change back to original directory
	defer os.Chdir(wd)
	mod, err := LoadMod(modPath, loadWorkspaceOptions)
	if err != nil {
		if test.expected != "ERROR" {
			t.Errorf(`Test: '%s'' FAILED : unexpected error %v`, name, err)
		}
		return
	}
	if test.expected == "ERROR" {
		t.Errorf(`Test: '%s'' FAILED : expected error but did not get one`, name)
		return
	}

	expectedStr := test.expected.(*modconfig.Mod).String()
	actualString := mod.String()

	if expectedStr != actualString {
		fmt.Printf("")
		t.Errorf("Test: '%s'' FAILED : expected:\n\n%s\n\ngot:\n\n%s", name, expectedStr, actualString)
	}

}
