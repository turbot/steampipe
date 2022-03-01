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
var toIntegerPointer = utils.ToIntegerPointer

type loadModTest struct {
	source   string
	expected interface{}
}

var testCasesLoadMod map[string]loadModTest

func init() {
	filepaths.SteampipeDir = "~/.steampipe"
	require, _ := modconfig.NewRequire()
	testCasesLoadMod = map[string]loadModTest{
		"no_mod_sql_files": {
			source: "testdata/mods/no_mod_sql_files",
			expected: &modconfig.Mod{
				ShortName: "local",
				FullName:  "mod.local",
				Require:   require,
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
				},
			},
		},
		"no_mod_hcl_queries": {
			source: "testdata/mods/no_mod_hcl_queries",
			expected: &modconfig.Mod{
				ShortName: "local",
				Title:     toStringPointer("no_mod_hcl_queries"),
				FullName:  "mod.local",
				Require:   require,
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
				Require:     require,
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
			},
		},
		"single_mod_one_query": {
			source: "testdata/mods/single_mod_one_query",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Require:     require,
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
				Require:     require,
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
						ShortName:   "q1",
						FullName:    "m1.query.q1",
						Title:       toStringPointer("Q1"),
						Description: toStringPointer("THIS IS QUERY 1"),
						SQL:         toStringPointer("select $1"),
						Params: []*modconfig.ParamDef{
							{
								Name:        "p1",
								FullName:    "param.p1",
								Description: utils.ToStringPointer("desc"),
								Default:     utils.ToStringPointer("'I am default'"),
							},
						},
					},
				},
			},
		},
		// "query_with_paramdefs_control_with_named_params": {
		// 	source: "testdata/mods/query_with_paramdefs_control_with_named_params",
		// 	expected: &modconfig.Mod{
		// 		ShortName:   "m1",
		// 		FullName:    "mod.m1",
		// 		Require:     require,
		// 		Title:       toStringPointer("M1"),
		// 		Description: toStringPointer("THIS IS M1"),
		// 		Queries: map[string]*modconfig.Query{
		// 			"m1.query.q1": {
		// 				ShortName:   "q1",
		// 				FullName:    "m1.query.q1",
		// 				Title:       toStringPointer("Q1"),
		// 				Description: toStringPointer("THIS IS QUERY 1"),
		// 				SQL:         toStringPointer("select 1"),
		// 				Params: []*modconfig.ParamDef{
		// 					{
		// 						Name:        "p1",
		// 						FullName:    "param.p1",
		// 						Description: utils.ToStringPointer("desc"),
		// 						Default:     utils.ToStringPointer("'I am default'"),
		// 					},
		// 					{
		// 						Name:        "p2",
		// 						FullName:    "param.p2",
		// 						Description: utils.ToStringPointer("desc 2"),
		// 						Default:     utils.ToStringPointer("'I am default 2'"),
		// 					},
		// 				},
		// 			},
		// 		},
		// 		Controls: map[string]*modconfig.Control{
		// 			"m1.control.c1": {
		// 				ShortName:   "c1",
		// 				FullName:    "m1.control.c1",
		// 				Title:       toStringPointer("C1"),
		// 				Description: toStringPointer("THIS IS CONTROL 1"),
		// 				SQL:         toStringPointer("select 'ok' as status, 'foo' as resource, 'bar' as reason"),
		// 				Params: []*modconfig.ParamDef{
		// 					{
		// 						Name:     "p1",
		// 						FullName: "param.p1",
		// 						Default:  utils.ToStringPointer("'val1'"),
		// 					},
		// 					{
		// 						Name:     "p2",
		// 						FullName: "param.p2",
		// 						Default:  utils.ToStringPointer("'val2'"),
		// 					},
		// 				},
		// 				Args: &modconfig.QueryArgs{
		// 					ArgMap:  map[string]string{},
		// 					ArgList: []*string{utils.ToStringPointer("'my val1'"), utils.ToStringPointer("'my val2'")}},
		// 			},
		// 		},
		// 	},
		// },
		"single_mod_one_query_one_control": {
			source: "testdata/mods/single_mod_one_query_one_control",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Require:     require,
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
				Require:     require,
				Title:       toStringPointer("M1"),
				Description: toStringPointer("THIS IS M1"),
				Queries: map[string]*modconfig.Query{
					"m1.query.q1": {
						ShortName:       "q1",
						FullName:        "m1.query.q1",
						SQL:             toStringPointer("select 1"),
						UnqualifiedName: "query.q1",
					},
				},
				Controls: map[string]*modconfig.Control{
					"m1.control.c1": {
						ShortName:       "c1",
						FullName:        "m1.control.c1",
						SQL:             toStringPointer("select 'pass' as result"),
						Args:            &modconfig.QueryArgs{},
						UnqualifiedName: "control.c1",
					},
					"m1.control.c2": {
						ShortName:       "c2",
						FullName:        "m1.control.c2",
						SQL:             toStringPointer("select 'pass' as result"),
						Args:            &modconfig.QueryArgs{},
						UnqualifiedName: "control.c2",
					},
					"m1.control.c3": {
						ShortName:       "c3",
						FullName:        "m1.control.c3",
						SQL:             toStringPointer("select 'pass' as result"),
						Args:            &modconfig.QueryArgs{},
						UnqualifiedName: "control.c3",
					},
					"m1.control.c4": {
						ShortName:       "c4",
						FullName:        "m1.control.c4",
						SQL:             toStringPointer("select 'pass' as result"),
						Args:            &modconfig.QueryArgs{},
						UnqualifiedName: "control.c4",
					},
					"m1.control.c5": {
						ShortName:       "c5",
						FullName:        "m1.control.c5",
						SQL:             toStringPointer("select 'pass' as result"),
						Args:            &modconfig.QueryArgs{},
						UnqualifiedName: "control.c5",
					},
					"m1.control.c6": {
						ShortName:       "c6",
						FullName:        "m1.control.c6",
						SQL:             toStringPointer("select 'fail' as result"),
						Args:            &modconfig.QueryArgs{},
						UnqualifiedName: "control.c6",
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
				Require:     require,
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
				Require:     require,
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
		// "single_mod_sql_file_and_clashing_hcl_query": {
		// 	source:   "testdata/mods/single_mod_sql_file_and_clashing_hcl_query",
		// 	expected: "ERROR",
		// },
		"single_mod_two_queries_diff_files": {
			source: "testdata/mods/single_mod_two_queries_diff_files",
			expected: &modconfig.Mod{
				ShortName:   "m1",
				FullName:    "mod.m1",
				Require:     require,
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
				Require:     require,
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
				Require:     require,
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
		"dashboard_simple_report": {
			source: "testdata/mods/dashboard_simple_report",
			expected: &modconfig.Mod{
				ShortName:   "simple_report",
				FullName:    "mod.simple_report",
				Require:     require,
				Title:       toStringPointer("simple report"),
				Description: toStringPointer("this mod contains a simple report"),
				Dashboards: map[string]*modconfig.Dashboard{
					"simple_report.dashboard.simple_report": {
						ShortName:       "simple_report",
						FullName:        "simple_report.dashboard.simple_report",
						UnqualifiedName: "dashboard.simple_report",
						ChildNames:      []string{"simple_report.text.dashboard_simple_report_anonymous_text_0", "simple_report.chart.dashboard_simple_report_anonymous_chart_0"},
						HclType:         "dashboard",
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"simple_report.chart.dashboard_simple_report_anonymous_chart_0": {
						FullName:        "simple_report.chart.dashboard_simple_report_anonymous_chart_0",
						ShortName:       "dashboard_simple_report_anonymous_chart_0",
						UnqualifiedName: "chart.dashboard_simple_report_anonymous_chart_0",
						Title:           toStringPointer("a simple query"),
						SQL:             toStringPointer("select 1"),
					},
				},
				DashboardTexts: map[string]*modconfig.DashboardText{
					"simple_report.text.dashboard_simple_report_anonymous_text_0": {
						FullName:        "simple_report.text.dashboard_simple_report_anonymous_text_0",
						ShortName:       "dashboard_simple_report_anonymous_text_0",
						UnqualifiedName: "text.dashboard_simple_report_anonymous_text_0",
						Value:           toStringPointer("a simple report"),
					},
				},
			},
		},
		"dashboard_simple_container": {
			source: "testdata/mods/dashboard_simple_container",
			expected: &modconfig.Mod{
				ShortName:   "simple_container_report",
				FullName:    "mod.simple_container_report",
				Require:     require,
				Title:       toStringPointer("simple report with container"),
				Description: toStringPointer("this mod contains a simple report with containers"),
				Dashboards: map[string]*modconfig.Dashboard{
					"simple_container_report.dashboard.simple_container_report": {
						ShortName:       "simple_container_report",
						FullName:        "simple_container_report.dashboard.simple_container_report",
						UnqualifiedName: "dashboard.simple_container_report",
						ChildNames:      []string{"simple_container_report.container.dashboard_simple_container_report_anonymous_container_0"},
						HclType:         "dashboard",
					},
				},
				DashboardContainers: map[string]*modconfig.DashboardContainer{
					"simple_container_report.container.dashboard_simple_container_report_anonymous_container_0": {
						ShortName:       "dashboard_simple_container_report_anonymous_container_0",
						FullName:        "simple_container_report.container.dashboard_simple_container_report_anonymous_container_0",
						UnqualifiedName: "container.dashboard_simple_container_report_anonymous_container_0",
						ChildNames:      []string{"simple_container_report.text.container_dashboard_simple_container_report_anonymous_container_0_anonymous_text_0", "simple_container_report.chart.container_dashboard_simple_container_report_anonymous_container_0_anonymous_chart_0"},
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"simple_container_report.chart.container_dashboard_simple_container_report_anonymous_container_0_anonymous_chart_0": {
						ShortName:       "container_dashboard_simple_container_report_anonymous_container_0_anonymous_chart_0",
						FullName:        "simple_container_report.chart.container_dashboard_simple_container_report_anonymous_container_0_anonymous_chart_0",
						UnqualifiedName: "chart.container_dashboard_simple_container_report_anonymous_container_0_anonymous_chart_0",
						Title:           toStringPointer("container 1 chart 1"),
						SQL:             toStringPointer("select 1 as container"),
					},
				},
				DashboardTexts: map[string]*modconfig.DashboardText{
					"simple_container_report.text.container_dashboard_simple_container_report_anonymous_container_0_anonymous_text_0": {
						ShortName:       "container_dashboard_simple_container_report_anonymous_container_0_anonymous_text_0",
						FullName:        "simple_container_report.text.container_dashboard_simple_container_report_anonymous_container_0_anonymous_text_0",
						UnqualifiedName: "text.container_dashboard_simple_container_report_anonymous_container_0_anonymous_text_0",
						Value:           toStringPointer("container 1"),
					},
				},
			},
		},
		"dashboard_sibling_containers": {
			source: "testdata/mods/dashboard_sibling_containers",
			expected: &modconfig.Mod{
				ShortName:   "sibling_containers_report",
				FullName:    "mod.sibling_containers_report",
				Require:     require,
				Title:       toStringPointer("report with multiple sibling containers"),
				Description: toStringPointer("this mod contains a report with multiple sibling containers"),
				Dashboards: map[string]*modconfig.Dashboard{
					"sibling_containers_report.dashboard.sibling_containers_report": {
						ShortName:       "sibling_containers_report",
						FullName:        "sibling_containers_report.dashboard.sibling_containers_report",
						UnqualifiedName: "dashboard.sibling_containers_report",
						ChildNames:      []string{"sibling_containers_report.container.dashboard_sibling_containers_report_anonymous_container_0", "sibling_containers_report.container.dashboard_sibling_containers_report_anonymous_container_1", "sibling_containers_report.container.dashboard_sibling_containers_report_anonymous_container_2"},
						HclType:         "dashboard",
					},
				},
				DashboardContainers: map[string]*modconfig.DashboardContainer{
					"sibling_containers_report.container.dashboard_sibling_containers_report_anonymous_container_0": {
						ShortName:       "dashboard_sibling_containers_report_anonymous_container_0",
						FullName:        "sibling_containers_report.container.dashboard_sibling_containers_report_anonymous_container_0",
						UnqualifiedName: "container.dashboard_sibling_containers_report_anonymous_container_0",
						ChildNames:      []string{"sibling_containers_report.text.container_dashboard_sibling_containers_report_anonymous_container_0_anonymous_text_0", "sibling_containers_report.chart.container_dashboard_sibling_containers_report_anonymous_container_0_anonymous_chart_0"},
					},
					"sibling_containers_report.container.dashboard_sibling_containers_report_anonymous_container_1": {
						ShortName:       "dashboard_sibling_containers_report_anonymous_container_1",
						FullName:        "sibling_containers_report.container.dashboard_sibling_containers_report_anonymous_container_1",
						UnqualifiedName: "container.dashboard_sibling_containers_report_anonymous_container_1",
						ChildNames:      []string{"sibling_containers_report.text.container_dashboard_sibling_containers_report_anonymous_container_1_anonymous_text_0", "sibling_containers_report.chart.container_dashboard_sibling_containers_report_anonymous_container_1_anonymous_chart_0"},
					},
					"sibling_containers_report.container.dashboard_sibling_containers_report_anonymous_container_2": {
						ShortName:       "dashboard_sibling_containers_report_anonymous_container_2",
						FullName:        "sibling_containers_report.container.dashboard_sibling_containers_report_anonymous_container_2",
						UnqualifiedName: "container.dashboard_sibling_containers_report_anonymous_container_2",
						ChildNames:      []string{"sibling_containers_report.text.container_dashboard_sibling_containers_report_anonymous_container_2_anonymous_text_0", "sibling_containers_report.chart.container_dashboard_sibling_containers_report_anonymous_container_2_anonymous_chart_0"},
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"sibling_containers_report.chart.container_dashboard_sibling_containers_report_anonymous_container_0_anonymous_chart_0": {
						FullName:        "sibling_containers_report.chart.container_dashboard_sibling_containers_report_anonymous_container_0_anonymous_chart_0",
						ShortName:       "container_dashboard_sibling_containers_report_anonymous_container_0_anonymous_chart_0",
						UnqualifiedName: "chart.container_dashboard_sibling_containers_report_anonymous_container_0_anonymous_chart_0",
						Title:           toStringPointer("container 1 chart 1"),
						SQL:             toStringPointer("select 1 as container"),
					},
					"sibling_containers_report.chart.container_dashboard_sibling_containers_report_anonymous_container_1_anonymous_chart_0": {
						FullName:        "sibling_containers_report.chart.container_dashboard_sibling_containers_report_anonymous_container_1_anonymous_chart_0",
						ShortName:       "container_dashboard_sibling_containers_report_anonymous_container_1_anonymous_chart_0",
						UnqualifiedName: "chart.container_dashboard_sibling_containers_report_anonymous_container_1_anonymous_chart_0",
						Title:           toStringPointer("container 2 chart 1"),
						SQL:             toStringPointer("select 2 as container"),
					},
					"sibling_containers_report.chart.container_dashboard_sibling_containers_report_anonymous_container_2_anonymous_chart_0": {
						FullName:        "sibling_containers_report.chart.container_dashboard_sibling_containers_report_anonymous_container_2_anonymous_chart_0",
						ShortName:       "container_dashboard_sibling_containers_report_anonymous_container_2_anonymous_chart_0",
						UnqualifiedName: "chart.container_dashboard_sibling_containers_report_anonymous_container_2_anonymous_chart_0",
						Title:           toStringPointer("container 3 chart 1"),
						SQL:             toStringPointer("select 3 as container"),
					},
				},
				DashboardTexts: map[string]*modconfig.DashboardText{
					"sibling_containers_report.text.container_dashboard_sibling_containers_report_anonymous_container_0_anonymous_text_0": {
						FullName:        "sibling_containers_report.text.container_dashboard_sibling_containers_report_anonymous_container_0_anonymous_text_0",
						ShortName:       "container_dashboard_sibling_containers_report_anonymous_container_0_anonymous_text_0",
						UnqualifiedName: "text.container_dashboard_sibling_containers_report_anonymous_container_0_anonymous_text_0",
						Value:           toStringPointer("container 1"),
					},
					"sibling_containers_report.text.container_dashboard_sibling_containers_report_anonymous_container_1_anonymous_text_0": {
						FullName:        "sibling_containers_report.text.container_dashboard_sibling_containers_report_anonymous_container_1_anonymous_text_0",
						ShortName:       "container_dashboard_sibling_containers_report_anonymous_container_1_anonymous_text_0",
						UnqualifiedName: "text.container_dashboard_sibling_containers_report_anonymous_container_1_anonymous_text_0",
						Value:           toStringPointer("container 2"),
					},
					"sibling_containers_report.text.container_dashboard_sibling_containers_report_anonymous_container_2_anonymous_text_0": {
						FullName:        "sibling_containers_report.text.container_dashboard_sibling_containers_report_anonymous_container_2_anonymous_text_0",
						ShortName:       "container_dashboard_sibling_containers_report_anonymous_container_2_anonymous_text_0",
						UnqualifiedName: "text.container_dashboard_sibling_containers_report_anonymous_container_2_anonymous_text_0",
						Value:           toStringPointer("container 3"),
					},
				},
			},
		},
		"dashboard_nested_containers": {
			source: "testdata/mods/dashboard_nested_containers",
			expected: &modconfig.Mod{
				ShortName:   "nested_containers_report",
				FullName:    "mod.nested_containers_report",
				Require:     require,
				Title:       toStringPointer("report with nested containers"),
				Description: toStringPointer("this mod contains a report with nested containers"),
				Dashboards: map[string]*modconfig.Dashboard{
					"nested_containers_report.dashboard.nested_containers_report": {
						ShortName:       "nested_containers_report",
						FullName:        "nested_containers_report.dashboard.nested_containers_report",
						UnqualifiedName: "dashboard.nested_containers_report",
						ChildNames:      []string{"nested_containers_report.container.dashboard_nested_containers_report_anonymous_container_0"},
						HclType:         "dashboard",
					},
				},
				DashboardContainers: map[string]*modconfig.DashboardContainer{
					"nested_containers_report.container.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0": {
						ShortName:       "container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0",
						FullName:        "nested_containers_report.container.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0",
						UnqualifiedName: "container.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0",
						ChildNames:      []string{"nested_containers_report.text.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0_anonymous_text_0", "nested_containers_report.chart.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0_anonymous_chart_0"},
					},
					"nested_containers_report.container.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0": {
						ShortName:       "container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0",
						FullName:        "nested_containers_report.container.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0",
						UnqualifiedName: "container.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0",
						ChildNames:      []string{"nested_containers_report.text.container_container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0_anonymous_text_0", "nested_containers_report.chart.container_container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0_anonymous_chart_0"},
					},
					"nested_containers_report.container.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1": {
						ShortName:       "container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1",
						FullName:        "nested_containers_report.container.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1",
						UnqualifiedName: "container.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1",
						ChildNames:      []string{"nested_containers_report.text.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_text_0", "nested_containers_report.chart.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_chart_0", "nested_containers_report.container.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0"},
					},
					"nested_containers_report.container.dashboard_nested_containers_report_anonymous_container_0": {
						ShortName:       "dashboard_nested_containers_report_anonymous_container_0",
						FullName:        "nested_containers_report.container.dashboard_nested_containers_report_anonymous_container_0",
						UnqualifiedName: "container.dashboard_nested_containers_report_anonymous_container_0",
						ChildNames:      []string{"nested_containers_report.text.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_text_0", "nested_containers_report.container.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0", "nested_containers_report.container.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1"},
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"nested_containers_report.chart.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0_anonymous_chart_0": {
						FullName:        "nested_containers_report.chart.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0_anonymous_chart_0",
						ShortName:       "container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0_anonymous_chart_0",
						UnqualifiedName: "chart.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0_anonymous_chart_0",
						Title:           toStringPointer("CHART 1"),
						SQL:             toStringPointer("select 1.1 as container"),
					},
					"nested_containers_report.chart.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_chart_0": {
						FullName:        "nested_containers_report.chart.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_chart_0",
						ShortName:       "container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_chart_0",
						UnqualifiedName: "chart.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_chart_0",
						Title:           toStringPointer("CHART 2"),
						SQL:             toStringPointer("select 1.2 as container"),
					},
					"nested_containers_report.chart.container_container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0_anonymous_chart_0": {
						FullName:        "nested_containers_report.chart.container_container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0_anonymous_chart_0",
						ShortName:       "container_container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0_anonymous_chart_0",
						UnqualifiedName: "chart.container_container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0_anonymous_chart_0",
						Title:           toStringPointer("CHART 3"),
						SQL:             toStringPointer("select 1.2.1 as container"),
					},
				},
				DashboardTexts: map[string]*modconfig.DashboardText{
					"nested_containers_report.text.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_text_0": {
						FullName:        "nested_containers_report.text.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_text_0",
						ShortName:       "container_dashboard_nested_containers_report_anonymous_container_0_anonymous_text_0",
						UnqualifiedName: "text.container_dashboard_nested_containers_report_anonymous_container_0_anonymous_text_0",
						Value:           toStringPointer("CONTAINER 1"),
					},
					"nested_containers_report.text.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0_anonymous_text_0": {
						FullName:        "nested_containers_report.text.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0_anonymous_text_0",
						ShortName:       "container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0_anonymous_text_0",
						UnqualifiedName: "text.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_0_anonymous_text_0",
						Value:           toStringPointer("CHILD CONTAINER 1.1"),
					},
					"nested_containers_report.text.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_text_0": {
						FullName:        "nested_containers_report.text.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_text_0",
						ShortName:       "container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_text_0",
						UnqualifiedName: "text.container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_text_0",
						Value:           toStringPointer("CHILD CONTAINER 1.2"),
					},
					"nested_containers_report.text.container_container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0_anonymous_text_0": {
						FullName:        "nested_containers_report.text.container_container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0_anonymous_text_0",
						ShortName:       "container_container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0_anonymous_text_0",
						UnqualifiedName: "text.container_container_container_dashboard_nested_containers_report_anonymous_container_0_anonymous_container_1_anonymous_container_0_anonymous_text_0",
						Value:           toStringPointer("NESTED CHILD CONTAINER 1.2.1"),
					},
				},
			},
		},
		"dashboard_base_override": { // this test checks the base values overriding while parsing
			source: "testdata/mods/dashboard_base_override",
			expected: &modconfig.Mod{
				ShortName:   "report_axes",
				FullName:    "mod.report_axes",
				Require:     require,
				Title:       toStringPointer("report with axes"),
				Description: toStringPointer("This mod tests base values overriding functionality"),
				Dashboards: map[string]*modconfig.Dashboard{
					"report_axes.dashboard.override_base_values": {
						ShortName:       "override_base_values",
						FullName:        "report_axes.dashboard.override_base_values",
						UnqualifiedName: "dashboard.override_base_values",
						Title:           toStringPointer("override_base_values"),
						ChildNames:      []string{"report_axes.chart.dashboard_override_base_values_anonymous_chart_0"},
						HclType:         "dashboard",
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"report_axes.chart.aws_bucket_info": {
						FullName:        "report_axes.chart.aws_bucket_info",
						ShortName:       "aws_bucket_info",
						UnqualifiedName: "chart.aws_bucket_info",
						Axes: &modconfig.DashboardChartAxes{
							X: &modconfig.DashboardChartAxesX{
								Title: &modconfig.DashboardChartAxisTitle{
									Display: toStringPointer("always"),
									Value:   toStringPointer("Foo"),
								},
							},
							Y: &modconfig.DashboardChartAxesY{
								Title: &modconfig.DashboardChartAxisTitle{
									Display: toStringPointer("always"),
									Value:   toStringPointer("Foo"),
								},
							},
						},
						Grouping: toStringPointer("compare"),
						Type:     toStringPointer("column"),
						Legend: &modconfig.DashboardChartLegend{
							Position: toStringPointer("bottom"),
						},
					},
					"report_axes.chart.dashboard_override_base_values_anonymous_chart_0": {
						FullName:        "report_axes.chart.dashboard_override_base_values_anonymous_chart_0",
						ShortName:       "dashboard_override_base_values_anonymous_chart_0",
						UnqualifiedName: "chart.dashboard_override_base_values_anonymous_chart_0",
						Axes: &modconfig.DashboardChartAxes{
							X: &modconfig.DashboardChartAxesX{
								Title: &modconfig.DashboardChartAxisTitle{
									Display: toStringPointer("always"),
									Value:   toStringPointer("OVERRIDE"),
								},
							},
							Y: &modconfig.DashboardChartAxesY{
								Title: &modconfig.DashboardChartAxisTitle{
									Display: toStringPointer("OVERRIDE"),
									Value:   toStringPointer("Foo"),
								},
							},
						},
						Grouping: toStringPointer("compare"),
						Type:     toStringPointer("column"),
						Legend: &modconfig.DashboardChartLegend{
							Position: toStringPointer("bottom"),
						},
					},
				},
			},
		},
		"dashboard_base_inheritance": { // this test checks inheriting and overriding base values while parsing
			source: "testdata/mods/dashboard_base_inheritance",
			expected: &modconfig.Mod{
				ShortName:   "report_base1",
				FullName:    "mod.report_base1",
				Require:     require,
				Description: toStringPointer("This mod tests inheriting from base functionality"),
				Title:       toStringPointer("report base 1"),
				Queries: map[string]*modconfig.Query{
					"report_base1.query.basic_query": {
						ShortName:       "basic_query",
						FullName:        "report_base1.query.basic_query",
						UnqualifiedName: "query.basic_query",
						SQL:             toStringPointer("select 1"),
					},
				},
				Dashboards: map[string]*modconfig.Dashboard{
					"report_base1.dashboard.inheriting_from_base": {
						ShortName:       "inheriting_from_base",
						FullName:        "report_base1.dashboard.inheriting_from_base",
						UnqualifiedName: "dashboard.inheriting_from_base",
						Title:           toStringPointer("inheriting_from_base"),
						ChildNames:      []string{"report_base1.chart.dashboard_inheriting_from_base_anonymous_chart_0"},
						HclType:         "dashboard",
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"report_base1.chart.basic_chart": {
						FullName:        "report_base1.chart.basic_chart",
						ShortName:       "basic_chart",
						UnqualifiedName: "chart.basic_chart",
						Type:            toStringPointer("column"),
						Legend: &modconfig.DashboardChartLegend{
							Position: toStringPointer("bottom"),
						},
						Axes: &modconfig.DashboardChartAxes{
							X: &modconfig.DashboardChartAxesX{
								Title: &modconfig.DashboardChartAxisTitle{
									Display: toStringPointer("always"),
									Value:   toStringPointer("Foo"),
								},
							},
							Y: &modconfig.DashboardChartAxesY{
								Title: &modconfig.DashboardChartAxisTitle{
									Display: toStringPointer("always"),
									Value:   toStringPointer("Foo"),
								},
							},
						},
						Grouping: toStringPointer("compare"),
						SQL:      toStringPointer("select 1"),
					},
					"report_base1.chart.dashboard_inheriting_from_base_anonymous_chart_0": {
						FullName:        "report_base1.chart.dashboard_inheriting_from_base_anonymous_chart_0",
						ShortName:       "dashboard_inheriting_from_base_anonymous_chart_0",
						UnqualifiedName: "chart.dashboard_inheriting_from_base_anonymous_chart_0",
						Width:           toIntegerPointer(8),
						Type:            toStringPointer("column"),
						Legend: &modconfig.DashboardChartLegend{
							Position: toStringPointer("bottom"),
						},
						Axes: &modconfig.DashboardChartAxes{
							X: &modconfig.DashboardChartAxesX{
								Title: &modconfig.DashboardChartAxisTitle{
									Display: toStringPointer("always"),
									Value:   toStringPointer("Barz"),
								},
							},
							Y: &modconfig.DashboardChartAxesY{
								Title: &modconfig.DashboardChartAxisTitle{
									Display: toStringPointer("always"),
									Value:   toStringPointer("Foo"),
								},
							},
						},
						Grouping: toStringPointer("compare"),
						SQL:      toStringPointer("select 1"),
					},
				},
			},
		},
		"dashboard_container_with_all_children": { // to test parsing of a container with all possible children
			source: "testdata/mods/dashboard_container_with_all_children",
			expected: &modconfig.Mod{
				ShortName:   "container_with_children",
				FullName:    "mod.container_with_children",
				Require:     require,
				Description: toStringPointer("This mod contains a dashboard with a container with all possible child resources"),
				Title:       toStringPointer("container with all possible child resources"),
				Dashboards: map[string]*modconfig.Dashboard{
					"container_with_children.dashboard.container_with_child_res": {
						ShortName:       "container_with_child_res",
						FullName:        "container_with_children.dashboard.container_with_child_res",
						UnqualifiedName: "dashboard.container_with_child_res",
						Title:           toStringPointer("container with child resources"),
						ChildNames:      []string{"container_with_children.container.dashboard_container_with_child_res_anonymous_container_0"},
						HclType:         "dashboard",
					},
				},
				DashboardContainers: map[string]*modconfig.DashboardContainer{
					"container_with_children.container.dashboard_container_with_child_res_anonymous_container_0": {
						ShortName:       "dashboard_container_with_child_res_anonymous_container_0",
						FullName:        "container_with_children.container.dashboard_container_with_child_res_anonymous_container_0",
						UnqualifiedName: "container.dashboard_container_with_child_res_anonymous_container_0",
						Title:           toStringPointer("example container with all possible child resources"),
						ChildNames:      []string{"container_with_children.chart.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_chart_0", "container_with_children.card.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_card_0", "container_with_children.hierarchy.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_hierarchy_0", "container_with_children.image.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_image_0", "container_with_children.table.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_table_0", "container_with_children.text.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_text_0"},
					},
				},
				DashboardCards: map[string]*modconfig.DashboardCard{
					"container_with_children.card.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_card_0": {
						FullName:        "container_with_children.card.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_card_0",
						ShortName:       "container_dashboard_container_with_child_res_anonymous_container_0_anonymous_card_0",
						UnqualifiedName: "card.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_card_0",
						Title:           toStringPointer("example card"),
						Type:            toStringPointer("ok"),
						SQL:             toStringPointer("select 1"),
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"container_with_children.chart.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_chart_0": {
						FullName:        "container_with_children.chart.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_chart_0",
						ShortName:       "container_dashboard_container_with_child_res_anonymous_container_0_anonymous_chart_0",
						UnqualifiedName: "chart.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_chart_0",
						Title:           toStringPointer("example chart"),
						SQL:             toStringPointer("select 1"),
					},
				},
				DashboardHierarchies: map[string]*modconfig.DashboardHierarchy{
					"container_with_children.hierarchy.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_hierarchy_0": {
						FullName:        "container_with_children.hierarchy.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_hierarchy_0",
						ShortName:       "container_dashboard_container_with_child_res_anonymous_container_0_anonymous_hierarchy_0",
						UnqualifiedName: "hierarchy.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_hierarchy_0",
						Title:           toStringPointer("example hierarchy"),
						Type:            toStringPointer("graph"),
					},
				},
				DashboardImages: map[string]*modconfig.DashboardImage{
					"container_with_children.image.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_image_0": {
						FullName:        "container_with_children.image.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_image_0",
						ShortName:       "container_dashboard_container_with_child_res_anonymous_container_0_anonymous_image_0",
						UnqualifiedName: "image.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_image_0",
						Title:           toStringPointer("example image"),
						Src:             toStringPointer("https://steampipe.io/images/logo.png"),
						Alt:             toStringPointer("steampipe"),
					},
				},
				DashboardTables: map[string]*modconfig.DashboardTable{
					"container_with_children.table.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_table_0": {
						FullName:        "container_with_children.table.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_table_0",
						ShortName:       "container_dashboard_container_with_child_res_anonymous_container_0_anonymous_table_0",
						UnqualifiedName: "table.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_table_0",
						Title:           toStringPointer("example table"),
						SQL:             toStringPointer("select 1"),
					},
				},
				DashboardTexts: map[string]*modconfig.DashboardText{
					"container_with_children.text.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_text_0": {
						FullName:        "container_with_children.text.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_text_0",
						ShortName:       "container_dashboard_container_with_child_res_anonymous_container_0_anonymous_text_0",
						UnqualifiedName: "text.container_dashboard_container_with_child_res_anonymous_container_0_anonymous_text_0",
						Value:           toStringPointer("example text"),
					},
				},
			},
		},
		"dashboard_with_all_children": { // this test parsing of a dashboard with all possible children
			source: "testdata/mods/dashboard_with_all_children",
			expected: &modconfig.Mod{
				ShortName:   "dashboard_with_children",
				FullName:    "mod.dashboard_with_children",
				Require:     require,
				Description: toStringPointer("This mod contains a dashboard with all possible child resources"),
				Title:       toStringPointer("dashboard with all possible child resources"),
				Dashboards: map[string]*modconfig.Dashboard{
					"dashboard_with_children.dashboard.dashboard_with_child_res": {
						ShortName:       "dashboard_with_child_res",
						FullName:        "dashboard_with_children.dashboard.dashboard_with_child_res",
						UnqualifiedName: "dashboard.dashboard_with_child_res",
						Title:           toStringPointer("dashboard with child resources"),
						ChildNames:      []string{"dashboard_with_children.container.dashboard_dashboard_with_child_res_anonymous_container_0", "dashboard_with_children.chart.dashboard_dashboard_with_child_res_anonymous_chart_0", "dashboard_with_children.card.dashboard_dashboard_with_child_res_anonymous_card_0", "dashboard_with_children.hierarchy.dashboard_dashboard_with_child_res_anonymous_hierarchy_0", "dashboard_with_children.image.dashboard_dashboard_with_child_res_anonymous_image_0", "dashboard_with_children.input.i1", "dashboard_with_children.table.dashboard_dashboard_with_child_res_anonymous_table_0", "dashboard_with_children.text.dashboard_dashboard_with_child_res_anonymous_text_0"},
						HclType:         "dashboard",
					},
				},
				DashboardContainers: map[string]*modconfig.DashboardContainer{
					"dashboard_with_children.container.dashboard_dashboard_with_child_res_anonymous_container_0": {
						ShortName:       "dashboard_dashboard_with_child_res_anonymous_container_0",
						FullName:        "dashboard_with_children.container.dashboard_dashboard_with_child_res_anonymous_container_0",
						UnqualifiedName: "container.dashboard_dashboard_with_child_res_anonymous_container_0",
						Title:           toStringPointer("example container"),
					},
				},
				DashboardCards: map[string]*modconfig.DashboardCard{
					"dashboard_with_children.card.dashboard_dashboard_with_child_res_anonymous_card_0": {
						FullName:        "dashboard_with_children.card.dashboard_dashboard_with_child_res_anonymous_card_0",
						ShortName:       "dashboard_dashboard_with_child_res_anonymous_card_0",
						UnqualifiedName: "card.dashboard_dashboard_with_child_res_anonymous_card_0",
						Title:           toStringPointer("example card"),
						Type:            toStringPointer("ok"),
						SQL:             toStringPointer("select 1"),
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"dashboard_with_children.chart.dashboard_dashboard_with_child_res_anonymous_chart_0": {
						FullName:        "dashboard_with_children.chart.dashboard_dashboard_with_child_res_anonymous_chart_0",
						ShortName:       "dashboard_dashboard_with_child_res_anonymous_chart_0",
						UnqualifiedName: "chart.dashboard_dashboard_with_child_res_anonymous_chart_0",
						Title:           toStringPointer("example chart"),
						SQL:             toStringPointer("select 1"),
					},
				},
				DashboardHierarchies: map[string]*modconfig.DashboardHierarchy{
					"dashboard_with_children.hierarchy.dashboard_dashboard_with_child_res_anonymous_hierarchy_0": {
						FullName:        "dashboard_with_children.hierarchy.dashboard_dashboard_with_child_res_anonymous_hierarchy_0",
						ShortName:       "dashboard_dashboard_with_child_res_anonymous_hierarchy_0",
						UnqualifiedName: "hierarchy.dashboard_dashboard_with_child_res_anonymous_hierarchy_0",
						Title:           toStringPointer("example hierarchy"),
						Type:            toStringPointer("graph"),
					},
				},
				DashboardImages: map[string]*modconfig.DashboardImage{
					"dashboard_with_children.image.dashboard_dashboard_with_child_res_anonymous_image_0": {
						FullName:        "dashboard_with_children.image.dashboard_dashboard_with_child_res_anonymous_image_0",
						ShortName:       "dashboard_dashboard_with_child_res_anonymous_image_0",
						UnqualifiedName: "image.dashboard_dashboard_with_child_res_anonymous_image_0",
						Title:           toStringPointer("example image"),
						Src:             toStringPointer("https://steampipe.io/images/logo.png"),
						Alt:             toStringPointer("steampipe"),
					},
				},
				DashboardInputs: map[string]map[string]*modconfig.DashboardInput{
					"dashboard_with_children.dashboard.dashboard_with_child_res": {
						"dashboard_with_children.input.i1": {
							FullName:        "dashboard_with_children.input.i1",
							ShortName:       "i1",
							UnqualifiedName: "input.i1",
							Title:           toStringPointer("example input"),
						},
					},
				},
				DashboardTables: map[string]*modconfig.DashboardTable{
					"dashboard_with_children.table.dashboard_dashboard_with_child_res_anonymous_table_0": {
						FullName:        "dashboard_with_children.table.dashboard_dashboard_with_child_res_anonymous_table_0",
						ShortName:       "dashboard_dashboard_with_child_res_anonymous_table_0",
						UnqualifiedName: "table.dashboard_dashboard_with_child_res_anonymous_table_0",
						Title:           toStringPointer("example table"),
						SQL:             toStringPointer("select 1"),
					},
				},
				DashboardTexts: map[string]*modconfig.DashboardText{
					"dashboard_with_children.text.dashboard_dashboard_with_child_res_anonymous_text_0": {
						FullName:        "dashboard_with_children.text.dashboard_dashboard_with_child_res_anonymous_text_0",
						ShortName:       "dashboard_dashboard_with_child_res_anonymous_text_0",
						UnqualifiedName: "text.dashboard_dashboard_with_child_res_anonymous_text_0",
						Value:           toStringPointer("example text"),
					},
				},
			},
		},
		"dashboard_resource_naming": { // to test parsing of a resource names properly
			source: "testdata/mods/dashboard_resource_naming",
			expected: &modconfig.Mod{
				ShortName:   "dashboard_resource_naming",
				FullName:    "mod.dashboard_resource_naming",
				Require:     require,
				Description: toStringPointer("this mod is to test the resource naming"),
				Title:       toStringPointer("dashboard resource naming"),
				Dashboards: map[string]*modconfig.Dashboard{
					"dashboard_resource_naming.dashboard.anonymous_naming": {
						FullName:        "dashboard_resource_naming.dashboard.anonymous_naming",
						ShortName:       "anonymous_naming",
						UnqualifiedName: "dashboard.anonymous_naming",
						ChildNames:      []string{"dashboard_resource_naming.chart.dashboard_anonymous_naming_anonymous_chart_0", "dashboard_resource_naming.container.dashboard_anonymous_naming_anonymous_container_0", "dashboard_resource_naming.container.dashboard_anonymous_naming_anonymous_container_1"},
						HclType:         "dashboard",
					},
				},
				DashboardContainers: map[string]*modconfig.DashboardContainer{
					"dashboard_resource_naming.container.dashboard_anonymous_naming_anonymous_container_0": {
						FullName:        "dashboard_resource_naming.container.dashboard_anonymous_naming_anonymous_container_0",
						ShortName:       "dashboard_anonymous_naming_anonymous_container_0",
						UnqualifiedName: "container.dashboard_anonymous_naming_anonymous_container_0",
						ChildNames:      []string{"dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_chart_0", "dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_chart_1", "dashboard_resource_naming.table.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_table_0"},
					},
					"dashboard_resource_naming.container.dashboard_anonymous_naming_anonymous_container_1": {
						FullName:        "dashboard_resource_naming.container.dashboard_anonymous_naming_anonymous_container_1",
						ShortName:       "dashboard_anonymous_naming_anonymous_container_1",
						UnqualifiedName: "container.dashboard_anonymous_naming_anonymous_container_1",
						ChildNames:      []string{"dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_chart_0", "dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_chart_1", "dashboard_resource_naming.table.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_table_0", "dashboard_resource_naming.table.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_table_1"},
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"dashboard_resource_naming.chart.top_level1": {
						FullName:        "dashboard_resource_naming.chart.top_level1",
						ShortName:       "top_level1",
						UnqualifiedName: "chart.top_level1",
						Title:           toStringPointer("top level 1"),
						SQL:             toStringPointer("select 1 as chart"),
					},
					"dashboard_resource_naming.chart.top_level2": {
						FullName:        "dashboard_resource_naming.chart.top_level2",
						ShortName:       "top_level2",
						UnqualifiedName: "chart.top_level2",
						Title:           toStringPointer("top level 2"),
						SQL:             toStringPointer("select 2 as chart"),
					},
					"dashboard_resource_naming.chart.dashboard_anonymous_naming_anonymous_chart_0": {
						FullName:        "dashboard_resource_naming.chart.dashboard_anonymous_naming_anonymous_chart_0",
						ShortName:       "dashboard_anonymous_naming_anonymous_chart_0",
						UnqualifiedName: "chart.dashboard_anonymous_naming_anonymous_chart_0",
						Title:           toStringPointer("chart within dashboard"),
						SQL:             toStringPointer("select 3 as chart"),
					},
					"dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_chart_0": {
						FullName:        "dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_chart_0",
						ShortName:       "container_dashboard_anonymous_naming_anonymous_container_0_anonymous_chart_0",
						UnqualifiedName: "chart.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_chart_0",
						Title:           toStringPointer("chart 1.1"),
						SQL:             toStringPointer("select 4 as chart"),
					},
					"dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_chart_1": {
						FullName:        "dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_chart_1",
						ShortName:       "container_dashboard_anonymous_naming_anonymous_container_0_anonymous_chart_1",
						UnqualifiedName: "chart.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_chart_1",
						Title:           toStringPointer("chart 1.2"),
						SQL:             toStringPointer("select 5 as chart"),
					},
					"dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_chart_0": {
						FullName:        "dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_chart_0",
						ShortName:       "container_dashboard_anonymous_naming_anonymous_container_1_anonymous_chart_0",
						UnqualifiedName: "chart.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_chart_0",
						Title:           toStringPointer("chart 2.1"),
						SQL:             toStringPointer("select 6 as chart"),
					},
					"dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_chart_1": {
						FullName:        "dashboard_resource_naming.chart.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_chart_1",
						ShortName:       "container_dashboard_anonymous_naming_anonymous_container_1_anonymous_chart_1",
						UnqualifiedName: "chart.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_chart_1",
						Title:           toStringPointer("chart 2.2"),
						SQL:             toStringPointer("select 7 as chart"),
					},
				},
				DashboardTables: map[string]*modconfig.DashboardTable{
					"dashboard_resource_naming.table.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_table_0": {
						FullName:        "dashboard_resource_naming.table.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_table_0",
						ShortName:       "container_dashboard_anonymous_naming_anonymous_container_0_anonymous_table_0",
						UnqualifiedName: "table.container_dashboard_anonymous_naming_anonymous_container_0_anonymous_table_0",
						Title:           toStringPointer("table 1.1"),
						SQL:             toStringPointer("select 1 as table"),
					},
					"dashboard_resource_naming.table.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_table_0": {
						FullName:        "dashboard_resource_naming.table.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_table_0",
						ShortName:       "container_dashboard_anonymous_naming_anonymous_container_1_anonymous_table_0",
						UnqualifiedName: "table.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_table_0",
						Title:           toStringPointer("table 2.1"),
						SQL:             toStringPointer("select 2 as table"),
					},
					"dashboard_resource_naming.table.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_table_1": {
						FullName:        "dashboard_resource_naming.table.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_table_1",
						ShortName:       "container_dashboard_anonymous_naming_anonymous_container_1_anonymous_table_1",
						UnqualifiedName: "table.container_dashboard_anonymous_naming_anonymous_container_1_anonymous_table_1",
						Title:           toStringPointer("table 2.2"),
						SQL:             toStringPointer("select 3 as table"),
					},
				},
			},
		},
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
	err = setChildren(expectedMod)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expectedMod.BuildResourceTree(nil)

	if !actualMod.Equals(expectedMod) {
		fmt.Printf("")

		t.Errorf("Test: '%s'' FAILED", name)
	}
}

// try to resolve mod resource children using their child names
func setChildren(mod *modconfig.Mod) error {
	for _, benchmark := range mod.Benchmarks {
		for _, childName := range benchmark.ChildNames {
			parsed, _ := modconfig.ParseResourceName(childName.Name)
			child, found := modconfig.GetResource(mod, parsed)
			if !found {
				return fmt.Errorf("failed to resolve child %s", childName)
			}
			benchmark.Children = append(benchmark.Children, child.(modconfig.ModTreeItem))
		}
	}
	for _, container := range mod.DashboardContainers {
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
	for _, dashboard := range mod.Dashboards {
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
