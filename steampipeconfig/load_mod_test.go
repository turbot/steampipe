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
				Require:   modconfig.NewRequire(),
				Title:     toStringPointer("no_mod_sql_files"),
				Queries: map[string]*modconfig.Query{
					"local.query.q1": {
						ShortName: "q1",
						FullName:  "local.query.q1",
						SQL:       toStringPointer("select 1"),
					},
					"local.query.q2": {
						ShortName: "q2",
						FullName:  "local.query.q2",
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
				Require:   modconfig.NewRequire(),
				Queries: map[string]*modconfig.Query{
					"local.query.q1": {
						ShortName:   "q1",
						FullName:    "local.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"local.query.q2": {
						ShortName:   "q2",
						FullName:    "local.query.q2",
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
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
			},
		},
		"single_mod_one_query": {
			source: "testdata/mods/single_mod_one_query",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
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
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
						Params: []*modconfig.ParamDef{
							{
								Name:        "p1",
								FullName:    "param.p1",
								Description: utils.ToStringPointer("desc"),
								Default:     utils.ToStringPointer("'I am default'"),
							},
							{
								Name:        "p2",
								FullName:    "param.p2",
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
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
						Params: []*modconfig.ParamDef{
							{
								Name:        "p1",
								FullName:    "param.p1",
								Description: utils.ToStringPointer("desc"),
								Default:     utils.ToStringPointer("'I am default'"),
							},
							{
								Name:        "p2",
								FullName:    "param.p2",
								Description: utils.ToStringPointer("desc 2"),
								Default:     utils.ToStringPointer("'I am default 2'"),
							},
						},
					},
				},
				Controls: map[string]*modconfig.Control{
					"m1.control.c1": {
						ShortName:   "c1",
						FullName:    "m1.control.c1",
						Title:       toStringPointer("C1"),
						Description: toStringPointer("THIS IS CONTROL 1"),
						SQL:         toStringPointer("select 'ok' as status, 'foo' as resource, 'bar' as reason"),
						Params: []*modconfig.ParamDef{
							{
								Name:     "p1",
								FullName: "param.p1",

								Default: utils.ToStringPointer("'val1'"),
							},
							{
								Name:     "p2",
								FullName: "param.p2",
								Default:  utils.ToStringPointer("'val2'"),
							},
						},
						Args: &modconfig.QueryArgs{ArgsList: []string{"'my val1'", "'my val2'"}},
					},
				},
			},
		},
		"single_mod_one_query_one_control": {
			source: "testdata/mods/single_mod_one_query_one_control",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
				},
				Controls: map[string]*modconfig.Control{
					"m1.control.c1": {
						ShortName:   "c1",
						FullName:    "m1.control.c1",
						Title:       toStringPointer("C1"),
						Description: toStringPointer("THIS IS CONTROL 1"),
						SQL:         toStringPointer("select 'ok' as status, 'foo' as resource, 'bar' as reason"),
						Args:        &modconfig.QueryArgs{},
					},
				},
			},
		},
		"controls_and_groups": {
			source: "testdata/mods/controls_and_groups",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
						ShortName: "q1",
						FullName:  "m1.query.q1",
						SQL:       toStringPointer("select 1"),
					},
				},
				Controls: map[string]*modconfig.Control{
					"m1.control.c1": {
						ShortName: "c1",
						FullName:  "m1.control.c1",
						SQL:       toStringPointer("select 'pass' as result"),
						Args:      &modconfig.QueryArgs{},
					},
					"m1.control.c2": {
						ShortName: "c2",
						FullName:  "m1.control.c2",
						SQL:       toStringPointer("select 'pass' as result"),
						Args:      &modconfig.QueryArgs{},
					},
					"m1.control.c3": {
						ShortName: "c3",
						FullName:  "m1.control.c3",
						SQL:       toStringPointer("select 'pass' as result"),
						Args:      &modconfig.QueryArgs{},
					},
					"m1.control.c4": {
						ShortName: "c4",
						FullName:  "m1.control.c4",
						SQL:       toStringPointer("select 'pass' as result"),
						Args:      &modconfig.QueryArgs{},
					},
					"m1.control.c5": {
						ShortName: "c5",
						FullName:  "m1.control.c5",
						SQL:       toStringPointer("select 'pass' as result"),
						Args:      &modconfig.QueryArgs{},
					},
					"m1.control.c6": {
						ShortName: "c6",
						FullName:  "m1.control.c6",
						SQL:       toStringPointer("select 'fail' as result"),
						Args:      &modconfig.QueryArgs{},
					},
				},
				Benchmarks: map[string]*modconfig.Benchmark{
					"m1.benchmark.cg_1": {
						ShortName:        "cg_1",
						FullName:         "m1.benchmark.cg_1",
						ChildNames:       []modconfig.NamedItem{{Name: "m1.benchmark.cg_1_1"}, {Name: "m1.benchmark.cg_1_2"}},
						ChildNameStrings: []string{"m1.benchmark.cg_1_1", "m1.benchmark.cg_1_2"},
					},
					"m1.benchmark.cg_1_1": {
						ShortName:        "cg_1_1",
						FullName:         "m1.benchmark.cg_1_1",
						ChildNames:       []modconfig.NamedItem{{Name: "m1.benchmark.cg_1_1_1"}, {Name: "m1.benchmark.cg_1_1_2"}},
						ChildNameStrings: []string{"m1.benchmark.cg_1_1_1", "m1.benchmark.cg_1_1_2"},
					},
					"m1.benchmark.cg_1_2": {
						ShortName:        "cg_1_2",
						FullName:         "m1.benchmark.cg_1_2",
						ChildNames:       []modconfig.NamedItem{},
						ChildNameStrings: []string{},
					},
					"m1.benchmark.cg_1_1_1": {
						ShortName:        "cg_1_1_1",
						FullName:         "m1.benchmark.cg_1_1_1",
						ChildNames:       []modconfig.NamedItem{{Name: "m1.control.c1"}},
						ChildNameStrings: []string{"m1.control.c1"},
					},
					"m1.benchmark.cg_1_1_2": {
						ShortName:        "cg_1_1_2",
						FullName:         "m1.benchmark.cg_1_1_2",
						ChildNames:       []modconfig.NamedItem{{Name: "m1.control.c2"}, {Name: "m1.control.c4"}, {Name: "m1.control.c5"}},
						ChildNameStrings: []string{"m1.control.c2", "m1.control.c4", "m1.control.c5"},
					},
				},
			},
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
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{"m1.query.q1": {ShortName: "q1", FullName: "m1.query.q1",
					SQL: toStringPointer("select 1")}},
			},
		},

		"single_mod_sql_file_and_hcl_query": {
			source: "testdata/mods/single_mod_sql_file_and_hcl_query",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"m1.query.q2": {
						ShortName: "q2",
						FullName:  "m1.query.q2",
						SQL:       toStringPointer("select 2"),
					},
				},
			},
		},
		// upto here
		// "single_mod_sql_file_and_clashing_hcl_query": {
		// 	source:   "testdata/mods/single_mod_sql_file_and_clashing_hcl_query",
		// 	expected: "ERROR",
		// },
		// till here
		"single_mod_two_queries_diff_files": {
			source: "testdata/mods/single_mod_two_queries_diff_files",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"m1.query.q2": {
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
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select 1"),
					},
					"m1.query.q2": {
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
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
						ShortName: "q1",
						FullName:  "m1.query.q1",
						SQL:       toStringPointer("select 1"),
					},
					"m1.query.q2": {
						ShortName: "q2",
						FullName:  "m1.query.q2",
						SQL:       toStringPointer("select 2"),
					},
				},
			},
		},
		"simple_report": {
			source: "testdata/mods/simple_report",
			expected: &modconfig.Mod{
				ShortName:   "simple_report",
				FullName:    "mod.simple_report",
				Require:     modconfig.NewRequire(),
				Title:       toStringPointer("simple report"),
				Description: toStringPointer("this mod contains a simple report"),
				Reports: map[string]*modconfig.ReportContainer{
					"simple_report.report.simple_report": {
						ShortName:       "simple_report",
						FullName:        "simple_report.report.simple_report",
						UnqualifiedName: "report.simple_report",
						ChildNames:      []string{"simple_report.text.report_simple_report_text_0", "simple_report.chart.report_simple_report_chart_0"},
					},
				},
				ReportCharts: map[string]*modconfig.ReportChart{
					"simple_report.chart.report_simple_report_chart_0": {
						FullName:        "simple_report.chart.report_simple_report_chart_0",
						ShortName:       "report_simple_report_chart_0",
						UnqualifiedName: "chart.report_simple_report_chart_0",
						Title:           toStringPointer("a simple query"),
						SQL:             toStringPointer("select 1"),
					},
				},
				ReportTexts: map[string]*modconfig.ReportText{
					"simple_report.text.report_simple_report_text_0": {
						FullName:        "simple_report.text.report_simple_report_text_0",
						ShortName:       "report_simple_report_text_0",
						UnqualifiedName: "text.report_simple_report_text_0",
						Value:           toStringPointer("a simple report"),
					},
				},
			},
		},
		// upto here
		// "simple_container_report": {
		// 	source: "testdata/mods/simple_container_report",
		// 	expected: &modconfig.Mod{
		// 		ShortName:   "simple_container_report",
		// 		FullName:    "mod.simple_container_report",
		// 		Require:     modconfig.NewRequire(),
		// 		Title:       toStringPointer("simple report with container"),
		// 		Description: toStringPointer("this mod contains a simple report with containers"),
		// 		Reports: map[string]*modconfig.ReportContainer{
		// 			"simple_report.report.simple_report": {
		// 				ShortName:       "simple_report",
		// 				FullName:        "simple_report.report.simple_report",
		// 				UnqualifiedName: "report.simple_report",
		// 				ChildNames:      []string{"simple_report.text.report_simple_report_text_0", "simple_report.chart.report_simple_report_chart_0"},
		// 			},
		// 		},
		// 	},
		// },
		// "sibling_containers_report": {
		// 	source: "testdata/mods/sibling_containers_report",
		// 	expected: &modconfig.Mod{
		// 		ShortName:   "sibling_containers_report",
		// 		FullName:    "mod.sibling_containers_report",
		// 		Require:     modconfig.NewRequire(),
		// 		Title:       toStringPointer("report with multiple sibling containers"),
		// 		Description: toStringPointer("this mod contains a report with multiple sibling containers"),
		// 	},
		// },
		// "nested_containers_report": {
		// 	source: "testdata/mods/nested_containers_report",
		// 	expected: &modconfig.Mod{
		// 		ShortName:   "nested_containers_report",
		// 		FullName:    "mod.nested_containers_report",
		// 		Require:     modconfig.NewRequire(),
		// 		Title:       toStringPointer("report with nested containers"),
		// 		Description: toStringPointer("this mod contains a report with nested containers"),
		// 	},
		// },
		//"two_mods": {
		//	source:   "testdata/mods/two_mods",
		//	expected: "ERROR",
		//},
	}
}

func TestLoadMod(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
		return
	}
	for name, test := range testCasesLoadMod {
		loadTest(t, name, test, wd)
	}
}

func loadTest(t *testing.T, name string, test loadModTest, wd string) {
	modPath, err := filepath.Abs(test.source)
	if err != nil {
		t.Errorf("failed to build absolute config filepath from %s", test.source)
	}

	var runCtx = parse.NewRunContext(
		nil,
		modPath,
		parse.CreatePseudoResources|parse.CreateDefaultMod,
		&filehelpers.ListOptions{
			Include: []string{"**/*.sp"},
			Exclude: []string{fmt.Sprintf("**/%s*", filepaths.WorkspaceDataDir)},
			Flags:   filehelpers.Files,
		})

	// set working directory to the mod path
	os.Chdir(modPath)
	// change back to original directory
	defer os.Chdir(wd)
	actualMod, err := LoadMod(modPath, runCtx)
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

	expectedMod := test.expected.(*modconfig.Mod)
	expectedMod.PopulateResourceMaps()
	// ensure parents and children are set correctly in expected mod (this is normally done as part of decode)
	setChildren(expectedMod)
	expectedMod.BuildResourceTree(nil)

	diff := actualMod.Diff(expectedMod)

	if diff.HasChanges() {
		fmt.Printf("")

		t.Errorf("Test: '%s'' FAILED", name)
	}
}

func setChildren(mod *modconfig.Mod) {
	for _, benchmark := range mod.Benchmarks {
		for _, childName := range benchmark.ChildNames {
			parsed, _ := modconfig.ParseResourceName(childName.Name)
			child, _ := modconfig.GetResource(mod, parsed)
			benchmark.Children = append(benchmark.Children, child.(modconfig.ModTreeItem))
		}
	}
	for _, container := range mod.ReportContainers {
		var children []modconfig.ModTreeItem
		for _, childName := range container.ChildNames {
			parsed, _ := modconfig.ParseResourceName(childName)
			child, _ := modconfig.GetResource(mod, parsed)
			children = append(children, child.(modconfig.ModTreeItem))
		}
		container.SetChildren(children)

	}
	for _, report := range mod.Reports {
		var children []modconfig.ModTreeItem
		for _, childName := range report.ChildNames {
			parsed, _ := modconfig.ParseResourceName(childName)
			child, _ := modconfig.GetResource(mod, parsed)
			children = append(children, child.(modconfig.ModTreeItem))
		}
		report.SetChildren(children)
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
				Include: []string{"**/*.sp"},
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
