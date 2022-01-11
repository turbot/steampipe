package modconfig

// NamedItem is a struct used by nebchmark, container and report to specify children of different types
type NamedItem struct {
	Name string `cty:"name"`
}

func (c NamedItem) String() string {
	return c.Name
}
