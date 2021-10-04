package controldisplay

import (
	"fmt"
	"testing"

	"github.com/logrusorgru/aurora"
)

var Gray = aurora.Gray
var Red = aurora.Red

type counterTest struct {
	failedControls    int
	totalControls     int
	maxFailedControls int
	maxTotalControls  int

	expected string
}

func testCasesCounter() map[string]counterTest {
	return map[string]counterTest{
		"1/10 max 1/10": {
			failedControls:    1,
			totalControls:     10,
			maxFailedControls: 1,
			maxTotalControls:  10,

			expected: fmt.Sprintf("%d %s %d", ControlColors.CountFail(1), ControlColors.CountDivider("/"), ControlColors.CountTotal(10)),
		},
		"1/10 max 10/10": {
			/*
				" 1 / 10" <
				"10 / 10"
			*/
			failedControls:    1,
			totalControls:     10,
			maxFailedControls: 10,
			maxTotalControls:  10,

			expected: fmt.Sprintf("%2d %s %d", ControlColors.CountFail(1), ControlColors.CountDivider("/"), ControlColors.CountTotal(10)),
		},
		"1/10 max 10/100": {
			/*
				" 1 /  10" <
				"10 / 100"
			*/
			failedControls:    1,
			totalControls:     10,
			maxFailedControls: 10,
			maxTotalControls:  100,

			expected: fmt.Sprintf("%2d %s %3d", ControlColors.CountFail(1), ControlColors.CountDivider("/"), ControlColors.CountTotal(10)),
		},
		"1/10 max 100/1000": {
			/*
				"  1 /    10" <
				"100 / 1,000"
			*/
			failedControls:    1,
			totalControls:     10,
			maxFailedControls: 100,
			maxTotalControls:  1000,

			expected: fmt.Sprintf("%3d %s %5d", ControlColors.CountFail(1), ControlColors.CountDivider("/"), ControlColors.CountTotal(10)),
		},
		"10/500 max 100/1000": {
			/*
				" 10 /   500" <
				"100 / 1,000"
			*/
			failedControls:    10,
			totalControls:     500,
			maxFailedControls: 100,
			maxTotalControls:  1000,

			expected: fmt.Sprintf("%3d %s %5d", ControlColors.CountFail(10), ControlColors.CountDivider("/"), ControlColors.CountTotal(500)),
		},
		"10/1000 max 100/1000": {
			/*
				" 10 / 1,000" <
				"100 / 1,000"
			*/
			failedControls:    10,
			totalControls:     1000,
			maxFailedControls: 100,
			maxTotalControls:  1000,

			expected: fmt.Sprintf("%3d %s %s", ControlColors.CountFail(10), ControlColors.CountDivider("/"), ControlColors.CountTotal("1,000")),
		},
		"0/1000 max 100/1000": {
			/*
				"  0 / 1,000" <
				"100 / 1,000"
			*/
			failedControls:    0,
			totalControls:     1000,
			maxFailedControls: 100,
			maxTotalControls:  1000,

			expected: fmt.Sprintf("%3d %s %s", ControlColors.CountZeroFail(0), ControlColors.CountZeroFailDivider("/"), ControlColors.CountTotalAllPassed("1,000")),
		},
	}
}

func TestCounter(t *testing.T) {
	themeDef := ColorSchemes["dark"]
	scheme, _ := NewControlColorScheme(themeDef)
	ControlColors = scheme

	for name, test := range testCasesCounter() {
		counter := NewCounterRenderer(test.failedControls, test.totalControls, test.maxFailedControls, test.maxTotalControls, CounterRendererOptions{AddLeadingSpace: true})
		output := counter.Render()

		if output != test.expected {
			t.Errorf("Test: '%s'' FAILED : \nexpected:\n%s \ngot:\n%s\n", name, test.expected, output)
		}
	}
}

//func TestColor(t *testing.T) {
//
//	// 	72
//	for i := uint8(16); i != 15; i += 36 {
//		if i > 231 {
//			i -= 231
//		}
//		fmt.Println(aurora.Index(i, fmt.Sprintf("COLOR %d", i)))
//	}
//
//}
