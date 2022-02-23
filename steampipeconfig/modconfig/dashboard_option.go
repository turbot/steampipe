package modconfig

// DashboardInputOption is a struct representing dashboard input option
type DashboardInputOption struct {
	Name  string  `hcl:"name,label" json:"name"`
	Label *string `cty:"label" hcl:"label" json:"label,omitempty"`
}
