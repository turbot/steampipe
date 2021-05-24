package modconfig

// Report is a struct representing the Report resource
type Report struct {
	Reports *[]Report `hcl:"report, block"`
	Panels  *[]Panel  ` hcl:"panel, block"`
}
