package controldisplay

import (
	"fmt"
	"log"
	"math"
	"strings"
)

type CounterGraphRenderer struct {
	failedControls int
	totalControls  int

	maxTotalControls int
	segmentSize      float64
}

func NewCounterGraphRenderer(failedControls, totalControls, maxTotalControls int) *CounterGraphRenderer {
	renderer := &CounterGraphRenderer{
		failedControls:   failedControls,
		totalControls:    totalControls,
		maxTotalControls: maxTotalControls,
		// there are 10 segments - determine the value of each segment
		segmentSize: float64(maxTotalControls) / 10.0,
	}
	return renderer
}

func (r CounterGraphRenderer) Render() string {
	log.Println("[TRACE] begin counter graph render")
	defer log.Println("[TRACE] end counter graph render")

	// the graph has the format " [=======   ]"

	// if no controls have been run, return empty graph
	if r.maxTotalControls == 0 {
		return r.buildGraphString(0, 0, 10)
	}
	// if each segment is 10 controls, count 1-10 => 1 segment, 11-20 => 2 segments
	var failSegments int

	if r.failedControls == 0 {
		failSegments = 0
	} else {
		// if there is a remainder round up
		failSegments = int(math.Ceil(float64(r.failedControls) / r.segmentSize))

	}
	totalSegments := int(math.Ceil(float64(r.totalControls) / r.segmentSize))

	passSegments := totalSegments - failSegments
	// allow for pass being rounded down to zero
	// if there are any successful runs, but there is no room for a successful bar,
	// increment totalSegments to allow room
	if passSegments == 0 && r.failedControls < r.totalControls && totalSegments < 10 {
		passSegments++
		totalSegments++
	}
	spaces := 10 - totalSegments
	return r.buildGraphString(failSegments, passSegments, spaces)
}

func (r CounterGraphRenderer) buildGraphString(failSegments int, passSegments int, spaces int) string {
	str := fmt.Sprintf("%s%s%s%s%s",
		ControlColors.CountGraphBracket("["),
		ControlColors.CountGraphFail(strings.Repeat("=", failSegments)),
		ControlColors.CountGraphPass(strings.Repeat("=", passSegments)),
		strings.Repeat(" ", spaces),
		ControlColors.CountGraphBracket("]"))
	return str
}
