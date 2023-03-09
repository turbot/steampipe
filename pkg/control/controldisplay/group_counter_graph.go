package controldisplay

import (
	"fmt"
	"math"
	"strings"
)

const counterGraphSegments = 10

type CounterGraphRenderer struct {
	failedControls int
	totalControls  int

	maxTotalControls int
	segmentSize      float64

	failedColorFunc colorFunc
}

type CounterGraphRendererOptions struct {
	FailedColorFunc colorFunc
}

func NewCounterGraphRenderer(failedControls, totalControls, maxTotalControls int, options CounterGraphRendererOptions) *CounterGraphRenderer {
	renderer := &CounterGraphRenderer{
		failedControls:   failedControls,
		totalControls:    totalControls,
		maxTotalControls: maxTotalControls,
		// there are 10 segments - determine the value of each segment
		segmentSize: float64(maxTotalControls) / float64(counterGraphSegments),

		failedColorFunc: options.FailedColorFunc,
	}
	return renderer
}

func (r CounterGraphRenderer) Render() string {
	// the graph has the format " [=======   ]"

	// if no controls have been run, return empty graph
	if r.maxTotalControls == 0 {
		return r.buildGraphString(0, 0, counterGraphSegments)
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
	if passSegments == 0 && r.failedControls < r.totalControls && totalSegments < counterGraphSegments {
		passSegments++
		totalSegments++
	}
	spaces := counterGraphSegments - totalSegments
	return r.buildGraphString(failSegments, passSegments, spaces)
}

func (r CounterGraphRenderer) buildGraphString(failSegments int, passSegments int, spaces int) string {
	str := fmt.Sprintf("%s%s%s%s%s",
		ControlColors.CountGraphBracket("["),
		r.failedColorFunc(strings.Repeat("=", failSegments)),
		ControlColors.CountGraphPass(strings.Repeat("=", passSegments)),
		strings.Repeat(" ", spaces),
		ControlColors.CountGraphBracket("]"))
	return str
}
