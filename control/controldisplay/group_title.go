package controldisplay

import (
	"fmt"
	"log"

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
func (r GroupTitleRenderer) Render() string {
	log.Println("[TRACE] begin group title render")
	defer log.Println("[TRACE] end group title render")

	if r.width <= 0 {
		log.Printf("[WARN] group renderer has width of %d\n", r.width)
		return ""
	}
	truncatedId := helpers.TruncateString(r.title, r.width)
	str := fmt.Sprintf("%s", colorId(truncatedId))
	return str
}
