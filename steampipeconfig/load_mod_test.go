package steampipeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

var toStringPointer = utils.ToStringPointer

type loadModTest struct {
	source   string
	expected interface{}
}

var loadWorkspaceOptions = &LoadModOptions{
	Exclude: []string{fmt.Sprintf("**/%s*", constants.WorkspaceDataDir)},
	Flags:   CreatePseudoResources | CreateDefaultMod,
}

var testCasesLoadMod = map[string]loadModTest{
	"no_mod_sql_files": {
		source: "test_data/mods/no_mod_sql_files",
		expected: &modconfig.Mod{
			ShortName: toStringPointer("local"),
			Queries: map[string]*modconfig.Query{
				"q1": {
					ShortName: toStringPointer("q1"), SQL: toStringPointer("select 1"),
				},
				"q2": {
					ShortName: toStringPointer("q2"), SQL: toStringPointer("select 2"),
				},
			}},
	},
	"no_mod_hcl_queries": {
		source: "test_data/mods/no_mod_hcl_queries",
		expected: &modconfig.Mod{
			ShortName: toStringPointer("local"),
			Queries: map[string]*modconfig.Query{
				"q1": {
					toStringPointer("q1"), toStringPointer("Q1"), toStringPointer("THIS IS QUERY 1"), toStringPointer("select 1"),
				},
				"q2": {
					toStringPointer("q2"), toStringPointer("Q2"), toStringPointer("THIS IS QUERY 2"), toStringPointer("select 2"),
				},
			},
		},
	},
	"single_mod_duplicate_query": {
		source:   "test_data/mods/single_mod_duplicate_query",
		expected: "ERROR",
	},
	"single_mod_no_query": {
		source: "test_data/mods/single_mod_no_query",
		expected: &modconfig.Mod{
			ShortName:   toStringPointer("m1"),
			Title:       toStringPointer("M1"),
			Description: toStringPointer("THIS IS M1"),
			ModDepends: []*modconfig.ModVersion{
				{"github.com/turbot/m2", "0.0.0", toStringPointer("_m1")},
			},
		},
	},
	"single_mod_one_query": {
		source: "test_data/mods/single_mod_one_query",
		expected: &modconfig.Mod{
			ShortName:   toStringPointer("m1"),
			Title:       toStringPointer("M1"),
			Description: toStringPointer("THIS IS M1"),
			ModDepends: []*modconfig.ModVersion{
				{"github.com/turbot/m2", "0.0.0", toStringPointer("_m1")},
			},
			Queries: map[string]*modconfig.Query{
				"q1": {
					ShortName: toStringPointer("q1"), Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
				},
			},
		},
	},
	"single_mod_one_query_one_control": {
		source: "test_data/mods/single_mod_one_query_one_control",
		expected: &modconfig.Mod{
			ShortName:   toStringPointer("m1"),
			Title:       toStringPointer("M1"),
			Description: toStringPointer("THIS IS M1"),
			Queries: map[string]*modconfig.Query{
				"q1": {

					ShortName: toStringPointer("q1"), Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
				},
			},
			Controls: map[string]*modconfig.Control{
				"c1": {
					ShortName:   toStringPointer("c1"),
					Title:       toStringPointer("C1"),
					Description: toStringPointer("THIS IS CONTROL 1"),
					Query:       toStringPointer("select 'pass' as result"),
					Labels:      &[]string{"demo", "prod", "steampipe"},
				},
			},
		},
	},
	"controls_and_groups": {
		source: "test_data/mods/controls_and_groups",
		expected: `Name: 
Title: M1
Description: THIS IS M1 
Mod Dependencies: []
Plugin Dependencies: []
Queries: 

Controls: 

  -----
  Name: c1
  Title: 
  Description: 
  Query: select 'pass' as result
  Parent: control_group.cg_1_1_1
  Labels: []
  Links: []


  -----
  Name: c2
  Title: 
  Description: 
  Query: select 'pass' as result
  Parent: control_group.cg_1_1_2
  Labels: []
  Links: []


  -----
  Name: c3
  Title: 
  Description: 
  Query: select 'pass' as result
  Parent: control_group.cg_1_1
  Labels: []
  Links: []


  -----
  Name: c4
  Title: 
  Description: 
  Query: select 'pass' as result
  Parent: control_group.cg_1_1_2
  Labels: []
  Links: []


  -----
  Name: c5
  Title: 
  Description: 
  Query: select 'pass' as result
  Parent: control_group.cg_1_1_2
  Labels: []
  Links: []


  -----
  Name: c6
  Title: 
  Description: 
  Query: select 'FAIL' as result
  Parent: 
  Labels: []
  Links: []

Control Groups: 

  -----
  Name: 
  Title: 
  Description: 
  Parent:  
  Labels: []
  Children: 
    control.cg_1_1
    control.cg_1_2


  -----
  Name: 
  Title: 
  Description: 
  Parent: control_group.cg_1 
  Labels: []
  Children: 
    control.c3
    control.cg_1_1_1
    control.cg_1_1_2


  -----
  Name: 
  Title: 
  Description: 
  Parent: control_group.cg_1_1 
  Labels: []
  Children: 
    control.c1


  -----
  Name: 
  Title: 
  Description: 
  Parent: control_group.cg_1_1 
  Labels: []
  Children: 
    control.c2
    control.c4
    control.c5


  -----
  Name: 
  Title: 
  Description: 
  Parent: control_group.cg_1 
  Labels: []
  Children: 
    
`,
	},
	"controls_and_groups_circular": {
		source:   "test_data/mods/controls_and_groups_circular",
		expected: "ERROR",
	},
	"single_mod_one_sql_file": {
		source: "test_data/mods/single_mod_one_sql_file",
		expected: &modconfig.Mod{
			ShortName:   toStringPointer("m1"),
			Title:       toStringPointer("M1"),
			Description: toStringPointer("THIS IS M1"),
			Queries:     map[string]*modconfig.Query{"q1": {ShortName: toStringPointer("q1"), SQL: toStringPointer("select 1")}},
		},
	},
	"single_mod_sql_file_and_hcl_query": {
		source: "test_data/mods/single_mod_sql_file_and_hcl_query",
		expected: &modconfig.Mod{
			ShortName:   toStringPointer("m1"),
			Title:       toStringPointer("M1"),
			Description: toStringPointer("THIS IS M1"),
			Queries: map[string]*modconfig.Query{
				"q1": {
					ShortName: toStringPointer("q1"), Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
				},
				"q2": {
					ShortName: toStringPointer("q2"), SQL: toStringPointer("select 2"),
				},
			},
		},
	},
	"single_mod_two_queries_diff_files": {
		source: "test_data/mods/single_mod_two_queries_diff_files",
		expected: &modconfig.Mod{
			ShortName:   toStringPointer("m1"),
			Title:       toStringPointer("M1"),
			Description: toStringPointer("THIS IS M1"),
			ModDepends: []*modconfig.ModVersion{
				{"github.com/turbot/m2", "0.0.0", toStringPointer("_m1")},
			},
			Queries: map[string]*modconfig.Query{
				"q1": {
					ShortName: toStringPointer("q1"), Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
				},
				"q2": {
					ShortName: toStringPointer("q2"), Title: toStringPointer("Q2"), Description: toStringPointer("THIS IS QUERY 2"), SQL: toStringPointer("select 2"),
				},
			},
		},
	},
	"single_mod_two_queries_same_file": {
		source: "test_data/mods/single_mod_two_queries_same_file",
		expected: &modconfig.Mod{
			ShortName:   toStringPointer("m1"),
			Title:       toStringPointer("M1"),
			Description: toStringPointer("THIS IS M1"),
			ModDepends: []*modconfig.ModVersion{
				{"github.com/turbot/m2", "0.0.0", toStringPointer("_m1")},
			},
			Queries: map[string]*modconfig.Query{
				"q1": {
					ShortName: toStringPointer("q1"), Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
				},
				"q2": {
					ShortName: toStringPointer("q2"), Title: toStringPointer("Q2"), Description: toStringPointer("THIS IS QUERY 2"), SQL: toStringPointer("select 2"),
				},
			},
		},
	},
	"single_mod_two_sql_files": {
		source: "test_data/mods/single_mod_two_sql_files",
		expected: &modconfig.Mod{
			ShortName:   toStringPointer("m1"),
			Title:       toStringPointer("M1"),
			Description: toStringPointer("THIS IS M1"),
			Queries: map[string]*modconfig.Query{
				"q1": {
					ShortName: toStringPointer("q1"), SQL: toStringPointer("select 1"),
				},
				"q2": {
					ShortName: toStringPointer("q2"), SQL: toStringPointer("select 2"),
				},
			},
		},
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

	expectedStr, ok := test.expected.(string)
	if !ok {
		expectedStr = test.expected.(*modconfig.Mod).String()
	}
	actualString := mod.String()

	if expectedStr != actualString {
		fmt.Printf("")
		t.Errorf("Test: '%s'' FAILED : expected:\n\n%s\n\ngot:\n\n%s", name, expectedStr, actualString)
	}

}
