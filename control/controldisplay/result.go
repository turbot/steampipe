package controldisplay

import "fmt"

type ResultRenderer struct {
	status string
	reason string

	// TODO dimensions
	// screen width
	width int
}

func NewResultRenderer(status, reason string, width int) *ResultRenderer {
	return &ResultRenderer{
		status: status,
		reason: reason,
		width:  width,
	}
}

func (r ResultRenderer) Render() string {
	status := NewResultStatusRenderer(r.status)
	statusString, statusWidth := status.Render()

	// figure out how much width we have available for the reason
	availableWidth := r.width - statusWidth

	// for now give this all to reason
	// TODO dimensions
	// now availableWidth is all we have - if it is not enough we need to truncate the reason
	reasonString, reasonWidth := NewResultReasonRenderer(r.status, r.reason, availableWidth).Render()

	// is there any room for a spacer

	spacerWidth := availableWidth - reasonWidth
	var spacerString string
	if spacerWidth > 0 {
		spacerString, _ = NewSpacerRenderer(spacerWidth).Render()
	}

	// now put these all together
	str := fmt.Sprintf("%s%s%s", statusString, reasonString, spacerString)
	return str
}
