package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Local struct {
	ShortName string
	Value     cty.Value
	DeclRange hcl.Range

	metadata *ResourceMetadata
}

// FullName :: implementation of HclResource
func (l *Local) FullName() string {
	return fmt.Sprintf("local.%s", l.ShortName)
}

// GetMetadata :: implementation of HclResource
func (l *Local) GetMetadata() *ResourceMetadata {
	return l.metadata
}

// SetMetadata :: implementation of HclResource
func (l *Local) SetMetadata(metadata *ResourceMetadata) {
	l.metadata = metadata
}

// CtyValue :: implementation of HclResource
func (l *Local) CtyValue() (cty.Value, error) {
	return l.Value, nil
}

// Schema :: implementation of HclResource
func (l *Local) Schema() *hcl.BodySchema {
	// no schema needed - we manual decode
	return nil
}
