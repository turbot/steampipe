package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func decodeInlineChildren(content *hcl.BodyContent, parent modconfig.ModTreeItem, runCtx *RunContext) ([]modconfig.ModTreeItem, *decodeResult) {
	var res = &decodeResult{}
	// if children are declared inline as blocks, add them
	var children []modconfig.ModTreeItem
	for _, b := range content.Blocks {
		resources, blockRes := decodeBlock(b, parent, runCtx)
		res.Merge(blockRes)
		if !blockRes.Success() {
			continue
		}
		for _, childResource := range resources {
			if child, ok := childResource.(modconfig.ModTreeItem); ok {
				children = append(children, child)
			}
		}
	}
	return children, res
}

func resolveChildrenFromNames(childNames []string, block *hcl.Block, supportedChildren []string, runCtx *RunContext) ([]modconfig.ModTreeItem, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	diags = checkForDuplicateChildren(childNames, block)
	if diags.HasErrors() {
		return nil, diags
	}

	// find the children in the eval context and populate control children
	children := make([]modconfig.ModTreeItem, len(childNames))

	for i, childName := range childNames {
		parsedName, err := modconfig.ParseResourceName(childName)
		if err != nil || !helpers.StringSliceContains(supportedChildren, parsedName.ItemType) {
			diags = append(diags, childErrorDiagnostic(childName, block))
			continue
		}

		// now get the resource from the parent mod
		var mod = runCtx.GetMod(parsedName.Mod)
		if mod == nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Could not resolve mod for child %s", childName),
				Subject:  &block.TypeRange,
			})
			break
		}

		resource, found := modconfig.GetResource(mod, parsedName)
		// ensure this item is a mod tree item
		child, ok := resource.(modconfig.ModTreeItem)
		if !found || !ok {
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

func checkForDuplicateChildren(names []string, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	// validate each child name appears only once
	nameMap := make(map[string]int)
	for _, n := range names {
		nameCount := nameMap[n]
		// raise an error if this name appears more than once (but only raise 1 error per name)
		if nameCount == 1 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("'%s.%s' has duplicate child name '%s'", block.Type, block.Labels[0], n),
				Subject:  &block.DefRange})
		}
		nameMap[n] = nameCount + 1
	}

	return diags
}

func childErrorDiagnostic(childName string, block *hcl.Block) *hcl.Diagnostic {
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
