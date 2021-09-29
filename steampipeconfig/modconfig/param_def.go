package modconfig

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

type ParamDef struct {
	ShortName   string      `cty:"name" json:"name"`
	FullName    string      `cty:"name" json:"-"`
	Description *string     `cty:"description" json:"description"`
	RawDefault  interface{} `json:"-"`
	Default     *string     `cty:"default" json:"default"`

	// list of all block referenced by the resource
	References []ResourceReference `json:"refs"`
	// references stored as a map for easy checking
	referencesMap map[ResourceReference]bool
	// list of resource names who reference this resource
	ReferencedBy []ResourceReference `json:"referenced_by"`

	DeclRange hcl.Range

	parent string
}

func NewParamDef(block *hcl.Block, parent string) *ParamDef {
	return &ParamDef{
		ShortName:     block.Labels[0],
		FullName:      fmt.Sprintf("param.%s", block.Labels[0]),
		referencesMap: make(map[ResourceReference]bool),
		parent:        parent,
		DeclRange:     block.DefRange,
	}
}

func (p ParamDef) String() string {
	return fmt.Sprintf("Name: %s, Description: %s, Default: %s", p.Name, typehelpers.SafeString(p.Description), typehelpers.SafeString(p.Default))
}

func (p ParamDef) Equals(other *ParamDef) bool {
	return p.ShortName == other.ShortName &&
		typehelpers.SafeString(p.Description) == typehelpers.SafeString(other.Description) &&
		typehelpers.SafeString(p.Default) == typehelpers.SafeString(other.Default)
}

// Name implements HclResource
func (p *ParamDef) Name() string {
	return p.FullName
}

// OnDecoded implements HclResource
func (p *ParamDef) OnDecoded(*hcl.Block) hcl.Diagnostics { return nil }

// AddReference implements HclResource
func (p *ParamDef) AddReference(ref ResourceReference) {
	p.References = append(p.References, ref)
	p.referencesMap[ref] = true
}

// AddReferencedBy implements HclResource
func (p *ParamDef) AddReferencedBy(ref ResourceReference) {
	p.ReferencedBy = append(p.ReferencedBy, ref)
}

// ReferencesResource implements HclResource
func (p *ParamDef) ReferencesResource(ref ResourceReference) bool {
	return p.referencesMap[ref]
}

// SetMod implements HclResource
func (p *ParamDef) SetMod(mod *Mod) {}

// GetMod implements HclResource
func (p *ParamDef) GetMod() *Mod { return nil }

// GetDeclRange implements HclResource
func (p *ParamDef) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

// CtyValue implements HclResource
func (p *ParamDef) CtyValue() (cty.Value, error) {
	return getCtyValue(p)
}
