package tabledisplay

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
)

type GroupIdRenderer struct {
	id string

	width int
}

func NewGroupIdRenderer(id string, width int) *GroupIdRenderer {
	return &GroupIdRenderer{
		id:    id,
		width: width,
	}
}

// String returns the id, truncated to the max length if necessary
func (d GroupIdRenderer) String() string {
	str := fmt.Sprintf("%s", colorId(helpers.TruncateString(d.id, d.width)))
	//fmt.Println(str)
	return str
}
