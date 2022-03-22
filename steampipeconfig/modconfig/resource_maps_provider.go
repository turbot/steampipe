package modconfig

import "fmt"

// GetResource tries to find a resource with the given name in the ModResourcesProvider
// NOTE: this does NOT support inputs, which are NOT uniquely named in a mod
func GetResource(provider ModResourcesProvider, parsedName *ParsedResourceName) (resource HclResource, found bool) {
	resourceMaps := provider.GetResourceMaps()
	modName := parsedName.Mod
	if modName == "" {
		modName = resourceMaps.Mod.ShortName
	}
	longName := fmt.Sprintf("%s.%s.%s", modName, parsedName.ItemType, parsedName.Name)

	switch parsedName.ItemType {
	case BlockTypeBenchmark:
		resource, found = resourceMaps.Benchmarks[longName]
	case BlockTypeControl:
		resource, found = resourceMaps.Controls[longName]
	case BlockTypeDashboard:
		resource, found = resourceMaps.Dashboards[longName]
	case BlockTypeContainer:
		resource, found = resourceMaps.DashboardContainers[longName]
	case BlockTypeCard:
		resource, found = resourceMaps.DashboardCards[longName]
	case BlockTypeChart:
		resource, found = resourceMaps.DashboardCharts[longName]
	case BlockTypeFlow:
		resource, found = resourceMaps.DashboardFlows[longName]
	case BlockTypeHierarchy:
		resource, found = resourceMaps.DashboardHierarchies[longName]
	case BlockTypeImage:
		resource, found = resourceMaps.DashboardImages[longName]
	case BlockTypeTable:
		resource, found = resourceMaps.DashboardTables[longName]
	case BlockTypeText:
		resource, found = resourceMaps.DashboardTexts[longName]
	case BlockTypeInput:
		// this function only supports global inputs
		// if the input has a parent dashboard, you must use GetDashboardInput
		resource, found = resourceMaps.GlobalDashboardInputs[longName]
	}
	return resource, found
}

// GetDashboardInput looks for an input with a given parent dashboard
// this is required as GetResource does not support Inputs
func GetDashboardInput(provider ModResourcesProvider, inputName, dashboardName string) (*DashboardInput, bool) {
	resourceMaps := provider.GetResourceMaps()

	dasboardInputs, ok := resourceMaps.DashboardInputs[dashboardName]
	if !ok {
		return nil, false
	}

	input, ok := dasboardInputs[inputName]

	return input, ok
}
