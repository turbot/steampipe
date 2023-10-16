package modconfig

import (
	"fmt"
	"github.com/turbot/go-kit/hcl_helpers"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
)

type SteampipeRequire struct {
	MinVersionString string `hcl:"min_version,optional"`
	Constraint       *semver.Constraints
	DeclRange        hcl.Range
}

func (r *SteampipeRequire) initialise(requireBlock *hcl.Block) hcl.Diagnostics {
	// find the steampipe block
	steampipeBlock := hcl_helpers.FindFirstChildBlock(requireBlock, BlockTypeSteampipe)
	if steampipeBlock == nil {
		// can happen if there is a legacy property - just use the parent block
		steampipeBlock = requireBlock
	}
	// set DeclRange
	r.DeclRange = hcl_helpers.BlockRange(steampipeBlock)

	if r.MinVersionString == "" {
		return nil
	}

	// convert min version into constraint (including prereleases)
	minVersion, err := semver.NewVersion(strings.TrimPrefix(r.MinVersionString, "v"))
	if err == nil {
		r.Constraint, err = semver.NewConstraint(fmt.Sprintf(">=%s-0", minVersion))
	}
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid required steampipe version %s", r.MinVersionString),
				Subject:  &r.DeclRange,
			}}
	}
	return nil
}
