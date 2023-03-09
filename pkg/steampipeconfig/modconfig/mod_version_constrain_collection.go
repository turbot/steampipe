package modconfig

// ModVersionConstraintCollection is a collection of ModVersionConstraint instances and implements the sort
// interface. See the sort package for more details.
// https://golang.org/pkg/sort/
type ModVersionConstraintCollection []*ModVersionConstraint

// Len returns the length of a collection. The number of Version instances
// on the slice.
func (c ModVersionConstraintCollection) Len() int {
	return len(c)
}

// Less is needed for the sort interface to compare two Version objects on the
// slice. If checks if one is less than the other.
func (c ModVersionConstraintCollection) Less(i, j int) bool {
	// sort by name
	return c[i].Name < (c[j].Name)
}

// Swap is needed for the sort interface to replace the Version objects
// at two different positions in the slice.
func (c ModVersionConstraintCollection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
