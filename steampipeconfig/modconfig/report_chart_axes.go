package modconfig

import (
	"github.com/turbot/steampipe/utils"
)

type ReportChartAxes struct {
	X *ReportChartAxesX `cty:"x" hcl:"x,block" json:"x,omitempty"`
	Y *ReportChartAxesY `cty:"y" hcl:"y,block" json:"y,omitempty"`
}

func (a *ReportChartAxes) Equals(other *ReportChartAxes) bool {
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

func (a *ReportChartAxes) Merge(other *ReportChartAxes) {
	if other == nil {
		return
	}
	if a.X == nil {
		a.X = other.X
	} else {
		a.X.Merge(other.X)
	}
	if a.Y == nil {
		a.Y = other.Y
	} else {
		a.Y.Merge(other.Y)
	}
}

type ReportChartAxesX struct {
	Title  *ReportChartAxisTitle `cty:"title" hcl:"title,block" json:"title,omitempty"`
	Labels *ReportChartLabels    `cty:"labels" hcl:"labels,block" json:"labels,omitempty"`
}

func (x *ReportChartAxesX) Equals(other *ReportChartAxesX) bool {
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

func (x *ReportChartAxesX) Merge(other *ReportChartAxesX) {
	if x.Title == nil {
		x.Title = other.Title
	} else {
		x.Title.Merge(other.Title)
	}
	if x.Labels == nil {
		x.Labels = other.Labels
	} else {
		x.Labels.Merge(other.Labels)
	}
}

type ReportChartAxesY struct {
	Title  *ReportChartAxisTitle `cty:"title" hcl:"title,block" json:"title,omitempty"`
	Labels *ReportChartLabels    `cty:"labels" hcl:"labels,block" json:"labels,omitempty"`
	Min    *int                  `cty:"min" hcl:"min" json:"min,omitempty"`
	Max    *int                  `cty:"max" hcl:"max" json:"max,omitempty"`
}

func (y *ReportChartAxesY) Equals(other *ReportChartAxesY) bool {
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

func (y *ReportChartAxesY) Merge(other *ReportChartAxesY) {
	if y.Title == nil {
		y.Title = other.Title
	} else {
		y.Title.Merge(other.Title)
	}
	if y.Labels == nil {
		y.Labels = other.Labels
	} else {
		y.Labels.Merge(other.Labels)
	}
	if y.Min == nil {
		y.Min = other.Min
	}
	if y.Max == nil {
		y.Max = other.Max
	}
}

type ReportChartAxisTitle struct {
	Display *string `cty:"display" hcl:"display" json:"display,omitempty"`
	Align   *string `cty:"align" hcl:"align" json:"align,omitempty"`
	Value   *string `cty:"value" hcl:"value" json:"value,omitempty"`
}

func (t *ReportChartAxisTitle) Equals(other *ReportChartAxisTitle) bool {
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

func (t *ReportChartAxisTitle) Merge(other *ReportChartAxisTitle) {
	if t.Display == nil {
		t.Display = other.Display
	}
	if t.Align == nil {
		t.Align = other.Align
	}
	if t.Value == nil {
		t.Value = other.Value
	}
}
