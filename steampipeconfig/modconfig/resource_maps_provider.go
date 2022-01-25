package modconfig

import "fmt"

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
	case BlockTypeReport:
		resource, found = resourceMaps.Reports[longName]
	case BlockTypeContainer:
		resource, found = resourceMaps.ReportContainers[longName]
	case BlockTypeChart:
		resource, found = resourceMaps.ReportCharts[longName]
	case BlockTypeCounter:
		resource, found = resourceMaps.ReportCounters[longName]
	case BlockTypeHierarchy:
		resource, found = resourceMaps.ReportHierarchies[longName]
	case BlockTypeImage:
		resource, found = resourceMaps.ReportImages[longName]
	case BlockTypeInput:
		resource, found = resourceMaps.ReportInputs[longName]
	case BlockTypeTable:
		resource, found = resourceMaps.ReportTables[longName]
	case BlockTypeText:
		resource, found = resourceMaps.ReportTexts[longName]
	}
	return resource, found
}
