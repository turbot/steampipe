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

func (l *DashboardNodeList) Get(name string) *DashboardNode {
	for _, n := range *l {
		if n.Name() == name {
			return n
		}
	}
	return nil
}
