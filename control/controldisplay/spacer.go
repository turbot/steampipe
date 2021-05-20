package controldisplay

import (
	"fmt"
	"log"
	"strings"
)

type SpacerRenderer struct {
	width int
}

func NewSpacerRenderer(width int) *SpacerRenderer {
	return &SpacerRenderer{width}
}

// Render returns a divider string of format: "....... "
// NOTE: adds a trailing space
func (r SpacerRenderer) Render() string {
	log.Println("[TRACE] begin spacer render")
	defer log.Println("[TRACE] end spacer render")

	if r.width <= 0 {
		log.Printf("[WARN] spacer renderer has width of %d\n", r.width)
		return ""
	}
	// we always have a trailing space
	if r.width == 1 {
		return " "
	}

	// allow for trailing space
	numberOfDots := r.width - 1
	return fmt.Sprintf("%s ", ControlColors.Spacer(strings.Repeat(".", numberOfDots)))
}
