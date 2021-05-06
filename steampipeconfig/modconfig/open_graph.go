package modconfig

import (
	"github.com/hashicorp/hcl/v2"
)

type OpenGraph struct {
	// The opengraph description (og:description) of the mod, for use in social media applications
	Description string `cty:"description" hcl:"description"`
	// The opengraph display title (og:title) of the mod, for use in social media applications.
	Title     string `cty:"title" hcl:"title"`
	DeclRange hcl.Range
}
