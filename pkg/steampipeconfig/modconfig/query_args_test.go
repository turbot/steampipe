package modconfig

import (
	"testing"

	"github.com/turbot/steampipe/pkg/utils"
)

type resolveParamsTest struct {
	baseArgs    *QueryArgs
	runtimeArgs *QueryArgs
	paramDefs   []*ParamDef
	expected    interface{}
}

var testCasesResolveParams = map[string]resolveParamsTest{

	"named params no defs (expect error)": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "'val1'",
				"p2": "'val2'",
			},
		},
		paramDefs: nil,
		expected:  "ERROR",
	},
	"named params with defs": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "'val1'",
				"p2": "'val2'",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: "('val1','val2')",
	},
	"named params with defs and partial runtime overrides": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "'val1'",
				"p2": "'val2'",
			},
		},
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p2": "'runtime val2'",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: "('val1','runtime val2')",
	},

	"named params with defs and full runtime overrides": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "'val1'",
				"p2": "'val2'",
			},
		},
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "'runtime val1'",
				"p2": "'runtime val2'",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: "('runtime val1','runtime val2')",
	},
	"named params with defs and invalid runtime override": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "'val1'",
				"p2": "'val2'",
			},
		},
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p2": "'runtime val2'",
				"p3": "'runtime val2'",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: "ERROR",
	},

	"named params overrides only with defs": {
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "'override val1'",
				"p2": "'override val2'",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: "('override val1','override val2')",
	},
	"named param defs with incomplete overrides": {
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p2": "'override val2'",
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
				"p3": "'override val2'",
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
				"p2": "'override val2'",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("'val1'")},
			{Name: "p2", Default: utils.ToStringPointer("'val2'")},
		},
		expected: "('val1','override val2')",
	},
	"named param defs with defaults with incomplete invalid overrides": {
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p3": "'override val2'",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("'val1'")},
			{Name: "p2", Default: utils.ToStringPointer("'val2'")},
		},
		expected: "ERROR",
	},

	"partial named params with defs and defaults": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "'val1'",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("'def_val1'")},
			{Name: "p2", Default: utils.ToStringPointer("'def_val2'")},
		},
		expected: "('val1','def_val2')",
	},
	"partial named params with defs defaults and partial override": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "'val1'",
			},
		},
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p2": "'override val2'",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("'def_val1'")},
			{Name: "p2", Default: utils.ToStringPointer("'def_val2'")},
			{Name: "p3", Default: utils.ToStringPointer("'def_val3'")},
		},
		expected: "('val1','override val2','def_val3')",
	},
	"partial named params with defs and unmatched defaults": {
		// only a default for first param, which is populated from the provided positional param
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "val1",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("def_val1")},
			{Name: "p2"},
		},
		expected: "ERROR",
	},

	"positional params no defs": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer("'val1'"), utils.ToStringPointer("'val2'")},
		},
		paramDefs: nil,
		expected:  "('val1','val2')",
	},
	"positional params with partial runtime override no defs": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer("'val1'"), utils.ToStringPointer("'val2'")},
		},
		runtimeArgs: &QueryArgs{
			ArgList: []*string{nil, utils.ToStringPointer("'override val2'")},
		},
		paramDefs: nil,
		expected:  "('val1','override val2')",
	},
	"positional params with full runtime override no defs": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer("'val1'"), utils.ToStringPointer("'val2'")},
		},
		runtimeArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer("'override val1'"), utils.ToStringPointer("'override val2'")},
		},
		paramDefs: nil,
		expected:  "('override val1','override val2')",
	},
	"partial positional params with defs and defaults": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer("'val1'")},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("'def_val1'")},
			{Name: "p2", Default: utils.ToStringPointer("'def_val2'")},
		},
		expected: "('val1','def_val2')",
	},
	"partial positional params with defs, overrides and defaults": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer("'val1'")},
		},
		runtimeArgs: &QueryArgs{
			ArgList: []*string{nil, utils.ToStringPointer("'override val2'")},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("'def_val1'")},
			{Name: "p2", Default: utils.ToStringPointer("'def_val2'")},
			{Name: "p3", Default: utils.ToStringPointer("'def_val3'")},
		},
		expected: "('val1','override val2','def_val3')",
	},
	"partial positional params with defs and unmatched defaults": {
		// only a default for first param, which is populated from the provided positional param
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer("val1")},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("def_val1")},
			{Name: "p2"},
		},
		expected: "ERROR",
	},

	"positional and named params (expect error)": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer("val1"), utils.ToStringPointer("val2")},
			ArgMap: map[string]string{
				"p1": "val1",
				"p2": "val2",
			},
		},
		paramDefs: nil,
		expected:  "ERROR",
	},
	"positional and override named params (expect error)": {
		baseArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer("val1"), utils.ToStringPointer("val2")},
		},
		runtimeArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "val1",
				"p2": "val2",
			},
		},
		paramDefs: nil,
		expected:  "ERROR",
	},
	"named and override params (expect error)": {
		baseArgs: &QueryArgs{
			ArgMap: map[string]string{
				"p1": "val1",
				"p2": "val2",
			},
		},
		runtimeArgs: &QueryArgs{
			ArgList: []*string{utils.ToStringPointer("val1"), utils.ToStringPointer("val2")},
		},
		paramDefs: nil,
		expected:  "ERROR",
	},
}

func TestResolveAsString(t *testing.T) {
	for name, test := range testCasesResolveParams {
		query := &Control{FullName: "control.test_control", Params: test.paramDefs, Args: test.baseArgs}
		res, _, err := ResolveArgsAsString(query, test.runtimeArgs)
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
		if test.expected != res {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v, \ngot:\n %v\n", name, test.expected, res)
		}
	}
}
