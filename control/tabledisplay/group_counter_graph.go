package tabledisplay

import (
	"fmt"
	"strings"
)

type CounterGraphRenderer struct {
	failedControls int
	totalControls  int

	maxTotalControls int
	segmentSize      int
}

func NewCounterGraphRenderer(failedControls, totalControls, maxTotalControls int) *CounterGraphRenderer {
	return &CounterGraphRenderer{
		failedControls:   failedControls,
		totalControls:    totalControls,
		maxTotalControls: maxTotalControls,
		// there are 10 segments - determine the value of each segment
		segmentSize: maxTotalControls / 10,
	}
}

func (d CounterGraphRenderer) Render() (string, int) {
	// the graph has the format " [=======   ]"
	// the graph is 10 segments long, so length is always 13
	length := 13

	// if each segment is 10 controls, count 1-10 => 1 segment, 11-20 => 2 segments
	var failSegments, passSegments, spaces int
	// TODO I'm sure we can tidy this up to avoid special cases
	if d.failedControls == 0 {
		passSegments = ((d.totalControls - 1) / d.segmentSize) + 1
		spaces = 10 - passSegments
		str := fmt.Sprintf(" [%s%s]",
			colorCountGraphPass(strings.Repeat("=", passSegments)),
			strings.Repeat(" ", spaces))
		return str, length
	}

	if d.failedControls == d.totalControls {
		failSegments = ((d.totalControls - 1) / d.segmentSize) + 1
		spaces = 10 - failSegments
		str := fmt.Sprintf(" [%s%s]",
			colorCountGraphFail(strings.Repeat("=", failSegments)),
			strings.Repeat(" ", spaces))
		return str, length
	}

	// so we have both pass and fail segments
	failSegments = ((d.failedControls - 1) / d.segmentSize) + 1
	passSegments = ((d.totalControls - d.failedControls - 1) / d.segmentSize) + 1

	// can happen with rounding up
	if passSegments+failSegments > 10 {
		passSegments = 10 - failSegments
	}
	spaces = 10 - failSegments - passSegments

	str := fmt.Sprintf(" [%s%s%s]",
		colorCountGraphFail(strings.Repeat("=", failSegments)),
		colorCountGraphPass(strings.Repeat("=", passSegments)),
		strings.Repeat(" ", spaces))

	return str, length

}
