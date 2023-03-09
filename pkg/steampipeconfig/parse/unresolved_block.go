package parse

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"

	"github.com/hashicorp/hcl/v2"
)

type unresolvedBlock struct {
	Name         string
	Block        *hcl.Block
	Dependencies map[string]*modconfig.ResourceDependency
}

func (b unresolvedBlock) String() string {
	depStrings := make([]string, len(b.Dependencies))
	idx := 0
	for _, dep := range b.Dependencies {
		depStrings[idx] = fmt.Sprintf(`%s -> %s`, b.Name, dep.String())
		idx++
	}
	return strings.Join(depStrings, "\n")
}
