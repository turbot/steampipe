package modconfig

type DashboardCategoryPropertyList []*DashboardCategoryProperty

func (c *DashboardCategoryPropertyList) Merge(other DashboardCategoryPropertyList) {
	if other == nil {
		return
	}
	var propertyMap = make(map[string]bool)
	for _, property := range *c {
		propertyMap[property.ShortName] = true
	}

	for _, otherProperty := range other {
		if !propertyMap[otherProperty.ShortName] {
			*c = append(*c, otherProperty)
		}
	}
}
