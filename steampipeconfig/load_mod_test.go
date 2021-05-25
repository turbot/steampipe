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

var testCasesLoadMod = map[string]loadModTest{
	"report": {
		source: "test_data/mods/report",
		expected: &modconfig.Mod{
			ShortName: "local",
			FullName:  "mod.local",
			Queries: map[string]*modconfig.Query{
				"q1": {
					FullName: "query.q1", SQL: toStringPointer("select 1"),
				},
				"q2": {
					FullName: "query.q2", SQL: toStringPointer("select 2"),
				},
			}},
	},
	//"no_mod_sql_files": {
	//	source: "test_data/mods/no_mod_sql_files",
	//	expected: &modconfig.Mod{
	//		ShortName: "local",
	//		FullName:  "mod.local",
	//		Title:     toStringPointer("no_mod_sql_files"),
	//		Queries: map[string]*modconfig.Query{
	//			"q1": {
	//				FullName: "query.q1", SQL: toStringPointer("select 1"),
	//			},
	//			"q2": {
	//				FullName: "query.q2", SQL: toStringPointer("select 2"),
	//			},
	//		}},
	//},
	//"no_mod_hcl_queries": {
	//	source: "test_data/mods/no_mod_hcl_queries",
	//	expected: &modconfig.Mod{
	//		ShortName: "local",
	//		FullName:  "mod.local",
	//		Title:     toStringPointer("no_mod_hcl_queries"),
	//		Queries: map[string]*modconfig.Query{
	//			"q1": {
	//				FullName: "query.q1", Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
	//			},
	//			"q2": {
	//				FullName: "query.q2", Title: toStringPointer("Q2"), Description: toStringPointer("THIS IS QUERY 2"), SQL: toStringPointer("select 2"),
	//			},
	//		},
	//	},
	//},
	//"single_mod_duplicate_query": {
	//	source:   "test_data/mods/single_mod_duplicate_query",
	//	expected: "ERROR",
	//},
	//"single_mod_no_query": {
	//	source: "test_data/mods/single_mod_no_query",
	//	expected: &modconfig.Mod{
	//		ShortName:   "m1",
	//		FullName:    "mod.m1",
	//		Title:       toStringPointer("M1"),
	//		Description: toStringPointer("THIS IS M1"),
	//	},
	//},
	//"single_mod_one_query": {
	//	source: "test_data/mods/single_mod_one_query",
	//	expected: &modconfig.Mod{
	//		ShortName:   "m1",
	//		FullName:    "mod.m1",
	//		Title:       toStringPointer("M1"),
	//		Description: toStringPointer("THIS IS M1"),
	//		Queries: map[string]*modconfig.Query{
	//			"q1": {
	//				FullName: "query.q1", Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
	//			},
	//		},
	//	},
	//},
	//
	//// TODO we need a way of setting parent on the control in the expected
	////"single_mod_one_query_one_control": {
	////	source: "test_data/mods/single_mod_one_query_one_control",
	////	expected: &modconfig.Mod{
	////		ShortName:   "m1",
	////		FullName:    "mod.m1",
	////		Title:       toStringPointer("M1"),
	////		Description: toStringPointer("THIS IS M1"),
	////		Queries: map[string]*modconfig.Query{
	////			"q1": {
	////
	////				FullName: "query.q1", Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
	////			},
	////		},
	////		Controls: map[string]*modconfig.Control{
	////			"c1": {
	////				ShortName:   "c1",
	////				Title:       toStringPointer("C1"),
	////				Description: toStringPointer("THIS IS CONTROL 1"),
	////				SQL:         toStringPointer("select 'pass' as result"),
	////			},
	////		},
	////	},
	////},
	//// TODO need to update to use children
	////"controls_and_groups": {
	////	source: "test_data/mods/controls_and_groups",
	////	expected: `ShortName:
	////Title: M1
	////Description: THIS IS M1
	////Mod Dependencies: []
	////Plugin Dependencies: []
	////Queries:
	////
	////Controls:
	////
	//// -----
	//// ShortName: c1
	//// Title:
	//// Description:
	//// Query: select 'pass' as result
	//// Parent: benchmark.cg_1_1_1
	//// Labels: []
	//// Links: []
	////
	////
	//// -----
	//// ShortName: c2
	//// Title:
	//// Description:
	//// Query: select 'pass' as result
	//// Parent: benchmark.cg_1_1_2
	//// Labels: []
	//// Links: []
	////
	////
	//// -----
	//// ShortName: c3
	//// Title:
	//// Description:
	//// Query: select 'pass' as result
	//// Parent: benchmark.cg_1_1
	//// Labels: []
	//// Links: []
	////
	////
	//// -----
	//// ShortName: c4
	//// Title:
	//// Description:
	//// Query: select 'pass' as result
	//// Parent: benchmark.cg_1_1_2
	//// Labels: []
	//// Links: []
	////
	////
	//// -----
	//// ShortName: c5
	//// Title:
	//// Description:
	//// Query: select 'pass' as result
	//// Parent: benchmark.cg_1_1_2
	//// Labels: []
	//// Links: []
	////
	////
	//// -----
	//// ShortName: c6
	//// Title:
	//// Description:
	//// Query: select 'FAIL' as result
	//// Parent:
	//// Labels: []
	//// Links: []
	////
	////Control Groups:
	////
	//// -----
	//// ShortName:
	//// Title:
	//// Description:
	//// Parent:
	//// Labels: []
	//// Children:
	////   control.cg_1_1
	////   control.cg_1_2
	////
	////
	//// -----
	//// ShortName:
	//// Title:
	//// Description:
	//// Parent: benchmark.cg_1
	//// Labels: []
	//// Children:
	////   control.c3
	////   control.cg_1_1_1
	////   control.cg_1_1_2
	////
	////
	//// -----
	//// ShortName:
	//// Title:
	//// Description:
	//// Parent: benchmark.cg_1_1
	//// Labels: []
	//// Children:
	////   control.c1
	////
	////
	//// -----
	//// ShortName:
	//// Title:
	//// Description:
	//// Parent: benchmark.cg_1_1
	//// Labels: []
	//// Children:
	////   control.c2
	////   control.c4
	////   control.c5
	////
	////
	//// -----
	//// ShortName:
	//// Title:
	//// Description:
	//// Parent: benchmark.cg_1
	//// Labels: []
	//// Children:
	////
	////`,
	////},
	//"controls_and_groups_circular": {
	//	source:   "test_data/mods/controls_and_groups_circular",
	//	expected: "ERROR",
	//},
	//"single_mod_one_sql_file": {
	//	source: "test_data/mods/single_mod_one_sql_file",
	//	expected: &modconfig.Mod{
	//		ShortName:   "m1",
	//		FullName:    "mod.m1",
	//		Title:       toStringPointer("M1"),
	//		Description: toStringPointer("THIS IS M1"),
	//		Queries:     map[string]*modconfig.Query{"q1": {FullName: "query.q1", SQL: toStringPointer("select 1")}},
	//	},
	//},
	//"single_mod_sql_file_and_hcl_query": {
	//	source: "test_data/mods/single_mod_sql_file_and_hcl_query",
	//	expected: &modconfig.Mod{
	//		ShortName:   "m1",
	//		FullName:    "mod.m1",
	//		Title:       toStringPointer("M1"),
	//		Description: toStringPointer("THIS IS M1"),
	//		Queries: map[string]*modconfig.Query{
	//			"q1": {
	//				FullName: "query.q1", Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
	//			},
	//			"q2": {
	//				FullName: "query.q2", SQL: toStringPointer("select 2"),
	//			},
	//		},
	//	},
	//},
	//"single_mod_two_queries_diff_files": {
	//	source: "test_data/mods/single_mod_two_queries_diff_files",
	//	expected: &modconfig.Mod{
	//		ShortName:   "m1",
	//		FullName:    "mod.m1",
	//		Title:       toStringPointer("M1"),
	//		Description: toStringPointer("THIS IS M1"),
	//		Queries: map[string]*modconfig.Query{
	//			"q1": {
	//				FullName: "query.q1", Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
	//			},
	//			"q2": {
	//				FullName: "query.q2", Title: toStringPointer("Q2"), Description: toStringPointer("THIS IS QUERY 2"), SQL: toStringPointer("select 2"),
	//			},
	//		},
	//	},
	//},
	//"single_mod_two_queries_same_file": {
	//	source: "test_data/mods/single_mod_two_queries_same_file",
	//	expected: &modconfig.Mod{
	//		ShortName:   "m1",
	//		FullName:    "mod.m1",
	//		Title:       toStringPointer("M1"),
	//		Description: toStringPointer("THIS IS M1"),
	//		Queries: map[string]*modconfig.Query{
	//			"q1": {
	//				FullName: "query.q1", Title: toStringPointer("Q1"), Description: toStringPointer("THIS IS QUERY 1"), SQL: toStringPointer("select 1"),
	//			},
	//			"q2": {
	//				FullName: "query.q2", Title: toStringPointer("Q2"), Description: toStringPointer("THIS IS QUERY 2"), SQL: toStringPointer("select 2"),
	//			},
	//		},
	//	},
	//},
	//"single_mod_two_sql_files": {
	//	source: "test_data/mods/single_mod_two_sql_files",
	//	expected: &modconfig.Mod{
	//		ShortName:   "m1",
	//		FullName:    "mod.m1",
	//		Title:       toStringPointer("M1"),
	//		Description: toStringPointer("THIS IS M1"),
	//		Queries: map[string]*modconfig.Query{
	//			"q1": {
	//				FullName: "query.q1", SQL: toStringPointer("select 1"),
	//			},
	//			"q2": {
	//				FullName: "query.q2", SQL: toStringPointer("select 2"),
	//			},
	//		},
	//	},
	//},
	//"two_mods": {
	//	source:   "test_data/mods/two_mods",
	//	expected: "ERROR",
	//},
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
