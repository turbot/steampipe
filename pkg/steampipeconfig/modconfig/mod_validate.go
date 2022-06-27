package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/utils"
)

// ensure we have resolved all children in the resource tree
func (m *Mod) validateResourceTree() error {
	var errors []error
	for _, child := range m.GetChildren() {
		if err := m.validateChildren(child); err != nil {
			errors = append(errors, err)
		}
	}
	return utils.CombineErrorsWithPrefix(fmt.Sprintf("failed to resolve children for %d resources", len(errors)), errors...)
}

func (m *Mod) validateChildren(item ModTreeItem) error {
	missing := 0
	for _, child := range item.GetChildren() {
		if child == nil {
			missing++

		}
	}
	if missing > 0 {
		return fmt.Errorf("%s has %d unresolved children", item.Name(), missing)
	}
	return nil
}
