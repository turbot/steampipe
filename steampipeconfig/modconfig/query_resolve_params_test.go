package modconfig

import (
	"testing"

	"github.com/turbot/steampipe/utils"
)

type resolveParamsTest struct {
	params    *QueryArgs
	paramDefs []*ParamDef
	expected  interface{}
}

var testCasesResolveParams = map[string]resolveParamsTest{
	"positional params no defs": {
		params: &QueryArgs{
			ArgsList: []string{"val1", "val2"},
		},
		paramDefs: nil,
		expected:  "(array['val1','val2'])",
	},
	"named params no defs": {
		params: &QueryArgs{
			Args: map[string]string{
				"p1": "val1",
				"p2": "val2",
			},
		},
		paramDefs: nil,
		expected:  "ERROR",
	},
	"named params with defs": {
		params: &QueryArgs{
			Args: map[string]string{
				"p1": "val1",
				"p2": "val2",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1"},
			{Name: "p2"},
		},
		expected: "(array['val1','val2'])",
	},
	"partial named params with defs and defaults": {
		params: &QueryArgs{
			Args: map[string]string{
				"p1": "val1",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("def_val1")},
			{Name: "p2", Default: utils.ToStringPointer("def_val2")},
		},
		expected: "(array['val1','def_val2'])",
	},
	"partial positional params with defs and defaults": {
		params: &QueryArgs{
			ArgsList: []string{"val1"},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("def_val1")},
			{Name: "p2", Default: utils.ToStringPointer("def_val2")},
		},
		expected: "(array['val1','def_val2'])",
	},
	"partial positional params with defs and unmatched defaults": {
		// only a default for first param, which is populated from the provided positional param
		params: &QueryArgs{
			ArgsList: []string{"val1"},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("def_val1")},
			{Name: "p2"},
		},
		expected: "ERROR",
	},
	"partial named params with defs and unmatched defaults": {
		// only a default for first param, which is populated from the provided positional param
		params: &QueryArgs{
			Args: map[string]string{
				"p1": "val1",
			},
		},
		paramDefs: []*ParamDef{
			{Name: "p1", Default: utils.ToStringPointer("def_val1")},
			{Name: "p2"},
		},
		expected: "ERROR",
	},
	"positional and named params": {
		params: &QueryArgs{
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

func TestResolveParams(t *testing.T) {
	for name, test := range testCasesResolveParams {
		query := &Query{ParamsDefs: test.paramDefs}
		res, err := query.ResolveParams(test.params)
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
