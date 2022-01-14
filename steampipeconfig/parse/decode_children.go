package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func decodeChildren(childNames []modconfig.NamedItem, block *hcl.Block, supportedChildren []string, runCtx *RunContext) ([]modconfig.ModTreeItem, hcl.Diagnostics) {
	if len(childNames) == 0 {
		return nil, nil
	}
	//
	var diags hcl.Diagnostics
	diags = checkForDuplicateChildren(childNames, block)
	if diags.HasErrors() {
		return nil, diags
	}

	// find the children in the eval context and populate control children
	children := make([]modconfig.ModTreeItem, len(childNames))

	for i, childName := range childNames {
		parsedName, err := modconfig.ParseResourceName(childName.Name)
		if err != nil || !helpers.StringSliceContains(supportedChildren, parsedName.ItemType) {
			diags = append(diags, childErrorDiagnostic(childName, block))
			continue
		}

		// now get the resource from the parent mod
		var mod *modconfig.Mod
		if parsedName.Mod == runCtx.CurrentMod.ShortName {
			mod = runCtx.CurrentMod
		} else {
			// we need to iterate through dependency mods - we cannot use parsedName.Mod as key as it is short name
			for _, dep := range runCtx.LoadedDependencyMods {
				if dep.ShortName == parsedName.Mod {
					mod = dep
					break
				}
			}
			if mod == nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Could not resolve mod for child %s", childName),
					Subject:  &block.TypeRange,
				})
				break
			}
		}

		child, found := mod.GetChildResource(parsedName)
		if !found {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Could not resolve child %s", childName),
				Subject:  &block.TypeRange,
			})
			continue
		}

		children[i] = child
	}
	if diags.HasErrors() {
		return nil, diags
	}

	return children, nil
}

func checkForDuplicateChildren(names []modconfig.NamedItem, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	// validate each child name appears only once
	nameMap := make(map[string]int)
	for _, n := range names {
		nameCount := nameMap[n.Name]
		// raise an error if this name appears more than once (but only raise 1 error per name)
		if nameCount == 1 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("'%s.%s' has duplicate child name '%s'", block.Type, block.Labels[0], n.Name),
				Subject:  &block.DefRange})
		}
		nameMap[n.Name] = nameCount + 1
	}

	return diags
}

func childErrorDiagnostic(childName modconfig.NamedItem, block *hcl.Block) *hcl.Diagnostic {
	return &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("Invalid child %s", childName),
		Subject:  &block.TypeRange,
	}
}

func getChildNameString(children []modconfig.ModTreeItem) []string {
	res := make([]string, len(children))
	for i, n := range children {
		res[i] = n.Name()
	}
	return res
}
