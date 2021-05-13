package tabledisplay

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

var testCasesResultReason = map[string]resultReasonTest{
	"error fit": {
		status:   "error",
		reason:   "short error reason",
		width:    100,
		expected: fmt.Sprintf("%s", reasonColors["error"]("short error reason")),
	},
	"ok fit": {
		status:   "ok",
		reason:   "short ok reason",
		width:    100,
		expected: fmt.Sprintf("%s", reasonColors["ok"]("short ok reason")),
	},
	"error truncate": {
		status:   "error",
		reason:   "long error reason is very long and goes on and on",
		width:    40,
		expected: fmt.Sprintf("%s", reasonColors["error"]("long error reason is very long and goes…")),
	},
}

func TestResultReason(t *testing.T) {
	for name, test := range testCasesResultReason {
		output := NewResultReasonRenderer(test.status, test.reason, test.width).String()
		if output != test.expected {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v \ngot:\n %v\n", name, test.expected, output)
		}
	}
}
