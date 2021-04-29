package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/zclconf/go-cty/cty"
)

type OpenGraph struct {
	// The opengraph description (og:description) of the mod, for use in social media applications
	Description string `json:"description"`
	// The opengraph display title (og:title) of the mod, for use in social media applications.
	Title     string    `json:"title"`
	DeclRange hcl.Range `json:"-"`
}

// Schema :: hcl schema for control
func (o *OpenGraph) Schema() *hcl.BodySchema {
	var attributes []hcl.AttributeSchema
	for attribute := range HclProperties(o) {
		attributes = append(attributes, hcl.AttributeSchema{Name: attribute})
	}
	return &hcl.BodySchema{Attributes: attributes}
}

func (o *OpenGraph) CtyValue() (cty.Value, error) {
	return getCtyValue(o, openGraphBlock)
}

// openGraphBlock :: return the block schema of a hydrated OpenGraph
// used to convert a openGraph into a cty type for block evaluation
// TODO autogenerate from OpenGraph struct by reflection?
var openGraphBlock = configschema.Block{
	Attributes: map[string]*configschema.Attribute{
		"description": {Optional: true, Type: cty.String},
		"title":       {Optional: true, Type: cty.String},
	},
}

func openGraphCtyType() cty.Type {
	spec := openGraphBlock.DecoderSpec()
	return hcldec.ImpliedType(spec)
}
