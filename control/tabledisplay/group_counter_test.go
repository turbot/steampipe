package tabledisplay

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

var testCasesCounter = map[string]counterTest{
	"1/10 max 1/10": {
		failedControls:    1,
		totalControls:     10,
		maxFailedControls: 1,
		maxTotalControls:  10,

		expected: fmt.Sprintf("%d %s %d", colorCountFail(1), colorCountDivider("/"), colorCountTotal(10)),
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

		expected: fmt.Sprintf("%2d %s %d", colorCountFail(1), colorCountDivider("/"), colorCountTotal(10)),
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

		expected: fmt.Sprintf("%2d %s %3d", colorCountFail(1), colorCountDivider("/"), colorCountTotal(10)),
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

		expected: fmt.Sprintf("%3d %s %5d", colorCountFail(1), colorCountDivider("/"), colorCountTotal(10)),
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

		expected: fmt.Sprintf("%3d %s %5d", colorCountFail(10), colorCountDivider("/"), colorCountTotal(500)),
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

		expected: fmt.Sprintf("%3d %s %s", colorCountFail(10), colorCountDivider("/"), colorCountTotal("1,000")),
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

		expected: fmt.Sprintf("%3d %s %s", colorCountZeroFail(0), colorCountZeroFailDivider("/"), colorCountTotalAllPassed("1,000")),
	},
}

func TestCounter(t *testing.T) {
	for name, test := range testCasesCounter {
		counter := NewCounterRenderer(test.failedControls, test.totalControls, test.maxFailedControls, test.maxTotalControls)
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
