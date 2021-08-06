package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/steampipeconfig/modconfig/var_config"
	"github.com/zclconf/go-cty/cty"
)

// Variable is a struct representing a Variable resource
type Variable struct {
	ShortName string
	FullName  string

	Description string
	Default     cty.Value
	Type        cty.Type
	//Sensitive      bool
	DescriptionSet bool
	//SensitiveSet   bool

	ParsingMode var_config.VariableParsingMode
	DeclRange   hcl.Range

	metadata *ResourceMetadata
}

func NewVariable(v *var_config.Variable) *Variable {
	return &Variable{
		ShortName:   v.Name,
		Description: v.Description,
		FullName:    fmt.Sprintf("var.%s", v.Name),
		Default:     v.Default,
		Type:        v.Type,
		//Sensitive:    v.Sensitive,
		//SensitiveSet: v.SensitiveSet,
		ParsingMode: v.ParsingMode,

		DeclRange: v.DeclRange,
	}
}

// Name implements HclResource, ResourceWithMetadata
func (v *Variable) Name() string {
	return v.FullName
}

// GetMetadata implements ResourceWithMetadata
func (v *Variable) GetMetadata() *ResourceMetadata {
	return v.metadata
}

// SetMetadata implements ResourceWithMetadata
func (v *Variable) SetMetadata(metadata *ResourceMetadata) {
	v.metadata = metadata
}

// OnDecoded implements HclResource
func (v *Variable) OnDecoded(*hcl.Block) hcl.Diagnostics { return nil }

// AddReference implements HclResource
func (v *Variable) AddReference(string) {}

// CtyValue implements HclResource
func (v *Variable) CtyValue() (cty.Value, error) {
	return v.Default, nil
}

// Required returns true if this variable is required to be set by the caller,
// or false if there is a default value that will be used when it isn't set.
func (v *Variable) Required() bool {
	return v.Default == cty.NilVal
}
