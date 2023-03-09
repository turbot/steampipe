package type_conversion

import (
	"testing"
)

type goToPostgresTest struct {
	input    any
	expected any
}

var goToPostgresTestCases = map[string]goToPostgresTest{
	"int": {
		input: 1, expected: "1",
	},
	"float": {
		input: 1.1, expected: "1.1",
	},
	"string": {
		input: "foo", expected: "'foo'",
	},
	"string slice": {
		input: []string{"foo", "bar"}, expected: `array['foo','bar']::text[]`,
	},
	"string interface slice": {
		input: []any{"foo", "bar"}, expected: `array['foo','bar']::text[]`,
	},
	"int slice": {
		input: []int{1, 2}, expected: `array[1,2]::numeric[]`,
	},
	"int any slice": {
		input: []any{1, 2}, expected: `array[1,2]::numeric[]`,
	},
	"slice of arrays": {
		input: []any{[]int{1, 2}, []int{3, 4}}, expected: `array['[1,2]'::jsonb,'[3,4]'::jsonb]::jsonb[]`,
	},

	"any slice mixed types": {
		input: []any{1, "foo"}, expected: `ERROR`,
	},
}

func TestGoToPostgres(t *testing.T) {
	for name, test := range goToPostgresTestCases {
		res, err := GoToPostgresString(test.input)
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
