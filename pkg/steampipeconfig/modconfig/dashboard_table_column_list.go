package modconfig

type DashboardTableColumnList []*DashboardTableColumn

func (c *DashboardTableColumnList) Merge(other DashboardTableColumnList) {
	if other == nil {
		return
	}
	var columnMap = make(map[string]bool)
	for _, column := range *c {
		columnMap[column.Name] = true
	}

	for _, otherColumn := range other {
		if !columnMap[otherColumn.Name] {
			*c = append(*c, otherColumn)
		}
	}
}
