package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Local struct {
	ShortName string
	FullName  string `cty:"name"`

	Value     cty.Value
	DeclRange hcl.Range

	metadata *ResourceMetadata
}

func NewLocal(name string, val cty.Value, declRange hcl.Range) *Local {
	return &Local{
		ShortName: name,
		FullName:  fmt.Sprintf("local.%s", name),
		Value:     val,
		DeclRange: declRange,
	}
}

// Name :: implementation of HclResource
func (l *Local) Name() string {
	return l.FullName
}

// GetMetadata :: implementation of HclResource
func (l *Local) GetMetadata() *ResourceMetadata {
	return l.metadata
}

// OnDecoded :: implementation of HclResource
func (l *Local) OnDecoded() {}

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
