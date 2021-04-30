package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type OpenGraph struct {
	// The opengraph description (og:description) of the mod, for use in social media applications
	Description string `cty:"description"`
	// The opengraph display title (og:title) of the mod, for use in social media applications.
	Title     string `cty:"title"`
	DeclRange hcl.Range
}

// Schema :: hcl schema for control
func (o *OpenGraph) Schema() *hcl.BodySchema {
	return buildAttributeSchema(o)
}

func (o *OpenGraph) CtyValue() (cty.Value, error) {
	return getCtyValue(o)
}

// Name :: implementation of  HclResource
func (o *OpenGraph) Name() string {
	return o.Title
}
