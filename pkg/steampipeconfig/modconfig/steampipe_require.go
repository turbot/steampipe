package modconfig

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
)

type SteampipeRequire struct {
	MinVersionString string `hcl:"min_version,optional"`
	Constraint       *semver.Constraints
	DeclRange        hcl.Range
}

func (r *SteampipeRequire) initialise(requireBlock *hcl.Block) hcl.Diagnostics {
	// find the steampipe block
	steampipeBlock := hclhelpers.FindFirstChildBlock(requireBlock, BlockTypeSteampipe)
	if steampipeBlock == nil {
		// can happen if there is a legacy property - just use th eparent block
		steampipeBlock = requireBlock
	}
	// set DeclRange
	r.DeclRange = steampipeBlock.DefRange

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
