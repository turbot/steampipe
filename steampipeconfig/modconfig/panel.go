package modconfig

// Panel is a struct representing the Report resource
type Panel struct {
	Title   *string   `hcl:"title"`
	Width   *int      `hcl:"width"`
	Source  *string   `hcl:"source"`
	SQL     *string   `hcl:"source"`
	Reports *[]Report `hcl:"panel,block"`
	Panels  *[]Panel  `hcl:"panel,block"`
}
