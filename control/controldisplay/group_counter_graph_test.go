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
var testCasesCounterGraph = map[string]counterGraphTest{

	"0/10 max 20  (zero pass)": {
		maxTotalControls: 20,
		failedControls:   0,
		totalControls:    10,
		// each segment is 12 -> 5 success, 5 blank
		//[xxxxx     ]
		expectedString: fmt.Sprintf(" [%s%s%s]",
			colorCountGraphFail(strings.Repeat("=", 0)),
			colorCountGraphPass(strings.Repeat("=", 5)),
			strings.Repeat(" ", 5)),
	},

	"1/10 max 20 (less than 1 segment failed)": {
		maxTotalControls: 20,
		totalControls:    10,
		failedControls:   1,
		// each segment is 2 -> 1 fail, 4 success, 5 blank
		//	[Xxxxx     ]
		expectedString: fmt.Sprintf(" [%s%s%s]",
			colorCountGraphFail(strings.Repeat("=", 1)),
			colorCountGraphPass(strings.Repeat("=", 4)),
			strings.Repeat(" ", 5)),
	},
	"2/10 max 20 (exactly 1 segment failed)": {
		maxTotalControls: 20,
		totalControls:    10,
		failedControls:   2,
		// each segment is 2 -> 1 fail, 4 success, 5 blank
		//	[Xxxxx     ]
		expectedString: fmt.Sprintf(" [%s%s%s]",
			colorCountGraphFail(strings.Repeat("=", 1)),
			colorCountGraphPass(strings.Repeat("=", 4)),
			strings.Repeat(" ", 5)),
	},
	"3/10 max 20 (more than 1 segment failed)": {
		maxTotalControls: 20,
		totalControls:    10,
		failedControls:   3,
		// each segment is 3 -> 2 fail, 3 success, 5 blank
		//	[Xxxxx     ]
		expectedString: fmt.Sprintf(" [%s%s%s]",
			colorCountGraphFail(strings.Repeat("=", 2)),
			colorCountGraphPass(strings.Repeat("=", 3)),
			strings.Repeat(" ", 5)),
	},

	"0/12 max 28 (zero pass)": {
		maxTotalControls: 28,
		totalControls:    12,
		failedControls:   0,
		// segment=2.8 -> 0 fail, 5 success, 5 blank
		// [Xxxxx     ]
		expectedString: fmt.Sprintf(" [%s%s%s]",
			colorCountGraphFail(strings.Repeat("=", 0)),
			colorCountGraphPass(strings.Repeat("=", 5)),
			strings.Repeat(" ", 5)),
	},

	"1/12 max 28 (less than 1 segment failed)": {
		maxTotalControls: 28,
		totalControls:    12,
		failedControls:   1,
		// segment=2.8 -> 1 fail, 4 success, 5 blank
		// [Xxxxx     ]
		expectedString: fmt.Sprintf(" [%s%s%s]",
			colorCountGraphFail(strings.Repeat("=", 1)),
			colorCountGraphPass(strings.Repeat("=", 4)),
			strings.Repeat(" ", 5)),
	},

	"3/12 max 28 (more than 1 segment failed)": {
		maxTotalControls: 28,
		totalControls:    12,
		failedControls:   3,
		// segment=2.8 -> 2 fail, 3 success, 5 blank
		// [Xxxxx     ]
		expectedString: fmt.Sprintf(" [%s%s%s]",
			colorCountGraphFail(strings.Repeat("=", 2)),
			colorCountGraphPass(strings.Repeat("=", 3)),
			strings.Repeat(" ", 5)),
	},

	" 0/17 max 51 (zero pass)": {
		maxTotalControls: 51,
		totalControls:    17,
		failedControls:   0,
		// segment=5.1 -> 0 fail, 4 success, 6 blank
		// [xxxx      ]
		expectedString: fmt.Sprintf(" [%s%s%s]",
			colorCountGraphFail(strings.Repeat("=", 0)),
			colorCountGraphPass(strings.Repeat("=", 4)),
			strings.Repeat(" ", 6)),
	},
	"4/17 max 51 (less than 1 segment failed)": {
		maxTotalControls: 51,
		totalControls:    17,
		failedControls:   4,
		// segment=5.1 -> 1 fail, 3 success, 6 blank
		// [xxxx      ]
		expectedString: fmt.Sprintf(" [%s%s%s]",
			colorCountGraphFail(strings.Repeat("=", 1)),
			colorCountGraphPass(strings.Repeat("=", 3)),
			strings.Repeat(" ", 6)),
	},
	"6/17 max 51 (more than 1 segment failed)": {
		maxTotalControls: 51,
		totalControls:    17,
		failedControls:   6,
		// segment=5.1 -> 2 fail, 2 success, 6 blank
		// [xxxx      ]
		expectedString: fmt.Sprintf(" [%s%s%s]",
			colorCountGraphFail(strings.Repeat("=", 2)),
			colorCountGraphPass(strings.Repeat("=", 2)),
			strings.Repeat(" ", 6)),
	},
}

func TestCounterGraph(t *testing.T) {
	for name, test := range testCasesCounterGraph {
		counterGraph := NewCounterGraphRenderer(test.failedControls, test.totalControls, test.maxTotalControls)
		output := counterGraph.Render()
		if output != test.expectedString {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n %s, \ngot:\n %s\n", name, test.expectedString, output)
		}
	}
}
