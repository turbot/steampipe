package modconfig

type DashboardEdgeList []*DashboardEdge

func (c *DashboardEdgeList) Merge(other DashboardEdgeList) {
	if other == nil {
		return
	}
	var edgeMap = make(map[string]bool)
	for _, edge := range *c {
		edgeMap[edge.ShortName] = true
	}

	for _, otherEdge := range other {
		if !edgeMap[otherEdge.ShortName] {
			*c = append(*c, otherEdge)
		}
	}
}
