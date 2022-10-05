package modconfig

type DashboardNodeList []*DashboardNode

func (l *DashboardNodeList) Merge(other DashboardNodeList) {
	if other == nil {
		return
	}
	var nodeMap = make(map[string]bool)
	for _, node := range *l {
		nodeMap[node.ShortName] = true
	}

	for _, otherNode := range other {
		if !nodeMap[otherNode.ShortName] {
			*l = append(*l, otherNode)
		}
	}
}

func (l *DashboardNodeList) Names() []string {
	res := make([]string, len(*l))
	for i, n := range *l {
		res[i] = n.Name()
	}
	return res
}
