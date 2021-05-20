package controldisplay

import (
	"fmt"
	"testing"
)

type resultReasonTest struct {
	status   string
	reason   string
	width    int
	expected string
}

func testCasesResultReason() map[string]resultReasonTest {
	return map[string]resultReasonTest{
		"error fit": {
			status:   "error",
			reason:   "short error reason",
			width:    100,
			expected: fmt.Sprintf("%s", ControlColors.ReasonColors["error"]("short error reason")),
		},
		"ok fit": {
			status:   "ok",
			reason:   "short ok reason",
			width:    100,
			expected: fmt.Sprintf("%s", ControlColors.ReasonColors["ok"]("short ok reason")),
		},
		"error truncate": {
			status:   "error",
			reason:   "long error reason is very long and goes on and on",
			width:    40,
			expected: fmt.Sprintf("%s", ControlColors.ReasonColors["error"]("long error reason is very long and goes…")),
		},
	}
}

func TestResultReason(t *testing.T) {
	themeDef := ColorSchemes["dark"]
	scheme, _ := NewControlColorScheme(themeDef)
	ControlColors = scheme

	for name, test := range testCasesResultReason() {
		output := NewResultReasonRenderer(test.status, test.reason, test.width).Render()
		if output != test.expected {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v \ngot:\n %v\n", name, test.expected, output)
		}
	}
}
