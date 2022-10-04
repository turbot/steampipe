package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
)

// enrich categories by fetching from ModResourcesProvider and add to the EdgeAndNodeProvider
func addEnrichedCategories(categories []*DashboardCategory, resource EdgeAndNodeProvider, provider ModResourcesProvider) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, c := range categories {
		fullCategory, moreDiags := enrichCategory(c, resource, provider)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			resource.AddCategory(fullCategory)
		}
	}
	return diags
}

// enrich the shell category by fetching from the ModResourcesProvider
// this is used when a category has been retrieved via a HCL reference - as cty does not serialise all properties
func enrichCategory(shellCategory *DashboardCategory, parent HclResource, resourceMapProvider ModResourcesProvider) (*DashboardCategory, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	resourceMaps := resourceMapProvider.GetResourceMaps()
	fullCategory, ok := resourceMaps.DashboardCategories[shellCategory.Name()]
	if !ok {
		diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s contains edge %s but this has not been loaded", parent.Name(), shellCategory.Name()),
			Subject:  parent.GetDeclRange(),
		})
		return nil, diags
	}
	return fullCategory, diags
}
