package modconfig

type DashboardNodeList []*DashboardNode

func (c *DashboardNodeList) Merge(other DashboardNodeList) {
	if other == nil {
		return
	}
	var nodeMap = make(map[string]bool)
	for _, node := range *c {
		nodeMap[node.ShortName] = true
	}

	for _, otherNode := range other {
		if !nodeMap[otherNode.ShortName] {
			*c = append(*c, otherNode)
		}
	}
}
