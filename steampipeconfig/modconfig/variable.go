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
	FullName  string `column:"name,text"`

	Description    string `column:"description,text"`
	Default        cty.Value
	Type           cty.Type
	DescriptionSet bool

	// set after value resolution
	Value            cty.Value
	ValueSourceType  string    `column:"value_source,text"`
	ValueSourceRange hcl.Range `column:"value_source_range,text"`
	DeclRange        hcl.Range `column:"decl_range,text"`
	ParsingMode      var_config.VariableParsingMode
	Mod              *Mod

	metadata      *ResourceMetadata
	ValueString   string `column:"value,jsonb"`
	TypeString    string `column:"var_type,text"`
	DefaultString string `column:"default_value,jsonb"`
}

func NewVariable(v *var_config.Variable) *Variable {
	return &Variable{
		ShortName:   v.Name,
		Description: v.Description,
		FullName:    fmt.Sprintf("var.%s", v.Name),
		Default:     v.Default,
		Type:        v.Type,
		ParsingMode: v.ParsingMode,

		DeclRange: v.DeclRange,
	}
}

func (v *Variable) Equals(other *Variable) bool {
	return v.ShortName == other.ShortName &&
		v.FullName == other.FullName &&
		v.Description == other.Description &&
		v.Default.RawEquals(other.Default) &&
		v.Value.RawEquals(other.Value)
}

// Name implements HclResource, ResourceWithMetadata
func (v *Variable) Name() string {
	return v.FullName
}

// QualifiedName returns the name in format: '<modName>.var.<shortName>'
func (v *Variable) QualifiedName() string {
	return fmt.Sprintf("%s.%s", v.metadata.ModShortName, v.FullName)
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

// SetMod implements HclResource
func (v *Variable) SetMod(mod *Mod) {
	v.Mod = mod
}

// CtyValue implements HclResource
func (v *Variable) CtyValue() (cty.Value, error) {
	return v.Default, nil
}

// Required returns true if this variable is required to be set by the caller,
// or false if there is a default value that will be used when it isn't set.
func (v *Variable) Required() bool {
	return v.Default == cty.NilVal
}

func (v *Variable) SetInputValue(value cty.Value, valueString, defaultValueString, varTypeString, sourceType string, sourceRange hcl.Range) {
	v.Value = value
	v.ValueSourceType = sourceType
	v.ValueSourceRange = sourceRange

	// also generate string values for default and value and type
	v.ValueString = valueString
	v.TypeString = varTypeString
	v.DefaultString = defaultValueString
}
