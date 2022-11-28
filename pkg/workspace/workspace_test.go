package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/steampipe/pkg/utils"

	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// Testing the runtime dependencies(dashboards) for workspaces

type loadWorkspaceTest struct {
	skip                        bool
	source                      string
	expected                    interface{}
	expectedRuntimeDependencies map[string]map[string]*modconfig.RuntimeDependency
}

var toStringPointer = utils.ToStringPointer
var toIntegerPointer = utils.ToIntegerPointer

var testCasesLoadWorkspace = map[string]loadWorkspaceTest{
	"dashboard_runtime_deps_named_arg": { // this is to test runtime dependencies for named arguments
		skip:   true,
		source: "test_data/dashboard_runtime_deps_named_arg",
		expected: &Workspace{
			Mod: &modconfig.Mod{
				ShortName:   "dashboard_runtime_deps_named_arg",
				FullName:    "mod.dashboard_runtime_deps_named_arg",
				Require:     &modconfig.Require{},
				Description: toStringPointer("this mod is to test runtime dependencies for named arguments"),
				Title:       toStringPointer("dashboard runtime dependencies named arguments"),
				ResourceMaps: &modconfig.ResourceMaps{
					Queries: map[string]*modconfig.Query{
						"dashboard_runtime_deps_named_arg.query.query1": {
							FullName:        "dashboard_runtime_deps_named_arg.query.query1",
							ShortName:       "query1",
							UnqualifiedName: "query.query1",
							SQL:             toStringPointer("select 1 as query1"),
						},
						"dashboard_runtime_deps_named_arg.query.query2": {
							FullName:        "dashboard_runtime_deps_named_arg.query.query2",
							ShortName:       "query2",
							UnqualifiedName: "query.query2",
							SQL:             toStringPointer("select 2 as query2"),
						},
					},
					Dashboards: map[string]*modconfig.Dashboard{
						"dashboard_runtime_deps_named_arg.dashboard.dashboard_named_args": {
							FullName:        "dashboard_runtime_deps_named_arg.dashboard.dashboard_named_args",
							ShortName:       "dashboard_named_args",
							UnqualifiedName: "dashboard.dashboard_named_args",
							Title:           toStringPointer("dashboard with named arguments"),
							ChildNames:      []string{"dashboard_runtime_deps_named_arg.input.user", "dashboard_runtime_deps_named_arg.table.dashboard_dashboard_named_args_anonymous_table_0"},
							//HclType:         "dashboard",
						},
					},
					DashboardInputs: map[string]map[string]*modconfig.DashboardInput{
						"dashboard_runtime_deps_named_arg.dashboard.dashboard_named_args": {
							"dashboard_runtime_deps_named_arg.input.user": {
								FullName:        "dashboard_runtime_deps_named_arg.input.user",
								ShortName:       "user",
								UnqualifiedName: "input.user",
								DashboardName:   "dashboard_runtime_deps_named_arg.dashboard.dashboard_named_args",
								Title:           toStringPointer("AWS IAM User"),
								Width:           toIntegerPointer(4),
								SQL:             toStringPointer("select 1 as query1"),
							},
						},
					},
					DashboardTables: map[string]*modconfig.DashboardTable{
						"dashboard_runtime_deps_named_arg.table.dashboard_dashboard_named_args_anonymous_table_0": {
							FullName:        "dashboard_runtime_deps_named_arg.table.dashboard_dashboard_named_args_anonymous_table_0",
							ShortName:       "dashboard_dashboard_named_args_anonymous_table_0",
							UnqualifiedName: "table.dashboard_dashboard_named_args_anonymous_table_0",
							ColumnList: modconfig.DashboardTableColumnList{
								&modconfig.DashboardTableColumn{
									Name:    "depth",
									Display: toStringPointer("none"),
								},
							},
							Columns: map[string]*modconfig.DashboardTableColumn{
								"depth": {
									Name:    "depth",
									Display: toStringPointer("none"),
								},
							},
							Query: &modconfig.Query{
								ShortName: "query2",
								FullName:  "dashboard_runtime_deps_named_arg.query.query2",
								SQL:       toStringPointer("select 2 as query2"),
							},
							Args: &modconfig.QueryArgs{},
						},
					},
					References: map[string]*modconfig.ResourceReference{
						"To: query.query1\nFrom: input.user\nBlockType: input\nBlockName: user\nAttribute: sql": {
							To:        "query.query1",
							From:      "input.user",
							BlockType: "input",
							BlockName: "user",
							Attribute: "sql",
						},
						"To: query.query2\nFrom: table.dashboard_dashboard_named_args_anonymous_table_0\nBlockType: table\nBlockName: \nAttribute: query": {
							To:        "query.query2",
							From:      "table.dashboard_dashboard_named_args_anonymous_table_0",
							BlockType: "table",
							BlockName: "",
							Attribute: "query",
						},
						"To: self.input.user\nFrom: table.dashboard_dashboard_named_args_anonymous_table_0\nBlockType: table\nBlockName: \nAttribute: args": {
							To:        "self.input.user",
							From:      "table.dashboard_dashboard_named_args_anonymous_table_0",
							BlockType: "table",
							BlockName: "",
							Attribute: "args",
						},
					},
				},
			},
		},
		expectedRuntimeDependencies: map[string]map[string]*modconfig.RuntimeDependency{
			"dashboard_runtime_deps_named_arg.table.dashboard_dashboard_named_args_anonymous_table_0": {
				"arg.iam_user_arn->self.input.user.value": {
					PropertyPath: &modconfig.ParsedPropertyPath{
						PropertyPath: []string{"value"},
					},
					SourceResource: &modconfig.DashboardInput{
						FullName: "dashboard_runtime_deps_named_arg.input.user",
					},
				},
			},
		},
	},
	"dashboard_runtime_deps_pos_arg": { // this is to test runtime dependencies for positional arguments
		skip:   true,
		source: "test_data/dashboard_runtime_deps_pos_arg",
		expected: &Workspace{
			Mod: &modconfig.Mod{
				ShortName:   "dashboard_runtime_deps_pos_arg",
				FullName:    "mod.dashboard_runtime_deps_pos_arg",
				Require:     &modconfig.Require{},
				Description: toStringPointer("this mod is to test runtime dependencies for positional arguments"),
				Title:       toStringPointer("dashboard runtime dependencies positional arguments"),
				ResourceMaps: &modconfig.ResourceMaps{
					Queries: map[string]*modconfig.Query{
						"dashboard_runtime_deps_pos_arg.query.query1": {
							FullName:  "dashboard_runtime_deps_pos_arg.query.query1",
							ShortName: "query1",
							SQL:       toStringPointer("select 1 as query1"),
						},
						"dashboard_runtime_deps_pos_arg.query.query2": {
							FullName:  "dashboard_runtime_deps_pos_arg.query.query2",
							ShortName: "query2",
							SQL:       toStringPointer("select 2 as query2"),
						},
					},
					Dashboards: map[string]*modconfig.Dashboard{
						"dashboard_runtime_deps_pos_arg.dashboard.dashboard_pos_args": {
							FullName:        "dashboard_runtime_deps_pos_arg.dashboard.dashboard_pos_args",
							ShortName:       "dashboard_pos_args",
							UnqualifiedName: "dashboard.dashboard_pos_args",
							Title:           toStringPointer("dashboard with positional arguments"),
							ChildNames:      []string{"dashboard_runtime_deps_pos_arg.input.user", "dashboard_runtime_deps_pos_arg.table.dashboard_dashboard_pos_args_anonymous_table_0"},
							//HclType:         "dashboard",
						},
					},
					DashboardInputs: map[string]map[string]*modconfig.DashboardInput{
						"dashboard_runtime_deps_pos_arg.dashboard.dashboard_pos_args": {
							"dashboard_runtime_deps_pos_arg.input.user": {
								FullName:        "dashboard_runtime_deps_pos_arg.input.user",
								ShortName:       "user",
								UnqualifiedName: "input.user",
								DashboardName:   "dashboard_runtime_deps_pos_arg.dashboard.dashboard_pos_args",
								Title:           toStringPointer("AWS IAM User"),
								Width:           toIntegerPointer(4),
								SQL:             toStringPointer("select 1 as query1"),
							},
						},
					},
					DashboardTables: map[string]*modconfig.DashboardTable{
						"dashboard_runtime_deps_pos_arg.table.dashboard_dashboard_pos_args_anonymous_table_0": {
							FullName:        "dashboard_runtime_deps_pos_arg.table.dashboard_dashboard_pos_args_anonymous_table_0",
							ShortName:       "dashboard_dashboard_pos_args_anonymous_table_0",
							UnqualifiedName: "table.dashboard_dashboard_pos_args_anonymous_table_0",
							ColumnList: modconfig.DashboardTableColumnList{
								&modconfig.DashboardTableColumn{
									Name:    "depth",
									Display: toStringPointer("none"),
								},
							},
							Columns: map[string]*modconfig.DashboardTableColumn{
								"depth": {
									Name:    "depth",
									Display: toStringPointer("none"),
								},
							},
							Query: &modconfig.Query{
								ShortName: "query2",
								FullName:  "dashboard_runtime_deps_pos_arg.query.query2",
								SQL:       toStringPointer("select 2 as query2"),
							},
							Args: &modconfig.QueryArgs{
								ArgList: []*string{nil},
							},
						},
					},
					References: map[string]*modconfig.ResourceReference{
						"To: query.query1\nFrom: input.user\nBlockType: input\nBlockName: user\nAttribute: sql": {
							To:        "query.query1",
							From:      "input.user",
							BlockType: "input",
							BlockName: "user",
							Attribute: "sql",
						},
						"To: query.query2\nFrom: table.dashboard_dashboard_pos_args_anonymous_table_0\nBlockType: table\nBlockName: \nAttribute: query": {
							To:        "query.query2",
							From:      "table.dashboard_dashboard_pos_args_anonymous_table_0",
							BlockType: "table",
							BlockName: "",
							Attribute: "query",
						},
						"To: self.input.user\nFrom: table.dashboard_dashboard_pos_args_anonymous_table_0\nBlockType: table\nBlockName: \nAttribute: args": {
							To:        "self.input.user",
							From:      "table.dashboard_dashboard_pos_args_anonymous_table_0",
							BlockType: "table",
							BlockName: "",
							Attribute: "args",
						},
					},
				},
			},
		},
		expectedRuntimeDependencies: map[string]map[string]*modconfig.RuntimeDependency{
			"dashboard_runtime_deps_pos_arg.table.dashboard_dashboard_pos_args_anonymous_table_0": {
				"arg.0->self.input.user.value": {
					PropertyPath: &modconfig.ParsedPropertyPath{
						PropertyPath: []string{"value"},
					},
					SourceResource: &modconfig.DashboardInput{
						FullName: "dashboard_runtime_deps_pos_arg.input.user",
					},
				},
			},
		},
	},
	"dependent_mod": {
		source: "test_data/dependent_mod",
		expected: &Workspace{
			Mod: &modconfig.Mod{
				ShortName: "local",
				FullName:  "mod.local",
				Title:     toStringPointer("dependent_mod"),
				Require:   &modconfig.Require{},
				ResourceMaps: &modconfig.ResourceMaps{
					Queries: map[string]*modconfig.Query{
						"m1.query.m1_q1": {
							ShortName:       "m1_q1",
							FullName:        "m1.query.m1_q1",
							SQL:             toStringPointer("select 1 as query"),
							UnqualifiedName: "query.m1_q1",
						},
					},
					Dashboards: map[string]*modconfig.Dashboard{
						"m1.dashboard.m1_d1": {
							ShortName:       "m1_d1",
							FullName:        "m1.dashboard.m1_d1",
							UnqualifiedName: "dashboard.m1_d1",
							Title:           toStringPointer("dashboard d1"),
							ChildNames:      []string{"m1.chart.dashboard_m1_d1_anonymous_chart_0", "m1.input.i1"},
							//HclType:         "dashboard",
						},
						"local.dashboard.local_d1": {
							ShortName:       "local_d1",
							FullName:        "local.dashboard.local_d1",
							UnqualifiedName: "dashboard.local_d1",
							Title:           toStringPointer("dashboard d1"),
							ChildNames:      []string{"local.chart.dashboard_local_d1_anonymous_chart_0", "local.input.i1"},
						},
					},
					DashboardCharts: map[string]*modconfig.DashboardChart{
						"m1.chart.global_chart": {
							FullName:        "m1.chart.global_chart",
							ShortName:       "global_chart",
							UnqualifiedName: "chart.global_chart",
							Title:           toStringPointer("chart global m1_c1"),
							SQL:             toStringPointer("select 'm1_global_c1' as chart"),
						},
						"m1.chart.dashboard_m1_d1_anonymous_chart_0": {
							FullName:        "m1.chart.dashboard_m1_d1_anonymous_chart_0",
							ShortName:       "dashboard_m1_d1_anonymous_chart_0",
							UnqualifiedName: "chart.dashboard_m1_d1_anonymous_chart_0",
							Title:           toStringPointer("chart c1"),
							SQL:             toStringPointer("select 'm1_d1_c1' as chart"),
						},
						"local.chart.chart1": {
							FullName:        "local.chart.chart1",
							ShortName:       "chart1",
							UnqualifiedName: "chart.chart1",
							Title:           toStringPointer("chart 1"),
							Query: &modconfig.Query{
								FullName:  "m1.query.m1_q1",
								ShortName: "m1_q1",
								SQL:       toStringPointer("select 1 as query"),
							},
						},
						"local.chart.dashboard_local_d1_anonymous_chart_0": {
							FullName:        "local.chart.dashboard_local_d1_anonymous_chart_0",
							ShortName:       "dashboard_local_d1_anonymous_chart_0",
							UnqualifiedName: "chart.dashboard_local_d1_anonymous_chart_0",
							Title:           toStringPointer("chart global m1_c1"),
							SQL:             toStringPointer("select 'm1_global_c1' as chart"),
						},
					},
					DashboardInputs: map[string]map[string]*modconfig.DashboardInput{
						"m1.dashboard.m1_d1": {
							"m1.input.i1": {
								FullName:        "m1.input.i1",
								ShortName:       "i1",
								UnqualifiedName: "input.i1",
								DashboardName:   "m1.dashboard.m1_d1",
								Title:           toStringPointer("input i1"),
							},
						},
						"local.dashboard.local_d1": {
							"local.input.i1": {
								FullName:        "local.input.i1",
								ShortName:       "i1",
								UnqualifiedName: "input.i1",
								DashboardName:   "local.dashboard.local_d1",
								Title:           toStringPointer("global input m1_i1"),
							},
						},
					},
					GlobalDashboardInputs: map[string]*modconfig.DashboardInput{
						"m1.input.global_input": {
							FullName:        "m1.input.global_input",
							ShortName:       "global_input",
							UnqualifiedName: "input.global_input",
							DashboardName:   "",
							Title:           toStringPointer("global input m1_i1"),
						},
					},
					References: map[string]*modconfig.ResourceReference{
						"To: m1.query.m1_q1\nFrom: chart.chart1\nBlockType: chart\nBlockName: chart1\nAttribute: query": {
							To:        "m1.query.m1_q1",
							From:      "chart.chart1",
							BlockType: "chart",
							BlockName: "chart1",
							Attribute: "query",
						},
						"To: m1.chart.global_chart\nFrom: chart.dashboard_local_d1_anonymous_chart_0\nBlockType: chart\nBlockName: \nAttribute: base": {
							To:        "m1.chart.global_chart",
							From:      "chart.dashboard_local_d1_anonymous_chart_0",
							BlockType: "chart",
							BlockName: "",
							Attribute: "base",
						},
						"To: m1.input.global_input\nFrom: input.i1\nBlockType: input\nBlockName: i1\nAttribute: base": {
							To:        "m1.input.global_input",
							From:      "input.i1",
							BlockType: "input",
							BlockName: "i1",
							Attribute: "base",
						},
					},
				},
			},
		},
		expectedRuntimeDependencies: map[string]map[string]*modconfig.RuntimeDependency{},
	},
	"dependent_mod_2": {
		skip:   true,
		source: "test_data/dependent_mod_2",
		expected: &Workspace{
			Mod: &modconfig.Mod{
				ShortName: "local",
				FullName:  "mod.local",
				Title:     toStringPointer("dependent_mod"),
				Require:   &modconfig.Require{},
				ResourceMaps: &modconfig.ResourceMaps{
					Queries: map[string]*modconfig.Query{
						"m1.query.m1_q1": {
							ShortName:       "m1_q1",
							FullName:        "m1.query.m1_q1",
							SQL:             toStringPointer("select 1 as query"),
							UnqualifiedName: "query.m1_q1",
						},
					},
					Dashboards: map[string]*modconfig.Dashboard{
						"m1.dashboard.m1_d1": {
							ShortName:       "m1_d1",
							FullName:        "m1.dashboard.m1_d1",
							UnqualifiedName: "dashboard.m1_d1",
							Title:           toStringPointer("dashboard d1"),
							ChildNames:      []string{"m1.chart.dashboard_m1_d1_anonymous_chart_0", "m1.input.i1"},
							//HclType:         "dashboard",
						},
						"local.dashboard.local_d1": {
							ShortName:       "local_d1",
							FullName:        "local.dashboard.local_d1",
							UnqualifiedName: "dashboard.local_d1",
							Title:           toStringPointer("dashboard d1"),
							ChildNames:      []string{"m1.input.i1", "m1.chart.dashboard_m1_d1_anonymous_chart_0", "m1.input.i1"},
						},
					},
					DashboardCharts: map[string]*modconfig.DashboardChart{
						"m1.chart.global_chart": {
							FullName:        "m1.chart.global_chart",
							ShortName:       "global_chart",
							UnqualifiedName: "chart.global_chart",
							Title:           toStringPointer("chart global m1_c1"),
							SQL:             toStringPointer("select 'm1_global_c1' as chart"),
						},
						"m1.chart.dashboard_m1_d1_anonymous_chart_0": {
							FullName:        "m1.chart.dashboard_m1_d1_anonymous_chart_0",
							ShortName:       "dashboard_m1_d1_anonymous_chart_0",
							UnqualifiedName: "chart.dashboard_m1_d1_anonymous_chart_0",
							Title:           toStringPointer("chart c1"),
							SQL:             toStringPointer("select 'm1_d1_c1' as chart"),
						},
					},
					DashboardInputs: map[string]map[string]*modconfig.DashboardInput{
						"m1.dashboard.m1_d1": {
							"m1.input.i1": {
								FullName:        "m1.input.i1",
								ShortName:       "i1",
								UnqualifiedName: "input.i1",
								DashboardName:   "m1.dashboard.m1_d1",
								Title:           toStringPointer("input i1"),
							},
						},
						"local.dashboard.local_d1": {
							"m1.input.i1": {
								FullName:        "m1.input.i1",
								ShortName:       "i1",
								UnqualifiedName: "input.i1",
								DashboardName:   "local.dashboard.local_d1",
								Title:           toStringPointer("input i1"),
							},
						},
					},
					GlobalDashboardInputs: map[string]*modconfig.DashboardInput{
						"m1.input.global_input": {
							FullName:        "m1.input.global_input",
							ShortName:       "global_input",
							UnqualifiedName: "input.global_input",
							DashboardName:   "",
							Title:           toStringPointer("global input m1_i1"),
						},
					},
					References: map[string]*modconfig.ResourceReference{
						"To: m1.dashboard.m1_d1\nFrom: dashboard.local_d1\nBlockType: dashboard\nBlockName: local_d1\nAttribute: base": {
							To:        "m1.dashboard.m1_d1",
							From:      "dashboard.local_d1",
							BlockType: "dashboard",
							BlockName: "local_d1",
							Attribute: "base",
						},
					},
				},
			},
		},
		expectedRuntimeDependencies: map[string]map[string]*modconfig.RuntimeDependency{},
	},
}

func TestLoadWorkspace(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
		return
	}
	for name, test := range testCasesLoadWorkspace {
		if test.skip {
			continue
		}
		executeWorkspaceLoadTest(t, name, test, wd)
	}
}

func executeWorkspaceLoadTest(t *testing.T, name string, test loadWorkspaceTest, wd string) {
	workspacePath, err := filepath.Abs(test.source)
	if err != nil {
		t.Errorf("failed to build absolute config filepath from %s", test.source)
	}

	actualWorkspace, err := Load(context.Background(), workspacePath)
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

	expectedWorkspace := test.expected.(*Workspace)

	// ensure parents and children are set correctly in expected mod (this is normally done as part of decode)
	err = setChildren(expectedWorkspace.Mod)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expectedWorkspace.Mod.BuildResourceTree(nil)

	// check runtime deps
	expectedRuntimeDeps := test.expectedRuntimeDependencies
	flag := ValidateRuntimeDeps(t, actualWorkspace, expectedRuntimeDeps)
	if !flag {
		fmt.Printf("")

		t.Errorf("Test: '%s'' FAILED due to runtime dependency mismatch", name)
	}

	if !actualWorkspace.Mod.Equals(expectedWorkspace.Mod) {
		fmt.Printf("")

		t.Errorf("Test: '%s'' FAILED.", name)
	}
}

// try to resolve mod resource children using their child names
func setChildren(mod *modconfig.Mod) error {
	for _, benchmark := range mod.ResourceMaps.Benchmarks {
		for _, childName := range benchmark.ChildNames {
			parsed, _ := modconfig.ParseResourceName(childName.Name)
			child, found := modconfig.GetResource(mod, parsed)
			if !found {
				return fmt.Errorf("failed to resolve child %s", childName)
			}
			benchmark.Children = append(benchmark.Children, child.(modconfig.ModTreeItem))
		}
	}
	for _, container := range mod.ResourceMaps.DashboardContainers {
		var children []modconfig.ModTreeItem
		for _, childName := range container.ChildNames {
			parsed, _ := modconfig.ParseResourceName(childName)
			var child modconfig.HclResource
			var found bool
			if parsed.ItemType == "input" {
				child, found = modconfig.GetDashboardInput(mod, parsed.ToResourceName(), container.Name())
			} else {
				child, found = modconfig.GetResource(mod, parsed)
			}
			if !found {
				return fmt.Errorf("failed to resolve child %s", childName)
			}
			children = append(children, child.(modconfig.ModTreeItem))
		}
		container.SetChildren(children)

	}
	for _, dashboard := range mod.ResourceMaps.Dashboards {
		var children []modconfig.ModTreeItem
		for _, childName := range dashboard.ChildNames {
			parsed, _ := modconfig.ParseResourceName(childName)
			var child modconfig.HclResource
			var found bool
			if parsed.ItemType == "input" {
				child, found = modconfig.GetDashboardInput(mod, childName, dashboard.Name())
			} else {
				child, found = modconfig.GetResource(mod, parsed)
			}
			if !found {
				return fmt.Errorf("failed to resolve child %s", childName)
			}
			children = append(children, child.(modconfig.ModTreeItem))
		}
		dashboard.SetChildren(children)
	}
	return nil
}

func ValidateRuntimeDeps(t *testing.T, workspace *Workspace, expected map[string]map[string]*modconfig.RuntimeDependency) bool {
	mod := workspace.Mod
	for name, expectedDeps := range expected {
		parsedName, err := modconfig.ParseResourceName(name)
		if err != nil {
			t.Fatalf(err.Error())
		}

		resource, found := modconfig.GetResource(mod, parsedName)
		if !found {
			t.Fatalf("Resource not found")
		}

		queryProvider := resource.(modconfig.QueryProvider)
		actualRuntimeDeps := queryProvider.GetRuntimeDependencies()
		// compare the lengths
		if len(actualRuntimeDeps) != len(expected) {
			t.Fatalf("Runtime dependencies not equal")
		}

		// if actual is equal to expected
		for i, expectedDep := range expectedDeps {
			rd, notFound := actualRuntimeDeps[i]
			if !notFound {
				t.Fatalf("Runtime dependencies not equal")
			}
			if !expectedDep.Equals(rd) {
				t.Fatalf("Runtime dependencies not equal")
			}
		}
	}
	return true
}

// old code (TBD to remove)
// the actual mod loading logic is tested more thoroughly in TestLoadMod (steampipeconfig/load_mod_test.go)
// this test is primarily to verify the QueryMap building
// type loadWorkspaceTest struct {
// 	source   string
// 	expected interface{}
// }

// var toStringPointer = utils.ToStringPointer

// var testCasesLoadWorkspace = map[string]loadWorkspaceTest{
// 	"single mod": {
// 		source: "test_data/w_1",
// 		expected: &Workspace{
// 			Mod: &modconfig.Mod{
// 				ShortName: "w_1",
// 				Title:     toStringPointer("workspace 1"),
// 				//ModDepends: []*modconfig.ModVersionConstraint{
// 				//	{ShortName: "github.com/turbot/m1", Version: "0.0.0"},
// 				//	{ShortName: "github.com/turbot/m2", Version: "0.0.0"},
// 				//},
// 				Queries: map[string]*modconfig.Query{
// 					"localq1": {
// 						ShortName: "localq1", Title: toStringPointer("LocalQ1"), Description: toStringPointer("THIS IS LOCAL QUERY 1"), SQL: toStringPointer(".tables"),
// 					},
// 					"localq2": {
// 						ShortName: "localq2", Title: toStringPointer("LocalQ2"), Description: toStringPointer("THIS IS LOCAL QUERY 2"), SQL: toStringPointer(".inspect"),
// 					},
// 				},
// 			},
// 			//Queries: map[string]*modconfig.Query{
// 			//	"w_1.query.localq1": {
// 			//		ShortName: "localq1", Title: toStringPointer("LocalQ1"), Description: toStringPointer("THIS IS LOCAL QUERY 1"), SQL: toStringPointer(".tables"),
// 			//	},
// 			//	"query.localq1": {
// 			//		ShortName: "localq1", Title: toStringPointer("LocalQ1"), Description: toStringPointer("THIS IS LOCAL QUERY 1"), SQL: toStringPointer(".tables"),
// 			//	},
// 			//	"w_2.query.localq2": {
// 			//		ShortName: "localq2", Title: toStringPointer("LocalQ2"), Description: toStringPointer("THIS IS LOCAL QUERY 2"), SQL: toStringPointer(".inspect"),
// 			//	},
// 			//	"query.localq2": {
// 			//		ShortName: "localq2", Title: toStringPointer("LocalQ2"), Description: toStringPointer("THIS IS LOCAL QUERY 2"), SQL: toStringPointer(".inspect"),
// 			//	},
// 			//	"m1.query.q1": {
// 			//		ShortName: "q1", FullName: "Q1", Description: toStringPointer("THIS IS QUERY 1"), Documentation: toStringPointer("select 1"),
// 			//	},
// 			//	"m2.query.q2": {
// 			//		ShortName: "q2", FullName: "Q2", Description: toStringPointer("THIS IS QUERY 2"), Documentation: toStringPointer("select 2"),
// 			//	},
// 			//},
// 		},
// 	},
// 	"single_mod_with_ignored_directory": {
// 		source: "test_data/single_mod_with_ignored_directory",
// 		expected: &Workspace{Mod: &modconfig.Mod{
// 			ShortName:   "m1",
// 			Title:       toStringPointer("M1"),
// 			Description: toStringPointer("THIS IS M1"),
// 			Queries: map[string]*modconfig.Query{
// 				"q1": {
// 					ShortName: "q1", FullName: "Q1", Description: toStringPointer("THIS IS QUERY 1"), Documentation: toStringPointer("select 1"),
// 				},
// 				"q2": {
// 					ShortName: "q2", FullName: "Q2", Description: toStringPointer("THIS IS QUERY 2"), Documentation: toStringPointer("select 2"),
// 				},
// 			},
// 		},
// 		},
// 	},
// 	"single_mod_with_ignored_sql_files": {
// 		source: "test_data/single_mod_with_ignored_sql_files",
// 		expected: &Workspace{Mod: &modconfig.Mod{
// 			ShortName:   "m1",
// 			Title:       toStringPointer("M1"),
// 			Description: toStringPointer("THIS IS M1"),
// 			Queries: map[string]*modconfig.Query{
// 				"q1": {
// 					ShortName: "q1", FullName: "Q1", Description: toStringPointer("THIS IS QUERY 1"), Documentation: toStringPointer("select 1"),
// 				},
// 			},
// 		}},
// 	},
// 	"single_mod_in_hidden_folder": {
// 		source: "test_data/.hidden/w_1",
// 		expected: &Workspace{
// 			Mod: &modconfig.Mod{
// 				ShortName: "w_1",
// 				Title:     toStringPointer("workspace 1"),
// 				Queries: map[string]*modconfig.Query{
// 					"localq1": {
// 						ShortName: "localq1", Title: toStringPointer("LocalQ1"), Description: toStringPointer("THIS IS LOCAL QUERY 1"), SQL: toStringPointer(".tables"),
// 					},
// 					"localq2": {
// 						ShortName: "localq2", Title: toStringPointer("LocalQ2"), Description: toStringPointer("THIS IS LOCAL QUERY 2"), SQL: toStringPointer(".inspect"),
// 					},
// 				},
// 			},
// 			//Queries: map[string]*modconfig.Query{
// 			//	"w_1.query.localq1": {
// 			//		ShortName: "localq1", Title: toStringPointer("LocalQ1"), Description: toStringPointer("THIS IS LOCAL QUERY 1"), SQL: toStringPointer(".tables"),
// 			//	},
// 			//	"query.localq1": {
// 			//		ShortName: "localq1", Title: toStringPointer("LocalQ1"), Description: toStringPointer("THIS IS LOCAL QUERY 1"), SQL: toStringPointer(".tables"),
// 			//	},
// 			//	"w_2.query.localq2": {
// 			//		ShortName: "localq2", Title: toStringPointer("LocalQ2"), Description: toStringPointer("THIS IS LOCAL QUERY 2"), SQL: toStringPointer(".inspect"),
// 			//	},
// 			//	"query.localq2": {
// 			//		ShortName: "localq2", Title: toStringPointer("LocalQ2"), Description: toStringPointer("THIS IS LOCAL QUERY 2"), SQL: toStringPointer(".inspect"),
// 			//	},
// 			//	"m1.query.q1": {
// 			//		ShortName: "q1", FullName: "Q1", Description: toStringPointer("THIS IS QUERY 1"), Documentation: toStringPointer("select 1"),
// 			//	},
// 			//	"m2.query.q2": {
// 			//		ShortName: "q2", FullName: "Q2", Description: toStringPointer("THIS IS QUERY 2"), Documentation: toStringPointer("select 2"),
// 			//	},
// 			//},
// 		},
// 	},
// }

// func TestLoadWorkspace(t *testing.T) {
// 	for name, test := range testCasesLoadWorkspace {
// 		workspacePath, err := filepath.Abs(test.source)
// 		workspace, err := Load(context.Background(), workspacePath)

// 		if err != nil {
// 			if test.expected != "ERROR" {
// 				t.Errorf("Test: '%s'' FAILED with unexpected error: %v", name, err)
// 			}
// 			continue
// 		}

// 		if test.expected == "ERROR" {
// 			t.Errorf("Test: '%s'' FAILED - expected error", name)
// 			continue
// 		}

// 		if match, message := WorkspacesEqual(test.expected.(*Workspace), workspace); !match {
// 			t.Errorf("Test: '%s'' FAILED : %s", name, message)
// 		}
// 	}
// }

// func WorkspacesEqual(expected, actual *Workspace) (bool, string) {

// 	errors := []string{}
// 	if actual.Mod.String() != expected.Mod.String() {
// 		errors = append(errors, fmt.Sprintf("workspace mods do not match - expected \n\n%s\n\nbut got\n\n%s\n", expected.Mod.String(), actual.Mod.String()))
// 	}
// 	expectedMaps := expected.GetResourceMaps()
// 	actualMaps := actual.GetResourceMaps()

// 	for name, expectedQuery := range expectedMaps.Queries {
// 		actualQuery, ok := actualMaps.Queries[name]
// 		if ok {
// 			if expectedQuery.String() != actualQuery.String() {
// 				errors = append(errors, fmt.Sprintf("query %s expected\n\n%s\n\n, got\na\n%s\n\n", name, expectedQuery.String(), actualQuery.String()))
// 			}
// 		} else {
// 			errors = append(errors, fmt.Sprintf("mod map missing expected key %s", name))
// 		}
// 	}
// 	for name := range actualMaps.Queries {
// 		if _, ok := expectedMaps.Queries[name]; ok {
// 			errors = append(errors, fmt.Sprintf("unexpected query %s in query map", name))
// 		}
// 	}
// 	return len(errors) > 0, strings.Join(errors, "\n")
// }
