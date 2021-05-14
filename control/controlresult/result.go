package controlresult

import (
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// TODO could this be merged with ControlRun?
// Result is a struct representing the result of a control run
// It will contain one or more result rows (i.e. for one or more resources)
type Result struct {
	ControlId   string            `json:"control_id"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	Tags        map[string]string `json:"tags"`
	Title       string            `json:"title"`
	Rows        []*ResultRow      `json:"results"`
}

func (r *Result) addResultRow(row *ResultRow) {
	r.Rows = append(r.Rows, row)
}

func NewResult(control *modconfig.Control) *Result {
	res := &Result{
		ControlId:   control.Name(),
		Description: typehelpers.SafeString(control.Description),
		Severity:    typehelpers.SafeString(control.Severity),
		Title:       typehelpers.SafeString(control.Title),
		Tags:        control.GetTags(),
		Rows:        []*ResultRow{},
	}

	return res
}
