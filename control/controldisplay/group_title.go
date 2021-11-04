package controldisplay

import (
	"fmt"
	"log"

	"github.com/turbot/go-kit/helpers"
)

// There will always be a space after the title, even if the title is empty
const minimumGroupTitleWidth = 1

type GroupTitleRenderer struct {
	title string
	width int
}

func NewGroupTitleRenderer(title string, width int) *GroupTitleRenderer {
	return &GroupTitleRenderer{
		title: title,
		width: width,
	}
}

// Render returns the title, truncated to the max length if necessary
// NOTE: adds a trailing space
func (r GroupTitleRenderer) Render() string {
	if r.width <= 0 {
		// this should never happen, since the minimum width is set by the formatter
		log.Printf("[WARN] group renderer has width of %d\n", r.width)
		return ""
	}
	// allow room for trailing space
	truncatedId := helpers.TruncateString(r.title, r.width-1)
	str := fmt.Sprintf("%s ", ControlColors.GroupTitle(truncatedId))
	return str
}
