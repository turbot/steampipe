package modconfig

//
//func GetResource(provider ResourceProvider, parsedName *ParsedResourceName, modShortName string) (resource HclResource, found bool) {
//	longName := fmt.Sprintf("%s.%s.%s", modShortName, parsedName.ItemType, parsedName.Name)
//
//	switch parsedName.ItemType {
//	case BlockTypeBenchmark:
//		resource, found = provider.GetBenchmarks()[longName]
//	case BlockTypeControl:
//		resource, found = provider.GetControls()[longName]
//		if !found {
//			resource, found = provider.GetReportControls()[longName]
//		}
//	case BlockTypeReport:
//		resource, found = provider.GetReports[longName]
//	case BlockTypeContainer:
//		resource, found = provider.GetReportContainers[longName]
//	case BlockTypeChart:
//		resource, found = provider.GetReportCharts[longName]
//	case BlockTypeCounter:
//		resource, found = provider.GetReportCounters[longName]
//	case BlockTypeHierarchy:
//		resource, found = provider.GetReportHierarchies[longName]
//	case BlockTypeImage:
//		resource, found = provider.GetReportImages[longName]
//	case BlockTypeTable:
//		resource, found = provider.GetReportTables[longName]
//	case BlockTypeText:
//		resource, found = provider.GetReportTexts[longName]
//	}
//	return resource, found
//
//}
