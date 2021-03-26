package steampipeconfig

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

const modExtension = ".sp"

func parseMod(block *hcl.Block) (*modconfig.Mod, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	content, diags := block.Body.Content(modSchema)
	if diags.HasErrors() {
		return nil, diags
	}
	mod := &modconfig.Mod{
		Name: block.Labels[0],
	}
	moreDiags := parseModAttributes(content, mod)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
	}

	for _, block := range content.Blocks {
		switch block.Type {
		case "mod_depends":
			modDependency, moreDiags := parseModVersion(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			mod.ModDepends = append(mod.ModDepends, modDependency)
		case "plugin_depends":
			pluginDependency, moreDiags := parsePluginDependency(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			mod.PluginDepends = append(mod.PluginDepends, pluginDependency)
		}
	}

	return mod, diags
}

func parseModAttributes(content *hcl.BodyContent, mod *modconfig.Mod) hcl.Diagnostics {

	var diags hcl.Diagnostics
	if content.Attributes["title"] != nil {
		moreDiags := gohcl.DecodeExpression(content.Attributes["title"].Expr, nil, &mod.Title)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		}
	}
	if content.Attributes["description"] != nil {
		moreDiags := gohcl.DecodeExpression(content.Attributes["description"].Expr, nil, &mod.Description)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		}
	}
	if content.Attributes["version"] != nil {
		moreDiags := gohcl.DecodeExpression(content.Attributes["version"].Expr, nil, &mod.Version)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		}
	}

	return diags
}

func parseModVersion(block *hcl.Block) (*modconfig.ModVersion, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var dest = &modconfig.ModVersion{}

	diags = gohcl.DecodeBody(block.Body, nil, dest)
	if diags.HasErrors() {
		return nil, diags
	}

	return dest, nil
}

func parsePluginDependency(block *hcl.Block) (*modconfig.PluginDependency, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var dest = &modconfig.PluginDependency{}

	diags = gohcl.DecodeBody(block.Body, nil, dest)
	if diags.HasErrors() {
		return nil, diags
	}

	return dest, nil
}

func parseQuery(block *hcl.Block) (*modconfig.Query, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var dest = &modconfig.Query{}

	diags = gohcl.DecodeBody(block.Body, nil, dest)
	if diags.HasErrors() {
		return nil, diags
	}

	dest.Name = block.Labels[0]
	return dest, nil
}
