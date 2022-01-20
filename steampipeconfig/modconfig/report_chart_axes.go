package modconfig

import (
	"github.com/turbot/steampipe/utils"
)

type ReportChartAxes struct {
	X *ReportChartAxesX `cty:"x" hcl:"x,block" json:"x,omitempty"`
	Y *ReportChartAxesY `cty:"y" hcl:"y,block" json:"y,omitempty"`
}

func (a ReportChartAxes) Equals(other *ReportChartAxes) bool {
	if other == nil {
		return false
	}

	if a.X != nil {
		if other.X == nil {
			return false
		}
		if !a.X.Equals(other.X) {
			return false
		}

	} else if other.X != nil {
		return false
	}

	if a.Y != nil {
		if other.Y == nil {
			return false
		}
		if !a.Y.Equals(other.Y) {
			return false
		}

	} else if other.Y != nil {
		return false
	}

	return true

}

type ReportChartAxesX struct {
	Title  *string            `cty:"title" hcl:"title" json:"title,omitempty"`
	Labels *ReportChartLabels `cty:"title" hcl:"labels,block" json:"labels,omitempty"`
}

func (x ReportChartAxesX) Equals(other *ReportChartAxesX) bool {
	if other == nil {
		return false
	}

	if !utils.SafeStringsEqual(x.Title, other.Title) {
		return false
	}
	if x.Labels != nil {
		return x.Labels.Equals(other.Labels)
	} else if other.Labels != nil {
		return false
	}

	return true
}

type ReportChartAxesY struct {
	Title  *string            `cty:"title" hcl:"title" json:"title,omitempty"`
	Labels *ReportChartLabels `cty:"labels" hcl:"labels,block" json:"labels,omitempty"`
	Min    *int               `cty:"min" hcl:"min" json:"min,omitempty"`
	Max    *int               `cty:"max" hcl:"max" json:"max,omitempty"`
	Steps  *int               `cty:"steps" hcl:"steps" json:"steps,omitempty"`
}

func (y ReportChartAxesY) Equals(other *ReportChartAxesY) bool {
	if other == nil {
		return false
	}

	if !(utils.SafeStringsEqual(y.Title, other.Title) &&
		utils.SafeIntEqual(y.Min, other.Min) &&
		utils.SafeIntEqual(y.Max, other.Max) &&
		utils.SafeIntEqual(y.Steps, other.Steps)) {
		return false
	}

	if y.Labels != nil {
		return y.Labels.Equals(other.Labels)
	} else if other.Labels != nil {
		return false
	}
	return true
}
