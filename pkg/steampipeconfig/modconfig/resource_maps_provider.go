package modconfig

import "fmt"

// GetResource tries to find a resource with the given name in the ResourceMapsProvider
// NOTE: this does NOT support inputs, which are NOT uniquely named in a mod
func GetResource(provider ResourceMapsProvider, parsedName *ParsedResourceName) (resource HclResource, found bool) {
	resourceMaps := provider.GetResourceMaps()
	modName := parsedName.Mod
	if modName == "" {
		modName = resourceMaps.Mod.ShortName
	}
	longName := fmt.Sprintf("%s.%s.%s", modName, parsedName.ItemType, parsedName.Name)

	// NOTE: we could use WalkResources, but this is quicker

	switch parsedName.ItemType {
	case BlockTypeBenchmark:
		resource, found = resourceMaps.Benchmarks[longName]
	case BlockTypeControl:
		resource, found = resourceMaps.Controls[longName]
	case BlockTypeDashboard:
		resource, found = resourceMaps.Dashboards[longName]
	case BlockTypeCard:
		resource, found = resourceMaps.DashboardCards[longName]
	case BlockTypeCategory:
		resource, found = resourceMaps.DashboardCategories[longName]
	case BlockTypeChart:
		resource, found = resourceMaps.DashboardCharts[longName]
	case BlockTypeContainer:
		resource, found = resourceMaps.DashboardContainers[longName]
	case BlockTypeEdge:
		resource, found = resourceMaps.DashboardEdges[longName]
	case BlockTypeFlow:
		resource, found = resourceMaps.DashboardFlows[longName]
	case BlockTypeGraph:
		resource, found = resourceMaps.DashboardGraphs[longName]
	case BlockTypeHierarchy:
		resource, found = resourceMaps.DashboardHierarchies[longName]
	case BlockTypeImage:
		resource, found = resourceMaps.DashboardImages[longName]
	case BlockTypeNode:
		resource, found = resourceMaps.DashboardNodes[longName]
	case BlockTypeTable:
		resource, found = resourceMaps.DashboardTables[longName]
	case BlockTypeText:
		resource, found = resourceMaps.DashboardTexts[longName]
	case BlockTypeInput:
		// this function only supports global inputs
		// if the input has a parent dashboard, you must use GetDashboardInput
		resource, found = resourceMaps.GlobalDashboardInputs[longName]
	case BlockTypeQuery:
		resource, found = resourceMaps.Queries[longName]
	case BlockTypeVariable:
		resource, found = resourceMaps.Variables[longName]
	}
	return resource, found
}

// GetDashboardInput looks for an input with a given parent dashboard
// this is required as GetResource does not support Inputs
func GetDashboardInput(provider ResourceMapsProvider, inputName, dashboardName string) (*DashboardInput, bool) {
	resourceMaps := provider.GetResourceMaps()

	dasboardInputs, ok := resourceMaps.DashboardInputs[dashboardName]
	if !ok {
		return nil, false
	}

	input, ok := dasboardInputs[inputName]

	return input, ok
}
