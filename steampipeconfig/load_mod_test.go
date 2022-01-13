package steampipeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/filepaths"
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

var testCasesLoadMod map[string]loadModTest

func init() {
	filepaths.SteampipeDir = "~/.steampipe"
	testCasesLoadMod = map[string]loadModTest{
		"no_mod_sql_files": {
			source: "testdata/mods/no_mod_sql_files",
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
			source: "testdata/mods/no_mod_hcl_queries",
			expected: &modconfig.Mod{
				ShortName: "local",
				Title:     toStringPointer("no_mod_hcl_queries"),
				FullName:  "mod.local",
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"q2": {
						ShortName:   "q2",
						FullName:    "m1.query.q2",
						Title:       toStringPointer("Q2"),
						Description: toStringPointer("THIS IS QUERY 2"),
						SQL:         toStringPointer("select 2"),
					},
				},
			},
		},
		"single_mod_duplicate_query": {
			source:   "testdata/mods/single_mod_duplicate_query",
			expected: "ERROR",
		},
		"single_mod_no_query": {
			source: "testdata/mods/single_mod_no_query",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
			},
		},
		"single_mod_one_query": {
			source: "testdata/mods/single_mod_one_query",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
				},
			},
		},
		"query_with_paramdefs": {
			source: "testdata/mods/query_with_paramdefs",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
						Params: []*modconfig.ParamDef{
							{Name: "p1",
								Description: utils.ToStringPointer("desc"),
								Default:     utils.ToStringPointer("'I am default'"),
							},
							{Name: "p2",
								Description: utils.ToStringPointer("desc 2"),
								Default:     utils.ToStringPointer("'I am default 2'"),
							},
						},
					},
				},
			},
		},
		"query_with_paramdefs_control_with_named_params": {
			source: "testdata/mods/query_with_paramdefs_control_with_named_params",
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
Params:
	Name: p1, Description: desc, Default: 'I am default'
	Name: p2, Description: desc 2, Default: 'I am default 2'
  
Controls: 

  -----
  Name: control.c1
  Title: C1
  Description: THIS IS CONTROL 1
  SQL: select 'ok' as status, 'foo' as resource, 'bar' as reason
  Parents: mod.m1
Params:
	Name: p1, Description: , Default: 'val1'
	Name: p2, Description: , Default: 'val2'
  Args:
	Args list: 'my val1','my val 2'
  
Benchmarks: 
`,
		},
		"single_mod_one_query_one_control": {
			source: "testdata/mods/single_mod_one_query_one_control",
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
			source: "testdata/mods/controls_and_groups",
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
			source:   "testdata/mods/controls_and_groups_circular",
			expected: "ERROR",
		},
		"controls_and_groups_duplicate_child": {
			source:   "testdata/mods/controls_and_groups_duplicate_child",
			expected: "ERROR",
		},
		"single_mod_one_sql_file": {
			source: "testdata/mods/single_mod_one_sql_file",
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
			source: "testdata/mods/single_mod_sql_file_and_hcl_query",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"query.q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"query.q2": {
						ShortName: "q2",
						FullName:  "query.q2",
						SQL:       toStringPointer("select 2"),
					},
				},
			},
		},
		"single_mod_two_queries_diff_files": {
			source: "testdata/mods/single_mod_two_queries_diff_files",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"q2": {
						ShortName:   "q2",
						FullName:    "m1.query.q2",
						Title:       toStringPointer("Q2"),
						Description: toStringPointer("THIS IS QUERY 2"),
						SQL:         toStringPointer("select 2"),
					},
				},
			},
		},
		"single_mod_two_queries_same_file": {
			source: "testdata/mods/single_mod_two_queries_same_file",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"q2": {
						ShortName:   "q2",
						FullName:    "m1.query.q2",
						Title:       toStringPointer("Q2"),
						Description: toStringPointer("THIS IS QUERY 2"),
						SQL:         toStringPointer("select 2"),
					},
				},
			},
		},
		"single_mod_two_sql_files": {
			source: "testdata/mods/single_mod_two_sql_files",
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
			source:   "testdata/mods/two_mods",
			expected: "ERROR",
		},
		"requires_single_simple": {
			source: "testdata/mods/requires_single_versioned",
			expected: &modconfig.Mod{
				ShortName: "m1",
				FullName:  "mod.m1",
				Require: &modconfig.Require{
					SteampipeVersionString: "v0.8.0",
					Mods: []*modconfig.ModVersionConstraint{
						{
							Name:          "github.com/turbot/aws-core",
							VersionString: "v1.0",
						},
					},
				},
			},
		},
		"requires_single_simple_aliased": {
			source: "testdata/mods/requires_single_versioned_aliased",
			expected: &modconfig.Mod{
				ShortName: "m1",
				FullName:  "mod.m1",
				Require: &modconfig.Require{
					SteampipeVersionString: "v0.8.0",
					Mods: []*modconfig.ModVersionConstraint{
						{
							Name:          "github.com/turbot/aws-core",
							VersionString: "v1.0",
							//Alias:         utils.ToStringPointer("core"),
						},
					},
				},
			},
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

	var runCtx = parse.NewRunContext(
		nil,
		modPath,
		parse.CreatePseudoResources|parse.CreateDefaultMod,
		&filehelpers.ListOptions{
			Exclude: []string{fmt.Sprintf("**/%s*", filepaths.WorkspaceDataDir)},
			Flags:   filehelpers.Files,
		})

	// set working directory to the mod path
	os.Chdir(modPath)
	// change back to original directory
	defer os.Chdir(wd)
	mod, err := LoadMod(modPath, runCtx)
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

type loadResourceNamesTest struct {
	source   string
	expected interface{}
}

var testCasesLoadResourceNames = map[string]loadResourceNamesTest{
	"test_load_mod_resource_names_workspace": {
		source: "testdata/mods/test_load_mod_resource_names_workspace",
		expected: &modconfig.WorkspaceResources{
			Benchmark: map[string]bool{"benchmark.test_workspace": true},
			Control:   map[string]bool{"control.test_workspace_1": true, "control.test_workspace_2": true, "control.test_workspace_3": true},
			Query:     map[string]bool{"query.query_control_1": true, "query.query_control_2": true, "query.query_control_3": true},
		},
	},
}

func TestLoadModResourceNames(t *testing.T) {
	for name, test := range testCasesLoadResourceNames {

		modPath, _ := filepath.Abs(test.source)
		var runCtx = parse.NewRunContext(
			nil,
			modPath,
			parse.CreatePseudoResources|parse.CreateDefaultMod,
			&filehelpers.ListOptions{
				Exclude: []string{fmt.Sprintf("**/%s*", filepaths.WorkspaceDataDir)},
				Flags:   filehelpers.Files,
			})
		names, err := LoadModResourceNames(modPath, runCtx)

		if err != nil {
			if test.expected != "ERROR" {
				t.Errorf("Test: '%s'' FAILED with unexpected error: %v", name, err)
			}
			continue
		}

		if test.expected == "ERROR" {
			t.Errorf("Test: '%s'' FAILED - expected error", name)
			continue
		}

		// to compare the benchmarks
		benchmark_expected := test.expected.(*modconfig.WorkspaceResources).Benchmark
		if reflect.DeepEqual(names.Benchmark, benchmark_expected) {
			t.Log(`"expected" is not equal to "output"`)
			t.Errorf("FAILED \nexpected: %#v\noutput: %#v", benchmark_expected, names.Benchmark)
		}

		// to compare the controls
		control_expected := test.expected.(*modconfig.WorkspaceResources).Control
		if reflect.DeepEqual(names.Control, control_expected) {
			t.Log(`"expected" is not equal to "output"`)
			t.Errorf("FAILED \nexpected: %#v\noutput: %#v", control_expected, names.Control)
		}

		// to compare the queries
		query_expected := test.expected.(*modconfig.WorkspaceResources).Query
		if reflect.DeepEqual(names.Query, query_expected) {
			t.Log(`"expected" is not equal to "output"`)
			t.Errorf("FAILED \nexpected: %#v\noutput: %#v", query_expected, names.Query)
		}
	}
}
