package controldisplay

import (
	"fmt"
	"strings"
	"testing"
)

type counterGraphTest struct {
	failedControls   int
	totalControls    int
	maxTotalControls int

	expectedString string
}

// when calculating chart elements, round UP success to next segment, round DOWN space count
func testCasesCounterGraph() map[string]counterGraphTest {
	return map[string]counterGraphTest{

		"0/10 max 20  (zero pass)": {
			maxTotalControls: 20,
			failedControls:   0,
			totalControls:    10,
			// each segment is 12 -> 5 success, 5 blank
			//[xxxxx     ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 0)),
				ControlColors.CountGraphPass(strings.Repeat("=", 5)),
				strings.Repeat(" ", 5),
				ControlColors.CountGraphBracket("]")),
		},

		"1/10 max 20 (less than 1 segment failed)": {
			maxTotalControls: 20,
			totalControls:    10,
			failedControls:   1,
			// each segment is 2 -> 1 fail, 4 success, 5 blank
			//	[Xxxxx     ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 1)),
				ControlColors.CountGraphPass(strings.Repeat("=", 4)),
				strings.Repeat(" ", 5),
				ControlColors.CountGraphBracket("]")),
		},
		"2/10 max 20 (exactly 1 segment failed)": {
			maxTotalControls: 20,
			totalControls:    10,
			failedControls:   2,
			// each segment is 2 -> 1 fail, 4 success, 5 blank
			//	[Xxxxx     ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 1)),
				ControlColors.CountGraphPass(strings.Repeat("=", 4)),
				strings.Repeat(" ", 5),
				ControlColors.CountGraphBracket("]")),
		},
		"3/10 max 20 (more than 1 segment failed)": {
			maxTotalControls: 20,
			totalControls:    10,
			failedControls:   3,
			// each segment is 3 -> 2 fail, 3 success, 5 blank
			//	[Xxxxx     ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 2)),
				ControlColors.CountGraphPass(strings.Repeat("=", 3)),
				strings.Repeat(" ", 5),
				ControlColors.CountGraphBracket("]")),
		},

		"0/12 max 28 (zero pass)": {
			maxTotalControls: 28,
			totalControls:    12,
			failedControls:   0,
			// segment=2.8 -> 0 fail, 5 success, 5 blank
			// [Xxxxx     ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 0)),
				ControlColors.CountGraphPass(strings.Repeat("=", 5)),
				strings.Repeat(" ", 5),
				ControlColors.CountGraphBracket("]")),
		},

		"1/12 max 28 (less than 1 segment failed)": {
			maxTotalControls: 28,
			totalControls:    12,
			failedControls:   1,
			// segment=2.8 -> 1 fail, 4 success, 5 blank
			// [Xxxxx     ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 1)),
				ControlColors.CountGraphPass(strings.Repeat("=", 4)),
				strings.Repeat(" ", 5),
				ControlColors.CountGraphBracket("]")),
		},

		"3/12 max 28 (more than 1 segment failed)": {
			maxTotalControls: 28,
			totalControls:    12,
			failedControls:   3,
			// segment=2.8 -> 2 fail, 3 success, 5 blank
			// [Xxxxx     ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 2)),
				ControlColors.CountGraphPass(strings.Repeat("=", 3)),
				strings.Repeat(" ", 5),
				ControlColors.CountGraphBracket("]")),
		},

		" 0/17 max 51 (zero pass)": {
			maxTotalControls: 51,
			totalControls:    17,
			failedControls:   0,
			// segment=5.1 -> 0 fail, 4 success, 6 blank
			// [xxxx      ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 0)),
				ControlColors.CountGraphPass(strings.Repeat("=", 4)),
				strings.Repeat(" ", 6),
				ControlColors.CountGraphBracket("]")),
		},
		"4/17 max 51 (less than 1 segment failed)": {
			maxTotalControls: 51,
			totalControls:    17,
			failedControls:   4,
			// segment=5.1 -> 1 fail, 3 success, 6 blank
			// [xxxx      ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 1)),
				ControlColors.CountGraphPass(strings.Repeat("=", 3)),
				strings.Repeat(" ", 6),
				ControlColors.CountGraphBracket("]")),
		},
		"6/17 max 51 (more than 1 segment failed)": {
			maxTotalControls: 51,
			totalControls:    17,
			failedControls:   6,
			// segment=5.1 -> 2 fail, 2 success, 6 blank
			// [xxxx      ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 2)),
				ControlColors.CountGraphPass(strings.Repeat("=", 2)),
				strings.Repeat(" ", 6),
				ControlColors.CountGraphBracket("]")),
		},
		"71/80 max 560 rounding error": {
			maxTotalControls: 560,
			totalControls:    80,
			failedControls:   71,
			// segment=5.1 -> 2 fail, 2 success, 6 blank
			// [xxxx      ]
			expectedString: fmt.Sprintf("%s%s%s%s%s",
				ControlColors.CountGraphBracket("["),
				ControlColors.CountGraphFail(strings.Repeat("=", 2)),
				ControlColors.CountGraphPass(strings.Repeat("=", 1)),
				strings.Repeat(" ", 7),
				ControlColors.CountGraphBracket("]")),
		},
	}
}

func TestCounterGraph(t *testing.T) {
	themeDef := ColorSchemes["dark"]
	scheme, _ := NewControlColorScheme(themeDef)
	ControlColors = scheme

	for name, test := range testCasesCounterGraph() {
		counterGraph := NewCounterGraphRenderer(test.failedControls, test.totalControls, test.maxTotalControls, CounterGraphRendererOptions{FailedColorFunc: ControlColors.CountGraphFail})
		output := counterGraph.Render()
		if output != test.expectedString {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %s, \ngot:\n %s\n", name, test.expectedString, output)
		}
	}
}
