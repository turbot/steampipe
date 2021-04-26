package modconfig

type OpenGraph struct {
	// The opengraph description (og:description) of the mod, for use in social media applications
	Description string `hcl:"description"`
	// The opengraph display title (og:title) of the mod, for use in social media applications.
	Title string `hcl:"title"`
}
