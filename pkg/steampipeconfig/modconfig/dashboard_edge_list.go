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

func (l *DashboardEdgeList) Get(name string) *DashboardEdge {
	for _, n := range *l {
		if n.Name() == name {
			return n
		}
	}
	return nil
}

func (l *DashboardEdgeList) Names() []string {
	res := make([]string, len(*l))
	for i, e := range *l {
		res[i] = e.Name()
	}
	return res
}

func (l *DashboardEdgeList) Contains(other *DashboardEdge) bool {
	for _, e := range *l {
		if e == other {
			return true
		}
	}
	return false
}
