package controldisplay

import (
	"fmt"
	"testing"
)

type resultStatusTest struct {
	status   string
	expected string
}

var testCasesResultStatus = map[string]resultStatusTest{
	"error": {
		status:   "error",
		expected: fmt.Sprintf("%-6s", statusColors["error"]("error")),
	},
	"ok": {
		status:   "ok",
		expected: fmt.Sprintf("%-6s", statusColors["ok"]("ok")),
	},
}

func TestResultStatus(t *testing.T) {
	for name, test := range testCasesResultStatus {
		output := NewResultStatusRenderer(test.status).Render()
		if output != test.expected {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v \ngot:\n %v\n", name, test.expected, output)
		}
	}
}
