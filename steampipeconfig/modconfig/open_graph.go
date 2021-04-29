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
	var attributes []hcl.AttributeSchema
	for attribute := range GetAttributeDetails(o) {
		attributes = append(attributes, hcl.AttributeSchema{Name: attribute})
	}
	return &hcl.BodySchema{Attributes: attributes}
}

func (o *OpenGraph) CtyValue() (cty.Value, error) {
	return getCtyValue(o)
}

// FullName :: implementation of  HclResource
func (o *OpenGraph) FullName() string {
	return o.Title
}
