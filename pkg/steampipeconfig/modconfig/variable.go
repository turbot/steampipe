package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/tfdiags"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig/var_config"
	"github.com/turbot/steampipe/pkg/type_conversion"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// Variable is a struct representing a Variable resource
type Variable struct {
	ResourceWithMetadataBase

	ShortName string `json:"name"`
	FullName  string `column:"name,text" json:"-"`

	Description    string    `column:"description,text" json:"description"`
	Default        cty.Value `column:"default_value,jsonb" json:"-"`
	Type           cty.Type  `column:"var_type,text" json:"-"`
	DescriptionSet bool      ` json:"-"`

	TypeString string      `json:"type"`
	DefaultGo  interface{} `json:"value_default"`
	ValueGo    interface{} `json:"value"`
	ModName    string      `json:"mod_name"`

	// set after value resolution `column:"value,jsonb"`
	Value                      cty.Value                      `column:"value,jsonb" json:"-"`
	ValueSourceType            string                         `column:"value_source,text" json:"-"`
	ValueSourceFileName        string                         `column:"value_source_file_name,text" json:"-"`
	ValueSourceStartLineNumber int                            `column:"value_source_start_line_number,integer" json:"-"`
	ValueSourceEndLineNumber   int                            `column:"value_source_end_line_number,integer" json:"-"`
	DeclRange                  hcl.Range                      `json:"-"`
	ParsingMode                var_config.VariableParsingMode `json:"-"`
	Mod                        *Mod                           `json:"-"`
	UnqualifiedName            string                         `json:"-"`
	Paths                      []NodePath                     `column:"path,jsonb" json:"-"`

	metadata *ResourceMetadata
	parents  []ModTreeItem
}

func NewVariable(v *var_config.Variable, mod *Mod) *Variable {
	var defaultGo interface{} = nil
	if !v.Default.IsNull() {
		defaultGo, _ = type_conversion.CtyToGo(v.Default)
	}

	return &Variable{
		ShortName:       v.Name,
		Description:     v.Description,
		FullName:        fmt.Sprintf("%s.var.%s", mod.ShortName, v.Name),
		UnqualifiedName: fmt.Sprintf("var.%s", v.Name),
		Default:         v.Default,
		Type:            v.Type,
		ParsingMode:     v.ParsingMode,
		Mod:             mod,
		DeclRange:       v.DeclRange,
		ModName:         mod.ShortName,
		DefaultGo:       defaultGo,
		TypeString:      type_conversion.CtyTypeToHclType(v.Type, v.Default.Type()),
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

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (v *Variable) GetUnqualifiedName() string {
	return v.UnqualifiedName
}

// OnDecoded implements HclResource
func (v *Variable) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	return nil
}

// AddReference implements ResourceWithMetadata
func (v *Variable) AddReference(*ResourceReference) {}

// GetReferences implements ResourceWithMetadata
func (v *Variable) GetReferences() []*ResourceReference {
	return nil
}

// GetMod implements ModTreeItem
func (v *Variable) GetMod() *Mod {
	return v.Mod
}

// CtyValue implements HclResource
func (v *Variable) CtyValue() (cty.Value, error) {
	return v.Default, nil
}

// GetDeclRange implements HclResource
func (v *Variable) GetDeclRange() *hcl.Range {
	return &v.DeclRange
}

// BlockType implements HclResource
func (*Variable) BlockType() string {
	return BlockTypeVariable
}

// Required returns true if this variable is required to be set by the caller,
// or false if there is a default value that will be used when it isn't set.
func (v *Variable) Required() bool {
	return v.Default == cty.NilVal
}

func (v *Variable) SetInputValue(value cty.Value, sourceType string, sourceRange tfdiags.SourceRange) error {
	// if no type is set and a default _is_ set, use default to set the type
	if v.Type.Equals(cty.DynamicPseudoType) && !v.Default.IsNull() {
		v.Type = v.Default.Type()
	}

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

// AddParent implements ModTreeItem
func (v *Variable) AddParent(parent ModTreeItem) error {
	v.parents = append(v.parents, parent)

	return nil
}

// GetParents implements ModTreeItem
func (v *Variable) GetParents() []ModTreeItem {
	return v.parents
}

// GetChildren implements ModTreeItem
func (v *Variable) GetChildren() []ModTreeItem {
	return nil
}

// GetDescription implements ModTreeItem
func (v *Variable) GetDescription() string {
	return ""
}

// GetTitle implements HclResource
func (v *Variable) GetTitle() string {
	return typehelpers.SafeString(v.ShortName)
}

// GetTags implements HclResource
func (v *Variable) GetTags() map[string]string {
	return map[string]string{}
}

// GetPaths implements ModTreeItem
func (v *Variable) GetPaths() []NodePath {
	// lazy load
	if len(v.Paths) == 0 {
		v.SetPaths()
	}
	return v.Paths
}

// SetPaths implements ModTreeItem
func (v *Variable) SetPaths() {
	for _, parent := range v.parents {
		for _, parentPath := range parent.GetPaths() {
			v.Paths = append(v.Paths, append(parentPath, v.Name()))
		}
	}
}

// GetDocumentation implement ModTreeItem
func (*Variable) GetDocumentation() string {
	return ""
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
