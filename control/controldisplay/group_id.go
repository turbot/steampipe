package controldisplay

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
)

type GroupIdRenderer struct {
	id    string
	width int
}

func NewGroupIdRenderer(id string, width int) *GroupIdRenderer {
	return &GroupIdRenderer{
		id:    id,
		width: width,
	}
}

// String returns the id, truncated to the max length if necessary
func (d GroupIdRenderer) String() (string, int) {
	truncatedId := helpers.TruncateString(d.id, d.width)
	length := len(truncatedId)
	str := fmt.Sprintf("%s", colorId(truncatedId))
	return str, length
}
