package controldisplay

import (
	"fmt"
	"testing"
)

type idTest struct {
	id       string
	width    int
	expected string
}

var testCasesId = map[string]idTest{
	"shorter": {
		id:       "group id",
		width:    100,
		expected: fmt.Sprintf("%s", colorId("group id")),
	},

	"equal": {
		id:       "group id",
		width:    8,
		expected: fmt.Sprintf("%s", colorId("group id")),
	},
	"longer trim on space": {
		id:       "group id",
		width:    7,
		expected: fmt.Sprintf("%s", colorId("group …")),
	},
	"longer trim on char": {
		id:       "group id",
		width:    5,
		expected: fmt.Sprintf("%s", colorId("grou…")),
	},
}

func TestId(t *testing.T) {
	for name, test := range testCasesId {
		renderer := NewGroupIdRenderer(test.id, test.width)
		output := renderer.String()

		if output != test.expected {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n%s \ngot:\n%s\n", name, test.expected, output)
		}
	}
}
