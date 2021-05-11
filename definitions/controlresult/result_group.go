package controlresult

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ResultGroup is a struct representing a grouping of control results
//
// It may correspond to a Benchmark, or some other arbitrary grouping
type ResultGroup struct {
	GroupId     string
	Title       string
	Description string
	Tags        map[string]string

	Summary struct {
		Status struct {
		}
	}
	Groups  []*ResultGroup
	Results []*Result
	// Controls are derived from results

	parent     *ResultGroup
	ChildItems []modconfig.ControlTreeItem
}

// NewResultGroup creates a result group from a ControlTreeItem
func NewResultGroup(item modconfig.ControlTreeItem) *ResultGroup {
	res := &ResultGroup{
		GroupId:     item.Name(),
		Title:       item.GetTitle(),
		Description: item.GetDescription(),
		Tags:        item.GetTags(),
	}
	// add child groups for children which are benchmarks
	for _, c := range item.GetChildren() {
		if benchmark, ok := c.(*modconfig.Benchmark); ok {
			res.Groups = append(res.Groups, NewResultGroup(benchmark))
		}
	}
	return res
}

// GetGroupMap mutates the passed in a map to return all child result groups
func (r *ResultGroup) GetGroupMap(groupMap map[string]*ResultGroup) {
	if groupMap == nil {
		groupMap = make(map[string]*ResultGroup)
	}
	// add self
	groupMap[r.GroupId] = r
	for _, g := range r.Groups {
		g.GetGroupMap(groupMap)
	}
}
