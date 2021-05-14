package controldisplay

type Result struct {
	status string

	reason        string
	totalControls int

	// screen width
	width int
}
