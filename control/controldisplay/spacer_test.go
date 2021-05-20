package controldisplay

import (
	"fmt"
	"testing"
)

type spacerTest struct {
	width    int
	expected string
}

func testCasesSpacer() map[string]spacerTest {
	return map[string]spacerTest{

		"2": {
			2, fmt.Sprintf("%s", ControlColors.Spacer("..")),
		},
		"3": {
			3, fmt.Sprintf("%s", ControlColors.Spacer("...")),
		},
		"10": {
			10, fmt.Sprintf("%s", ControlColors.Spacer("..........")),
		},
	}
}

func TestSpacer(t *testing.T) {
	themeDef := ColorSchemes["dark"]
	scheme, _ := NewControlColorScheme(themeDef)
	ControlColors = scheme

	for name, test := range testCasesSpacer() {
		output := NewSpacerRenderer(test.width).Render()
		if output != test.expected {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v \ngot:\n %v\n", name, test.expected, output)
		}
	}
}
