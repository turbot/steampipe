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
		id:       "group title",
		width:    100,
		expected: fmt.Sprintf("%s", colorId("group title")),
	},

	"equal": {
		id:       "group title",
		width:    8,
		expected: fmt.Sprintf("%s", colorId("group title")),
	},
	"longer trim on space": {
		id:       "group title",
		width:    7,
		expected: fmt.Sprintf("%s", colorId("group …")),
	},
	"longer trim on char": {
		id:       "group title",
		width:    5,
		expected: fmt.Sprintf("%s", colorId("grou…")),
	},
}

func TestId(t *testing.T) {
	for name, test := range testCasesId {
		renderer := NewGroupTitleRenderer(test.id, test.width)
		output := renderer.String()

		if output != test.expected {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n%s \ngot:\n%s\n", name, test.expected, output)
		}
	}
}
