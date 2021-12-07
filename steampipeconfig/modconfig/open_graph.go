package modconfig

import (
	"github.com/hashicorp/hcl/v2"
)

// OpenGraph is a struct representing the OpenGraph group mod resource
type OpenGraph struct {
	// The opengraph description (og:description) of the mod, for use in social media applications
	Description *string `cty:"description" hcl:"description" json:"description"`
	// The opengraph display title (og:title) of the mod, for use in social media applications.
	Title     *string   `cty:"title" hcl:"title" json:"title"`
	Image     *string   `cty:"image" hcl:"image" json:"image"`
	DeclRange hcl.Range `json:"-"`
}
