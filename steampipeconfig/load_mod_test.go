package steampipeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/utils"
)

// TODO add tests for reflection data

var toStringPointer = utils.ToStringPointer

type loadModTest struct {
	source   string
	expected interface{}
}

var loadWorkspaceOptions = &parse.ParseModOptions{
	Flags: parse.CreatePseudoResources | parse.CreateDefaultMod,
	ListOptions: &filehelpers.ListOptions{
		Exclude: []string{fmt.Sprintf("**/%s*", constants.WorkspaceDataDir)},
		Flags:   filehelpers.Files,
	},
}
var testCasesLoadMod map[string]loadModTest

func init() {
	constants.SteampipeDir = "~/.steampipe"
	testCasesLoadMod = map[string]loadModTest{
		"no_mod_sql_files": {
			source: "test_data/mods/no_mod_sql_files",
			expected: &modconfig.Mod{
				ShortName: "local",
				FullName:  "mod.local",
				Title:     toStringPointer("no_mod_sql_files"),
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName: "q1",
						FullName:  "query.q1",
						SQL:       toStringPointer("select 1"),
					},
					"q2": {
						ShortName: "q2",
						FullName:  "query.q2",
						SQL:       toStringPointer("select 2"),
					},
				}},
		},
		"no_mod_hcl_queries": {
			source: "test_data/mods/no_mod_hcl_queries",
			expected: &modconfig.Mod{
				ShortName: "local",
				Title:     toStringPointer("no_mod_hcl_queries"),
				FullName:  "mod.local",
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName:   "q1",
						FullName:    "query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"q2": {
						ShortName:   "q2",
						FullName:    "query.q2",
						Title:       toStringPointer("Q2"),
						Description: toStringPointer("THIS IS QUERY 2"),
						SQL:         toStringPointer("select 2"),
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
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
			},
		},
		"single_mod_one_query": {
			source: "test_data/mods/single_mod_one_query",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName:   "q1",
						FullName:    "query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
				},
			},
		},
		"query_with_paramdefs_control_with_positional_params": {
			source: "test_data/mods/query_with_paramdefs_control_with_positional_params",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName:   "q1",
						FullName:    "query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
				},
			},
		},
		"single_mod_one_query_one_control": {
			source: "test_data/mods/single_mod_one_query_one_control",
			expected: `Name: mod.m1
Title: M1
Description: THIS IS M1 
Version: 
Queries: 

  -----
  Name: query.q1
  Title: Q1
  Description: THIS IS QUERY 1
  SQL: select 1

Controls: 

  -----
  Name: control.c1
  Title: C1
  Description: THIS IS CONTROL 1
  SQL: select 'ok' as status, 'foo' as resource, 'bar' as reason
  Parents: mod.m1

Benchmarks: 
`,
		},
		"controls_and_groups": {
			source: "test_data/mods/controls_and_groups",
			expected: `Name: mod.m1
Title: M1
Description: THIS IS M1 
Version: 
Queries: 

Controls: 

  -----
  Name: control.c1
  Title: 
  Description: 
  SQL: select 'pass' as result
  Parents: benchmark.cg_1_1_1


  -----
  Name: control.c2
  Title: 
  Description: 
  SQL: select 'pass' as result
  Parents: benchmark.cg_1_1_2


  -----
  Name: control.c3
  Title: 
  Description: 
  SQL: select 'pass' as result
  Parents: benchmark.cg_1_1


  -----
  Name: control.c4
  Title: 
  Description: 
  SQL: select 'pass' as result
  Parents: benchmark.cg_1_1_2


  -----
  Name: control.c5
  Title: 
  Description: 
  SQL: select 'pass' as result
  Parents: benchmark.cg_1_1_2


  -----
  Name: control.c6
  Title: 
  Description: 
  SQL: select 'FAIL' as result
  Parents: mod.m1

Benchmarks: 

	 -----
	 Name: benchmark.cg_1
	 Title: 
	 Description: 
	 Parent: mod.m1
	 Children:
	   benchmark.cg_1_1
    benchmark.cg_1_2
	

	 -----
	 Name: benchmark.cg_1_1
	 Title: 
	 Description: 
	 Parent: benchmark.cg_1
	 Children:
	   benchmark.cg_1_1_1
    benchmark.cg_1_1_2
    control.c3
	

	 -----
	 Name: benchmark.cg_1_1_1
	 Title: 
	 Description: 
	 Parent: benchmark.cg_1_1
	 Children:
	   control.c1
	

	 -----
	 Name: benchmark.cg_1_1_2
	 Title: 
	 Description: 
	 Parent: benchmark.cg_1_1
	 Children:
	   control.c2
    control.c4
    control.c5
	

	 -----
	 Name: benchmark.cg_1_2
	 Title: 
	 Description: 
	 Parent: benchmark.cg_1
	 Children:
	   
	`,
		},
		"controls_and_groups_circular": {
			source:   "test_data/mods/controls_and_groups_circular",
			expected: "ERROR",
		},
		"controls_and_groups_duplicate_child": {
			source:   "test_data/mods/controls_and_groups_duplicate_child",
			expected: "ERROR",
		},
		"single_mod_one_sql_file": {
			source: "test_data/mods/single_mod_one_sql_file",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{"q1": {ShortName: "q1", FullName: "query.q1",
					SQL: toStringPointer("select 1")}},
			},
		},
		"single_mod_sql_file_and_hcl_query": {
			source: "test_data/mods/single_mod_sql_file_and_hcl_query",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					// TODO investigate why pseudo resources have no "query." at start of key
					"query.q1": {
						ShortName:   "q1",
						FullName:    "query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"q2": {
						ShortName: "q2",
						FullName:  "query.q2",
						SQL:       toStringPointer("select 2"),
					},
				},
			},
		},
		"single_mod_two_queries_diff_files": {
			source: "test_data/mods/single_mod_two_queries_diff_files",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName:   "q1",
						FullName:    "query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"q2": {
						ShortName:   "q2",
						FullName:    "query.q2",
						Title:       toStringPointer("Q2"),
						Description: toStringPointer("THIS IS QUERY 2"),
						SQL:         toStringPointer("select 2"),
					},
				},
			},
		},
		"single_mod_two_queries_same_file": {
			source: "test_data/mods/single_mod_two_queries_same_file",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName:   "q1",
						FullName:    "query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"q2": {
						ShortName:   "q2",
						FullName:    "query.q2",
						Title:       toStringPointer("Q2"),
						Description: toStringPointer("THIS IS QUERY 2"),
						SQL:         toStringPointer("select 2"),
					},
				},
			},
		},
		"single_mod_two_sql_files": {
			source: "test_data/mods/single_mod_two_sql_files",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName: "q1",
						FullName:  "query.q1",
						SQL:       toStringPointer("select 1"),
					},
					"q2": {
						ShortName: "q2",
						FullName:  "query.q2",
						SQL:       toStringPointer("select 2"),
					},
				},
			},
		},
		"two_mods": {
			source:   "test_data/mods/two_mods",
			expected: "ERROR",
		},
	}
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
