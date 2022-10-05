package modconfig

type DashboardEdgeList []*DashboardEdge

func (l *DashboardEdgeList) Merge(other DashboardEdgeList) {
	if other == nil {
		return
	}
	var edgeMap = make(map[string]bool)
	for _, edge := range *l {
		edgeMap[edge.ShortName] = true
	}

	for _, otherEdge := range other {
		if !edgeMap[otherEdge.ShortName] {
			*l = append(*l, otherEdge)
		}
	}
}

func (l *DashboardEdgeList) Names() []string {
	res := make([]string, len(*l))
	for i, e := range *l {
		res[i] = e.Name()
	}
	return res
}
