package tabledisplay

import (
	"fmt"
	"strings"
)

type SpacerRenderer struct {
	width int
}

func NewSpacerRenderer(width int) *SpacerRenderer {
	return &SpacerRenderer{width}
}

// String returns a divider string os format: "......."
func (d SpacerRenderer) String() string {
	return fmt.Sprintf("%s", colorSpacer(strings.Repeat(".", d.width)))
}
