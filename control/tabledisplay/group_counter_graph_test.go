package tabledisplay

import (
	"fmt"
	"testing"
)

type counterGraphTest struct {
	failedControls   int
	totalControls    int
	maxTotalControls int

	expectedString string
}

var testCasesCounterGraph = map[string]counterGraphTest{
	"1/10 max 10": {
		failedControls:   1,
		totalControls:    10,
		maxTotalControls: 10,
		// each segment is 1
		expectedString: fmt.Sprintf("[%s%s]", colorCountGraphFail("="), colorCountGraphPass("=========")),
	},

	"1/10 max 100": {
		failedControls:   1,
		totalControls:    10,
		maxTotalControls: 100,
		// each segment is 10 - 1 fail segment, 1 pass segment, 8 spaces
		expectedString: fmt.Sprintf("[%s%s        ]", colorCountGraphFail("="), colorCountGraphPass("=")),
	},
	"10/10 max 100": {
		failedControls:   10,
		totalControls:    10,
		maxTotalControls: 100,
		// each segment is 10 - 1 fail segment, 0 pass segment, 9 spaces
		expectedString: fmt.Sprintf("[%s         ]", colorCountGraphFail("=")),
	},
	"1/10 max 1000": {
		failedControls:   1,
		totalControls:    10,
		maxTotalControls: 1000,
		// each segment is 100 - 1 fail segment, 1 pass segment, 8 spaces
		expectedString: fmt.Sprintf("[%s%s        ]", colorCountGraphFail("="), colorCountGraphPass("=")),
	},
	"10/200 max 1000": {
		failedControls:   20,
		totalControls:    200,
		maxTotalControls: 1000,
		// each segment is 100 - 1 fail segment, 2 pass segments, 7 spaces
		expectedString: fmt.Sprintf("[%s%s       ]", colorCountGraphFail("="), colorCountGraphPass("==")),
	},
	"100/500 max 1000": {
		failedControls:   100,
		totalControls:    500,
		maxTotalControls: 1000,
		// each segment is 100 - 1 fail segment, 4 pass segments, 5 spaces
		expectedString: fmt.Sprintf("[%s%s     ]", colorCountGraphFail("="), colorCountGraphPass("====")),
	},
	"0/500 max 1000": {
		failedControls:   0,
		totalControls:    500,
		maxTotalControls: 1000,
		// each segment is 100 - 0 fail segment, 5 pass segments, 5 spaces
		expectedString: fmt.Sprintf("[%s     ]", colorCountGraphPass("=====")),
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
