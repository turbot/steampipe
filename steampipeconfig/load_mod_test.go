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
		// "single_mod_duplicate_query": {
		// 	source:   "testdata/mods/single_mod_duplicate_query",
		// 	expected: "ERROR",
		// },
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
						ChildNames:      []string{"simple_report.text.anonymous_text", "simple_report.chart.anonymous_chart"},
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"simple_report.chart.anonymous_chart": {
						FullName:        "simple_report.chart.anonymous_chart",
						ShortName:       "anonymous_chart",
						UnqualifiedName: "chart.anonymous_chart",
						Title:           toStringPointer("a simple query"),
						SQL:             toStringPointer("select 1"),
					},
				},
				DashboardTexts: map[string]*modconfig.DashboardText{
					"simple_report.text.anonymous_text": {
						FullName:        "simple_report.text.anonymous_text",
						ShortName:       "anonymous_text",
						UnqualifiedName: "text.anonymous_text",
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
						ChildNames:      []string{"simple_container_report.container.anonymous_container"},
						HclType:         "report",
					},
				},
				DashboardContainers: map[string]*modconfig.DashboardContainer{
					"simple_container_report.container.anonymous_container": {
						ShortName:       "anonymous_container",
						FullName:        "simple_container_report.container.anonymous_container",
						UnqualifiedName: "container.anonymous_container",
						ChildNames:      []string{"simple_container_report.text.anonymous_text", "simple_container_report.chart.anonymous_chart"},
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"simple_container_report.chart.anonymous_chart": {
						ShortName:       "anonymous_chart",
						FullName:        "simple_container_report.chart.anonymous_chart",
						UnqualifiedName: "chart.anonymous_chart",
						Title:           toStringPointer("container 1 chart 1"),
						SQL:             toStringPointer("select 1 as container"),
					},
				},
				DashboardTexts: map[string]*modconfig.DashboardText{
					"simple_container_report.text.anonymous_text": {
						ShortName:       "anonymous_text",
						FullName:        "simple_container_report.text.anonymous_text",
						UnqualifiedName: "text.anonymous_text",
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
						ChildNames:      []string{"sibling_containers_report.container.anonymous_container", "sibling_containers_report.container.anonymous_container_1", "sibling_containers_report.container.anonymous_container_2"},
						HclType:         "report",
					},
				},
				DashboardContainers: map[string]*modconfig.DashboardContainer{
					"sibling_containers_report.container.anonymous_container": {
						ShortName:       "anonymous_container",
						FullName:        "sibling_containers_report.container.anonymous_container",
						UnqualifiedName: "container.anonymous_container",
						ChildNames:      []string{"sibling_containers_report.text.anonymous_text", "sibling_containers_report.chart.anonymous_chart"},
					},
					"sibling_containers_report.container.anonymous_container_1": {
						ShortName:       "anonymous_container_1",
						FullName:        "sibling_containers_report.container.anonymous_container_1",
						UnqualifiedName: "container.anonymous_container_1",
						ChildNames:      []string{"sibling_containers_report.text.anonymous_text_1", "sibling_containers_report.chart.anonymous_chart_1"},
					},
					"sibling_containers_report.container.anonymous_container_2": {
						ShortName:       "anonymous_container_2",
						FullName:        "sibling_containers_report.container.anonymous_container_2",
						UnqualifiedName: "container.anonymous_container_2",
						ChildNames:      []string{"sibling_containers_report.text.anonymous_text_2", "sibling_containers_report.chart.anonymous_chart_2"},
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"sibling_containers_report.chart.anonymous_chart": {
						FullName:        "sibling_containers_report.chart.anonymous_chart",
						ShortName:       "anonymous_chart",
						UnqualifiedName: "chart.anonymous_chart",
						Title:           toStringPointer("container 1 chart 1"),
						SQL:             toStringPointer("select 1 as container"),
					},
					"sibling_containers_report.chart.anonymous_chart_1": {
						FullName:        "sibling_containers_report.chart.anonymous_chart_1",
						ShortName:       "anonymous_chart_1",
						UnqualifiedName: "chart.anonymous_chart_1",
						Title:           toStringPointer("container 2 chart 1"),
						SQL:             toStringPointer("select 2 as container"),
					},
					"sibling_containers_report.chart.anonymous_chart_2": {
						FullName:        "sibling_containers_report.chart.anonymous_chart_2",
						ShortName:       "anonymous_chart_2",
						UnqualifiedName: "chart.anonymous_chart_2",
						Title:           toStringPointer("container 3 chart 1"),
						SQL:             toStringPointer("select 3 as container"),
					},
				},
				DashboardTexts: map[string]*modconfig.DashboardText{
					"sibling_containers_report.text.anonymous_text": {
						FullName:        "sibling_containers_report.text.anonymous_text",
						ShortName:       "anonymous_text",
						UnqualifiedName: "text.anonymous_text",
						Value:           toStringPointer("container 1"),
					},
					"sibling_containers_report.text.anonymous_text_1": {
						FullName:        "sibling_containers_report.text.anonymous_text_1",
						ShortName:       "anonymous_text_1",
						UnqualifiedName: "text.anonymous_text_1",
						Value:           toStringPointer("container 2"),
					},
					"sibling_containers_report.text.anonymous_text_2": {
						FullName:        "sibling_containers_report.text.anonymous_text_2",
						ShortName:       "anonymous_text_2",
						UnqualifiedName: "text.anonymous_text_2",
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
						ChildNames:      []string{"nested_containers_report.container.anonymous_container"},
						HclType:         "dashboard",
					},
				},
				DashboardContainers: map[string]*modconfig.DashboardContainer{
					"nested_containers_report.container.anonymous_container_1": {
						ShortName:       "anonymous_container_1",
						FullName:        "nested_containers_report.container.anonymous_container_1",
						UnqualifiedName: "container.anonymous_container_1",
						ChildNames:      []string{"nested_containers_report.text.anonymous_text_1", "nested_containers_report.chart.anonymous_chart"},
					},
					"nested_containers_report.container.anonymous_container_3": {
						ShortName:       "anonymous_container_3",
						FullName:        "nested_containers_report.container.anonymous_container_3",
						UnqualifiedName: "container.anonymous_container_3",
						ChildNames:      []string{"nested_containers_report.text.anonymous_text_3", "nested_containers_report.chart.anonymous_chart_2"},
					},
					"nested_containers_report.container.anonymous_container_2": {
						ShortName:       "anonymous_container_2",
						FullName:        "nested_containers_report.container.anonymous_container_2",
						UnqualifiedName: "container.anonymous_container_2",
						ChildNames:      []string{"nested_containers_report.text.anonymous_text_2", "nested_containers_report.chart.anonymous_chart_1", "nested_containers_report.container.anonymous_container_3"},
					},
					"nested_containers_report.container.anonymous_container": {
						ShortName:       "anonymous_container",
						FullName:        "nested_containers_report.container.anonymous_container",
						UnqualifiedName: "container.anonymous_container",
						ChildNames:      []string{"nested_containers_report.text.anonymous_text", "nested_containers_report.container.anonymous_container_1", "nested_containers_report.container.anonymous_container_2"},
					},
				},
				DashboardCharts: map[string]*modconfig.DashboardChart{
					"nested_containers_report.chart.anonymous_chart": {
						FullName:        "nested_containers_report.chart.anonymous_chart",
						ShortName:       "anonymous_chart",
						UnqualifiedName: "chart.anonymous_chart",
						Title:           toStringPointer("CHART 1"),
						SQL:             toStringPointer("select 1.1 as container"),
					},
					"nested_containers_report.chart.anonymous_chart_1": {
						FullName:        "nested_containers_report.chart.anonymous_chart_1",
						ShortName:       "anonymous_chart_1",
						UnqualifiedName: "chart.anonymous_chart_1",
						Title:           toStringPointer("CHART 2"),
						SQL:             toStringPointer("select 1.2 as container"),
					},
					"nested_containers_report.chart.anonymous_chart_2": {
						FullName:        "nested_containers_report.chart.anonymous_chart_2",
						ShortName:       "anonymous_chart_2",
						UnqualifiedName: "chart.anonymous_chart_2",
						Title:           toStringPointer("CHART 3"),
						SQL:             toStringPointer("select 1.2.1 as container"),
					},
				},
				DashboardTexts: map[string]*modconfig.DashboardText{
					"nested_containers_report.text.anonymous_text": {
						FullName:        "nested_containers_report.text.anonymous_text",
						ShortName:       "anonymous_text",
						UnqualifiedName: "text.anonymous_text",
						Value:           toStringPointer("CONTAINER 1"),
					},
					"nested_containers_report.text.anonymous_text_1": {
						FullName:        "nested_containers_report.text.anonymous_text_1",
						ShortName:       "anonymous_text_1",
						UnqualifiedName: "text.anonymous_text_1",
						Value:           toStringPointer("CHILD CONTAINER 1.1"),
					},
					"nested_containers_report.text.anonymous_text_2": {
						FullName:        "nested_containers_report.text.anonymous_text_2",
						ShortName:       "anonymous_text_2",
						UnqualifiedName: "text.anonymous_text_2",
						Value:           toStringPointer("CHILD CONTAINER 1.2"),
					},
					"nested_containers_report.text.anonymous_text_3": {
						FullName:        "nested_containers_report.text.anonymous_text_3",
						ShortName:       "anonymous_text_3",
						UnqualifiedName: "text.anonymous_text_3",
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
						ChildNames:      []string{"report_axes.chart.anonymous_chart"},
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
					"report_axes.chart.anonymous_chart": {
						FullName:        "report_axes.chart.anonymous_chart",
						ShortName:       "anonymous_chart",
						UnqualifiedName: "chart.anonymous_chart",
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
						ChildNames:      []string{"report_base1.chart.anonymous_chart"},
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
					"report_base1.chart.anonymous_chart": {
						FullName:        "report_base1.chart.anonymous_chart",
						ShortName:       "anonymous_chart",
						UnqualifiedName: "chart.anonymous_chart",
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
			child, found := modconfig.GetResource(mod, parsed)
			if !found {
				return fmt.Errorf("failed to resolve child %s", childName)
			}
			children = append(children, child.(modconfig.ModTreeItem))
		}
		container.SetChildren(children)

	}
	for _, report := range mod.Dashboards {
		var children []modconfig.ModTreeItem
		for _, childName := range report.ChildNames {
			parsed, _ := modconfig.ParseResourceName(childName)
			child, found := modconfig.GetResource(mod, parsed)
			if !found {
				return fmt.Errorf("failed to resolve child %s", childName)
			}
			children = append(children, child.(modconfig.ModTreeItem))
		}
		report.SetChildren(children)
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
