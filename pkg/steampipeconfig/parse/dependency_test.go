package parse

import (
	"reflect"
	"testing"

	"github.com/turbot/go-kit/helpers"
)

type dependencyTreeTest struct {
	input    [][]string
	expected []string
}

var testCasesDependencyTree = map[string]dependencyTreeTest{
	"no overlap": {
		input:    [][]string{{"a", "b", "c"}, {"d", "e", "f"}},
		expected: []string{"a", "b", "c", "d", "e", "f"},
	},
	"overlap": {
		input:    [][]string{{"a", "b", "c"}, {"b", "c"}},
		expected: []string{"a", "b", "c"},
	},
	"multiple overlaps": {
		input:    [][]string{{"a", "b", "c"}, {"b", "c"}, {"d", "e", "f", "g", "h", "i"}, {"g", "h", "i"}, {"h", "i"}},
		expected: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"},
	},
}

func TestDependencyTree(t *testing.T) {

	for name, test := range testCasesDependencyTree {

		res := combineDependencyOrders(test.input)
		if !reflect.DeepEqual(res, test.expected) {
			t.Errorf("Test %s FAILED. Expected %v, got %v", name, test.expected, res)
		}
	}
}

func combineDependencyOrders(deps [][]string) []string {

	// we assume every dep is unique
	// for each dep, if first element exists in any other dep, then it cannot be the longest
	// first dedupe
	var longestDeps []string
	for i, d1 := range deps {
		longest := true
		for j, d2 := range deps {
			if i == j {
				continue
			}
			if helpers.StringSliceContains(d2, d1[0]) {
				longest = false
				continue
			}
		}
		if longest {
			longestDeps = append(longestDeps, d1...)
		}
	}

	return longestDeps
}
