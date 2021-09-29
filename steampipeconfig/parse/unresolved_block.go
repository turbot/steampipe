package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

type unresolvedBlock struct {
	Name         string
	Block        *hcl.Block
	Dependencies []*dependency
}

func (b unresolvedBlock) String() string {
	depStrings := make([]string, len(b.Dependencies))
	for i, dep := range b.Dependencies {
		depStrings[i] = fmt.Sprintf(`%s -> %s`, b.Name, dep.String())
	}
	return strings.Join(depStrings, "\n")
}
