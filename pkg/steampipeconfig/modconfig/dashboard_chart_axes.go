package modconfig

import (
	"github.com/turbot/steampipe/pkg/utils"
)

type DashboardChartAxes struct {
	X *DashboardChartAxesX `cty:"x" hcl:"x,block" json:"x,omitempty"`
	Y *DashboardChartAxesY `cty:"y" hcl:"y,block" json:"y,omitempty"`
}

func (a *DashboardChartAxes) Equals(other *DashboardChartAxes) bool {
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

func (a *DashboardChartAxes) Merge(other *DashboardChartAxes) {
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

type DashboardChartAxesX struct {
	Title  *DashboardChartAxisTitle `cty:"title" hcl:"title,block" json:"title,omitempty"`
	Labels *DashboardChartLabels    `cty:"labels" hcl:"labels,block" json:"labels,omitempty"`
	Min    *int                     `cty:"min" hcl:"min" json:"min,omitempty"`
	Max    *int                     `cty:"max" hcl:"max" json:"max,omitempty"`
}

func (x *DashboardChartAxesX) Equals(other *DashboardChartAxesX) bool {
	if other == nil {
		return false
	}

	if !(utils.SafeIntEqual(x.Min, other.Min) &&
		utils.SafeIntEqual(x.Max, other.Max)) {
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

func (x *DashboardChartAxesX) Merge(other *DashboardChartAxesX) {
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
	if x.Min == nil {
		x.Min = other.Min
	}
	if x.Max == nil {
		x.Max = other.Max
	}
}

type DashboardChartAxesY struct {
	Title  *DashboardChartAxisTitle `cty:"title" hcl:"title,block" json:"title,omitempty"`
	Labels *DashboardChartLabels    `cty:"labels" hcl:"labels,block" json:"labels,omitempty"`
	Min    *int                     `cty:"min" hcl:"min" json:"min,omitempty"`
	Max    *int                     `cty:"max" hcl:"max" json:"max,omitempty"`
}

func (y *DashboardChartAxesY) Equals(other *DashboardChartAxesY) bool {
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

func (y *DashboardChartAxesY) Merge(other *DashboardChartAxesY) {
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

type DashboardChartAxisTitle struct {
	Display *string `cty:"display" hcl:"display" json:"display,omitempty"`
	Align   *string `cty:"align" hcl:"align" json:"align,omitempty"`
	Value   *string `cty:"value" hcl:"value" json:"value,omitempty"`
}

func (t *DashboardChartAxisTitle) Equals(other *DashboardChartAxisTitle) bool {
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

func (t *DashboardChartAxisTitle) Merge(other *DashboardChartAxisTitle) {
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
