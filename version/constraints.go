package version

import (
	"github.com/Masterminds/semver"
)

// Constraints wraps semver.Constraints type, adding the Original property
type Constraints struct {
	constraint *semver.Constraints
	Original   string
}

func NewConstraint(c string) (*Constraints, error) {
	constraints, err := semver.NewConstraint(c)
	if err != nil {
		return nil, err
	}
	return &Constraints{
		constraint: constraints,
		Original:   c,
	}, nil
}

// Check tests if a version satisfies the constraints.
func (c Constraints) Check(v *semver.Version) bool {
	return c.constraint.Check(v)
}

// Validate checks if a version satisfies a constraint. If not a slice of
// reasons for the failure are returned in addition to a bool.
func (c Constraints) Validate(v *semver.Version) (bool, []error) {
	return c.constraint.Validate(v)
}
