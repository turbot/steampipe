package tabledisplay

import (
	"fmt"
	"testing"
)

type spacerTest struct {
	width    int
	expected string
}

var testCasesSpacer = map[string]spacerTest{

	"2": {
		2, fmt.Sprintf("%s", colorSpacer("..")),
	},
	"3": {
		3, fmt.Sprintf("%s", colorSpacer("...")),
	},
	"10": {
		10, fmt.Sprintf("%s", colorSpacer("..........")),
	},
}

func TestSpacer(t *testing.T) {
	for name, test := range testCasesSpacer {
		output := NewSpacerRenderer(test.width).Render()
		if output != test.expected {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v \ngot:\n %v\n", name, test.expected, output)
		}
	}
}
