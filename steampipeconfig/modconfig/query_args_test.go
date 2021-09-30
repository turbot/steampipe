package modconfig

import (
	"testing"

	"github.com/turbot/steampipe/utils"
)

type resolveParamsTest struct {
	args      *QueryArgs
	paramDefs []*ParamDef
	expected  interface{}
}

var testCasesResolveParams = map[string]resolveParamsTest{
	"positional params no defs": {
		args: &QueryArgs{
			ArgsList: []string{"'val1'", "'val2'"},
		},
		paramDefs: nil,
		expected:  "('val1','val2')",
	},
	"named params no defs": {
		args: &QueryArgs{
			Args: map[string]string{
				"p1": "'val1'",
				"p2": "'val2'",
			},
		},
		paramDefs: nil,
		expected:  "ERROR",
	},
	"named params with defs": {
		args: &QueryArgs{
			Args: map[string]string{
				"p1": "'val1'",
				"p2": "'val2'",
			},
		},
		paramDefs: []*ParamDef{
			{ShortName: "p1"},
			{ShortName: "p2"},
		},
		expected: "('val1','val2')",
	},
	"partial named params with defs and defaults": {
		args: &QueryArgs{
			Args: map[string]string{
				"p1": "'val1'",
			},
		},
		paramDefs: []*ParamDef{
			{ShortName: "p1", Default: utils.ToStringPointer("'def_val1'")},
			{ShortName: "p2", Default: utils.ToStringPointer("'def_val2'")},
		},
		expected: "('val1','def_val2')",
	},
	"partial positional params with defs and defaults": {
		args: &QueryArgs{
			ArgsList: []string{"'val1'"},
		},
		paramDefs: []*ParamDef{
			{ShortName: "p1", Default: utils.ToStringPointer("'def_val1'")},
			{ShortName: "p2", Default: utils.ToStringPointer("'def_val2'")},
		},
		expected: "('val1','def_val2')",
	},
	"partial positional params with defs and unmatched defaults": {
		// only a default for first param, which is populated from the provided positional param
		args: &QueryArgs{
			ArgsList: []string{"val1"},
		},
		paramDefs: []*ParamDef{
			{ShortName: "p1", Default: utils.ToStringPointer("def_val1")},
			{ShortName: "p2"},
		},
		expected: "ERROR",
	},
	"partial named params with defs and unmatched defaults": {
		// only a default for first param, which is populated from the provided positional param
		args: &QueryArgs{
			Args: map[string]string{
				"p1": "val1",
			},
		},
		paramDefs: []*ParamDef{
			{ShortName: "p1", Default: utils.ToStringPointer("def_val1")},
			{ShortName: "p2"},
		},
		expected: "ERROR",
	},
	"positional and named params": {
		args: &QueryArgs{
			ArgsList: []string{"val1", "val2"},
			Args: map[string]string{
				"p1": "val1",
				"p2": "val2",
			},
		},
		paramDefs: nil,
		expected:  "ERROR",
	},
}

func TestResolveAsString(t *testing.T) {
	for name, test := range testCasesResolveParams {
		query := &Query{Params: test.paramDefs}
		res, err := test.args.ResolveAsString(query)
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
