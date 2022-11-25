package modconfig

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/utils"
	"reflect"
	"testing"
)

type resolveParamsTest struct {
	baseArgs    *QueryArgs
	runtimeArgs *QueryArgs
	paramDefs   []*ParamDef
	expected    interface{}
}

// NOTE: all QueryArgs values are Json representations of the arg value
// TODO really we should update the trest to set stringNamedArgs and stringPositionalArgs for each args object
// then we can store the string args as normal strings, not json strings
// TODO add other args types - arrays, json etc.

var testCasesResolveParams = map[string]resolveParamsTest{

	"named argsno defs": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"val1"`,
				"p2": `"val2"`,
			},
		},
		paramDefs: nil,
		expected:  []any(nil),
	},
	"named args with defs": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"val1"`,
				"p2": `"val2"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: []any{"val1", "val2"},
	},
	"named args with defs and partial runtime overrides": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"val1"`,
				"p2": `"val2"`,
			},
		},
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p2": `"runtime val2"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: []any{"val1", "runtime val2"},
	},

	"named args with defs and full runtime overrides": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"val1"`,
				"p2": `"val2"`,
			},
		},
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"runtime val1"`,
				"p2": `"runtime val2"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: []any{"runtime val1", "runtime val2"},
	},
	"named args with defs and runtime overrides with additional undefined arg": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"val1"`,
				"p2": `"val2"`,
			},
		},
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p2": `"runtime val2"`,
				"p3": `"runtime val3"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: []any{"val1", "runtime val2"},
	},

	"named arg overrides only with defs": {
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"override val1"`,
				"p2": `"override val2"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: []any{"override val1", "override val2"},
	},
	"named param defs with incomplete overrides": {
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p2": `"override val2"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: "ERROR",
	},
	"named param defs with incomplete invalid overrides": {
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p3": `"override val2"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: "ERROR",
	},
	"named param defs with defaults with incomplete overrides": {
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p2": `"override val2"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer(`"val1"`)},
			{Name: "p2", Default: utils.ToStringPointer(`"val2"`)},
		},
		expected: []any{"val1", "override val2"},
	},
	"named param defs with defaults with undefined override": {
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p3": `"override val2"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer(`"val1"`)},
			{Name: "p2", Default: utils.ToStringPointer(`"val2"`)},
		},
		expected: []any{"val1", "val2"},
	},

	"partial named args with defs and defaults": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"val1"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer(`"def_val1"`)},
			{Name: "p2", Default: utils.ToStringPointer(`"def_val2"`)},
		},
		expected: []any{"val1", "def_val2"},
	},
	"partial named args with defs defaults and partial override": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"val1"`,
			},
		},
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p2": `"override val2"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer(`"def_val1"`)},
			{Name: "p2", Default: utils.ToStringPointer(`"def_val2"`)},
			{Name: "p3", Default: utils.ToStringPointer(`"def_val3"`)},
		},

		expected: []any{"val1", "override val2", "def_val3"},
	},
	"partial named args with defs and unmatched defaults": {
		// only a default for first param, which is populated from the provided positional param
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"val1"`,
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer(`"def_val1"`)},
			{Name: "p2"},
		},
		expected: "ERROR",
	},

	"positional params no defs": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer(`"val1"`), utils.ToStringPointer(`"val2"`)},
		},
		paramDefs: nil,

		expected: []any{"val1", "val2"},
	},
	"positional params with partial runtime override no defs": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer(`"val1"`), utils.ToStringPointer(`"val2"`)},
		},
		runtimeArgs: &QueryArgs{
			ArgList: []*string{nil, utils.ToStringPointer(`"override val2"`)},
		},
		paramDefs: nil,
		expected:  []any{"val1", "override val2"},
	},
	"positional params with full runtime override no defs": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer(`"val1"`), utils.ToStringPointer(`"val2"`)},
		},
		runtimeArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer(`"override val1"`), utils.ToStringPointer(`"override val2"`)},
		},
		paramDefs: nil,
		expected:  []any{"override val1", "override val2"},
	},
	"partial positional args with defs and defaults": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer(`"val1"`)},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer(`"def_val1"`)},
			{Name: "p2", Default: utils.ToStringPointer(`"def_val2"`)},
		},
		expected: []any{"val1", "def_val2"},
	},
	"partial positional args with defs, overrides and defaults": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer(`"val1"`)},
		},
		runtimeArgs: &QueryArgs{
			ArgList: []*string{nil, utils.ToStringPointer(`"override val2"`)},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer(`"def_val1"`)},
			{Name: "p2", Default: utils.ToStringPointer(`"def_val2"`)},
			{Name: "p3", Default: utils.ToStringPointer(`"def_val3"`)},
		},
		expected: []any{"val1", "override val2", "def_val3"},
	},
	"partial positional args with defs and unmatched defaults": {
		// only a default for first param, which is populated from the provided positional param
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer(`"val1"`)},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer(`"def_val1"`)},
			{Name: "p2"},
		},
		expected: "ERROR",
	},

	"positional and named args(expect error)": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer(`"val1"`), utils.ToStringPointer(`"val2"`)},
			ArgMap: map[string]string{
				"p1": `"val1"`,
				"p2": `"val2"`,
			},
		},
		paramDefs: nil,
		expected:  "ERROR",
	},
	"positional and override named args (expect error)": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer(`"val1"`), utils.ToStringPointer(`"val2"`)},
		},
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"val1"`,
				"p2": `"val2"`,
			},
		},
		paramDefs: nil,
		expected:  "ERROR",
	},
	"named and override params (expect error)": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": `"val1"`,
				"p2": `"val2"`,
			},
		},
		runtimeArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer(`"val1"`), utils.ToStringPointer(`"val2"`)},
		},
		paramDefs: nil,
		expected:  "ERROR",
	},
}

func TestResolveAsString(t *testing.T) {
	testsToRun := []string{}

	for name, test := range testCasesResolveParams {
		if len(testsToRun) > 0 && !helpers.StringSliceContains(testsToRun, name) {
			continue
		}
		query := &Control{FullName: "control.test_control", Params: test.paramDefs, Args: test.baseArgs}
		res, err := ResolveArgs(query, test.runtimeArgs)
		if err != nil {
			if test.expected != "ERROR" {
				t.Errorf("Test: '%s'' FAILED : \nunexpected error %v", name, err)
			}
			continue
		}
		if test.expected == "ERROR" {
			t.Errorf("Test: '%s'' FAILED - expected error", name)
			continue
		}
		expected := test.expected.([]any)
		if !reflect.DeepEqual(expected, res) {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v, \ngot:\n %v\n", name, test.expected, res)
		}
	}
}
