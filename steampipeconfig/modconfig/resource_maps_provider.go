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
	case BlockTypeHierarchy:
		resource, found = resourceMaps.DashboardHierarchies[longName]
	case BlockTypeImage:
		resource, found = resourceMaps.DashboardImages[longName]
	case BlockTypeTable:
		resource, found = resourceMaps.DashboardTables[longName]
	case BlockTypeText:
		resource, found = resourceMaps.DashboardTexts[longName]
	}
	return resource, found
}
