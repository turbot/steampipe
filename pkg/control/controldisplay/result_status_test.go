package controldisplay

import (
	"fmt"
	"testing"
)

type resultStatusTest struct {
	status   string
	expected string
}

func testCasesResultStatus() map[string]resultStatusTest {
	return map[string]resultStatusTest{
		"error": {
			status:   "error",
			expected: fmt.Sprintf("%-6s", ControlColors.StatusColors["error"]("ERROR: ")),
		},
		"ok": {
			status:   "ok",
			expected: fmt.Sprintf("%-6s", ControlColors.StatusColors["ok"]("OK   : ")),
		},
	}
}

func TestResultStatus(t *testing.T) {
	themeDef := ColorSchemes["plain"]
	scheme, _ := NewControlColorScheme(themeDef)
	ControlColors = scheme
	for name, test := range testCasesResultStatus() {
		output := NewResultStatusRenderer(test.status).Render()
		if output != test.expected {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %v \ngot:\n %v\n", name, test.expected, output)
		}
	}
}
