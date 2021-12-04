package version

import (
	"strings"

	"github.com/Masterminds/semver"
)

// Constraints encapsulates a list of Constraint objects
// and provides transparent checking of multiple constraints
type Constraints []*Constraint

// Check tests if a version satisfies ALL the constraints.
func (c Constraints) Check(v *semver.Version) bool {
	for _, constraint := range c {
		if !constraint.Check(v) {
			return false
		}
	}
	return true
}

func (c *Constraints) Add(constraint *Constraint) {
	*c = append(*c, constraint)
}

func (c Constraints) String() string {
	var strs = make([]string, len(c))
	for i, constraint := range c {
		strs[i] = constraint.Original
	}
	return strings.Join(strs, ",")

}
