package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/type_conversion"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig/var_config"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/terraform-components/tfdiags"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// TODO check DescriptionSet - still required?

// Variable is a struct representing a Variable resource
type Variable struct {
	ResourceWithMetadataImpl
	ModTreeItemImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Default cty.Value `column:"default_value,jsonb" json:"-"`
	Type    cty.Type  `column:"var_type,text" json:"-"`

	TypeString string `json:"type"`
	DefaultGo  any    `json:"value_default"`
	ValueGo    any    `json:"value"`
	ModName    string `json:"mod_name"`

	// set after value resolution `column:"value,jsonb"`
	Value                      cty.Value                      `column:"value,jsonb" json:"-"`
	ValueSourceType            string                         `column:"value_source,text" json:"-"`
	ValueSourceFileName        string                         `column:"value_source_file_name,text" json:"-"`
	ValueSourceStartLineNumber int                            `column:"value_source_start_line_number,integer" json:"-"`
	ValueSourceEndLineNumber   int                            `column:"value_source_end_line_number,integer" json:"-"`
	ParsingMode                var_config.VariableParsingMode `json:"-"`

	metadata *ResourceMetadata
}

func NewVariable(v *var_config.Variable, mod *Mod) *Variable {
	var defaultGo interface{} = nil
	if !v.Default.IsNull() {
		defaultGo, _ = type_conversion.CtyToGo(v.Default)
	}
	fullName := fmt.Sprintf("%s.var.%s", mod.ShortName, v.Name)
	res := &Variable{
		ModTreeItemImpl: ModTreeItemImpl{
			HclResourceImpl: HclResourceImpl{
				ShortName:       v.Name,
				Description:     &v.Description,
				FullName:        fullName,
				DeclRange:       v.DeclRange,
				UnqualifiedName: fmt.Sprintf("var.%s", v.Name),
				blockType:       BlockTypeVariable,
			},
			Mod: mod,
		},
		Default:   v.Default,
		DefaultGo: defaultGo,
		// initialise the value to the default - may be set later
		Value:   v.Default,
		ValueGo: defaultGo,

		Type:        v.Type,
		ParsingMode: v.ParsingMode,
		ModName:     mod.ShortName,
		TypeString:  type_conversion.CtyTypeToHclType(v.Type, v.Default.Type()),
	}
	// if no type is set and a default _is_ set, use default to set the type
	if res.Type.Equals(cty.DynamicPseudoType) && !res.Default.IsNull() {
		res.Type = res.Default.Type()
	}
	return res
}

func (v *Variable) Equals(other *Variable) bool {
	return v.ShortName == other.ShortName &&
		v.FullName == other.FullName &&
		v.Description == other.Description &&
		v.Default.RawEquals(other.Default) &&
		v.Value.RawEquals(other.Value)
}

// OnDecoded implements HclResource
func (v *Variable) OnDecoded(block *hcl.Block, _ ResourceMapsProvider) hcl.Diagnostics {
	return nil
}

// Required returns true if this variable is required to be set by the caller,
// or false if there is a default value that will be used when it isn't set.
func (v *Variable) Required() bool {
	return v.Default == cty.NilVal
}

func (v *Variable) SetInputValue(value cty.Value, sourceType string, sourceRange tfdiags.SourceRange) error {
	// if the value type is a tuple with no elem type, and we have a type, set the variable to have our type
	if value.Type().Equals(cty.Tuple(nil)) && !v.Type.Equals(cty.DynamicPseudoType) {
		var err error
		value, err = convert.Convert(value, v.Type)
		if err != nil {
			return err
		}
	}

	v.Value = value
	v.ValueSourceType = sourceType
	v.ValueSourceFileName = sourceRange.Filename
	v.ValueSourceStartLineNumber = sourceRange.Start.Line
	v.ValueSourceEndLineNumber = sourceRange.End.Line
	v.ValueGo, _ = type_conversion.CtyToGo(value)
	// if type string is not set, derive from the type of value
	if v.TypeString == "" {
		v.TypeString = type_conversion.CtyTypeToHclType(value.Type())
	}

	return nil
}

func (v *Variable) Diff(other *Variable) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: v,
		Name: v.Name(),
	}

	if !utils.SafeStringsEqual(v.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(v.Value, other.Value) {
		res.AddPropertyDiff("Value")
	}

	res.populateChildDiffs(v, other)
	return res
}

// CtyValue implements CtyValueProvider
func (v *Variable) CtyValue() (cty.Value, error) {
	return GetCtyValue(v)
}
