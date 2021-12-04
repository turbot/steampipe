package version

import (
	"github.com/Masterminds/semver"
)

// Constraint wraps semver.Constraints type, adding the Original property
type Constraint struct {
	constraint *semver.Constraints
	Original   string
}

func NewConstraint(c string) (*Constraint, error) {
	constraints, err := semver.NewConstraint(c)
	if err != nil {
		return nil, err
	}
	return &Constraint{
		constraint: constraints,
		Original:   c,
	}, nil
}

// Check tests if a version satisfies the constraints.
func (c Constraint) Check(v *semver.Version) bool {
	return c.constraint.Check(v)
}

// Validate checks if a version satisfies a constraint. If not a slice of
// reasons for the failure are returned in addition to a bool.
func (c Constraint) Validate(v *semver.Version) (bool, []error) {
	return c.constraint.Validate(v)
}
