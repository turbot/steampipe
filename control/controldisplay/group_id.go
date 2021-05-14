package controldisplay

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
)

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
func (d GroupTitleRenderer) Render() (string, int) {
	truncatedId := helpers.TruncateString(d.title, d.width)
	length := len(truncatedId)
	str := fmt.Sprintf("%s", colorId(truncatedId))
	return str, length
}
