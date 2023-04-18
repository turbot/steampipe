package modconfig

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"strings"
)

type SteampipeRequire struct {
	MinVersionString string `hcl:"min_version,optional"`
	Constraint       *semver.Constraints
	DeclRange        hcl.Range
}

func (r *SteampipeRequire) initialise() hcl.Diagnostics {
	if r.MinVersionString == "" {
		return nil
	}
	constraint, err := semver.NewConstraint(fmt.Sprintf(">=%s", strings.TrimPrefix(r.MinVersionString, "v")))
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid required steampipe version %s", r.MinVersionString),
				Subject:  &r.DeclRange,
			}}
	}

	r.Constraint = constraint
	return nil
}
