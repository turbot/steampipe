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
	Title  *ReportChartAxisTitle `cty:"title" hcl:"title,block" json:"title,omitempty"`
	Labels *ReportChartLabels    `cty:"title" hcl:"labels,block" json:"labels,omitempty"`
}

func (x ReportChartAxesX) Equals(other *ReportChartAxesX) bool {
	if other == nil {
		return false
	}

	if x.Title != nil {
		if !x.Title.Equals(other.Title) {
			return false
		}
	} else if other.Title != nil {
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
	Title  *ReportChartAxisTitle `cty:"title" hcl:"title,block" json:"title,omitempty"`
	Labels *ReportChartLabels    `cty:"labels" hcl:"labels,block" json:"labels,omitempty"`
	Min    *int                  `cty:"min" hcl:"min" json:"min,omitempty"`
	Max    *int                  `cty:"max" hcl:"max" json:"max,omitempty"`
}

func (y ReportChartAxesY) Equals(other *ReportChartAxesY) bool {
	if other == nil {
		return false
	}

	if !(utils.SafeIntEqual(y.Min, other.Min) &&
		utils.SafeIntEqual(y.Max, other.Max)) {
		return false
	}

	if y.Title != nil {
		if !y.Title.Equals(other.Title) {
			return false
		}
	} else if other.Title != nil {
		return false
	}

	if y.Labels != nil {
		return y.Labels.Equals(other.Labels)
	} else if other.Labels != nil {
		return false
	}
	return true
}

type ReportChartAxisTitle struct {
	Display *string `cty:"display" hcl:"display" json:"display,omitempty"`
	Align   *string `cty:"align" hcl:"align" json:"align,omitempty"`
	Value   *string `cty:"value" hcl:"value" json:"value,omitempty"`
}

func (t ReportChartAxisTitle) Equals(other *ReportChartAxisTitle) bool {
	if other == nil {
		return false
	}

	if !(utils.SafeStringsEqual(t.Display, other.Display) &&
		utils.SafeStringsEqual(t.Align, other.Align) &&
		utils.SafeStringsEqual(t.Value, other.Value)) {
		return false
	}

	return true
}
