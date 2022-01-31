package modconfig

// NamedItem is a struct used by nebchmark, container and report to specify children of different types
type NamedItem struct {
	Name string `cty:"name"`
}

func (c NamedItem) String() string {
	return c.Name
}

type NamedItemList []NamedItem

func (l NamedItemList) StringList() []string {
	res := make([]string, len(l))
	for i, n := range l {
		res[i] = n.Name
	}
	return res
}
