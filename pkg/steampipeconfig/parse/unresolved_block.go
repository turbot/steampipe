package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type unresolvedBlock struct {
	Name         string
	Block        *hcl.Block
	DeclRange    hcl.Range
	Dependencies map[string]*modconfig.ResourceDependency
}

func newUnresolvedBlock(block *hcl.Block, name string, dependencies map[string]*modconfig.ResourceDependency) *unresolvedBlock {
	return &unresolvedBlock{
		Name:         name,
		Block:        block,
		Dependencies: dependencies,
		DeclRange:    hclhelpers.BlockRange(block),
	}
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
