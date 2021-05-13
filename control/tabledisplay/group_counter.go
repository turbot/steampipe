package tabledisplay

import (
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type CounterRenderer struct {
	failedControls int
	totalControls  int

	maxFailedControls int
	maxTotalControls  int
}

func NewCounterRenderer(failedControls, totalControls, maxFailedControls, maxTotalControls int) *CounterRenderer {
	return &CounterRenderer{
		failedControls:    failedControls,
		totalControls:     totalControls,
		maxFailedControls: maxFailedControls,
		maxTotalControls:  maxTotalControls,
	}
}

/* Render returns the counter string in format "<failed> / <total>.

The alignment depends on the maximum failed and maximum total parameters, as the counters are aligned as follows:
"  3 /   123"
" 13 /    23"
"111 /   123"
"  1 /     4"
"  1 / 1,020"

*/
func (d CounterRenderer) Render() (string, int) {
	p := message.NewPrinter(language.English)
	// get strings for fails and total - format with commas for thousands
	failedString := p.Sprintf("%d", d.failedControls)
	totalString := p.Sprintf("%d", d.totalControls)
	// get max strings - format with commas for thousands
	maxFailedString := p.Sprintf("%d", d.maxFailedControls)
	maxTotalString := p.Sprintf("%d", d.maxTotalControls)

	// calculate the width of the fails and total columns
	failedWidth := len(maxFailedString)
	totalWidth := len(maxTotalString)

	// build format string, specifying widths of failedString and totalString
	// this will generate a format string like: "%3s / %4s"
	formatString := fmt.Sprintf("%%%ds %%s %%%ds", failedWidth, totalWidth)

	// calculate length - the 3 is the " / "
	length := failedWidth + totalWidth + 3

	if d.failedControls == 0 {
		return fmt.Sprintf(formatString, colorCountZeroFail(failedString), colorCountZeroFailDivider("/"), colorCountTotalAllPassed(totalString)), length
	}

	str := fmt.Sprintf(formatString, colorCountFail(failedString), colorCountDivider("/"), colorCountTotal(totalString))
	//fmt.Println(str)
	return str, length
}
