package modconfig

type DashboardGraphCategoryList []*DashboardGraphCategory

func (c DashboardGraphCategoryList) Merge(other DashboardGraphCategoryList) {
	if other == nil {
		return
	}
	var categoryMap = make(map[string]bool)
	for _, category := range c {
		categoryMap[category.Name] = true
	}

	for _, otherCategory := range other {
		if !categoryMap[otherCategory.Name] {
			c = append(c, otherCategory)
		}
	}
}
