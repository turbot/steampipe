package modconfig

import (
	"github.com/turbot/go-kit/helpers"
	"reflect"
	"testing"
)

type parsePropertyPathTest struct {
	input    string
	expected any
}

var parsePropertyPathTestCases = map[string]parsePropertyPathTest{
	"unqualified resource name": {
		input: "query.q1",
		expected: &ParsedPropertyPath{
			ItemType: "query",
			Name:     "q1",
			Original: "query.q1",
		},
	},
	"qualified resource name": {
		input: "m1.query.q1",
		expected: &ParsedPropertyPath{
			Mod:      "m1",
			ItemType: "query",
			Name:     "q1",
			Original: "m1.query.q1",
		},
	},
	"unqualified resource property path": {
		input: "query.q1.foo.bar",
		expected: &ParsedPropertyPath{
			ItemType:     "query",
			Name:         "q1",
			PropertyPath: []string{"foo", "bar"},
			Original:     "query.q1.foo.bar",
		},
	},
	"qualified resource property path": {
		input: "m1.query.q1.foo.bar",
		expected: &ParsedPropertyPath{
			Mod:          "m1",
			ItemType:     "query",
			Name:         "q1",
			PropertyPath: []string{"foo", "bar"},
			Original:     "m1.query.q1.foo.bar",
		},
	},
	"self input": {
		input: "self.input.foo",
		expected: &ParsedPropertyPath{
			ItemType: "input",
			Name:     "foo",
			Scope:    "self",
			Original: "self.input.foo",
		},
	},
	"with": {
		input: "with.w1",
		expected: &ParsedPropertyPath{
			ItemType: "with",
			Name:     "w1",
			Original: "with.w1",
		},
	},
	"with property path": {
		input: "with.w1.c1",
		expected: &ParsedPropertyPath{
			ItemType:     "with",
			Name:         "w1",
			PropertyPath: []string{"c1"},
			Original:     "with.w1.c1",
		},
	},
}

func TestParsePropertyPath(t *testing.T) {
	testsToRun := []string{"self input"}

	for name, test := range parsePropertyPathTestCases {
		if len(testsToRun) > 0 && !helpers.StringSliceContains(testsToRun, name) {
			continue
		}

		res, err := ParseResourcePropertyPath(test.input)
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
		if !propertyPathsEqual(res, test.expected.(*ParsedPropertyPath)) {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v, \ngot:\n %v\n", name, test.expected, res)
		}
	}
}

func propertyPathsEqual(l, r *ParsedPropertyPath) bool {
	return l.Mod == r.Mod &&
		l.ItemType == r.ItemType &&
		l.Name == r.Name &&
		reflect.DeepEqual(l.PropertyPath, r.PropertyPath) &&
		l.Scope == r.Scope &&
		l.Original == r.Original

}
