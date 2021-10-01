package modconfig

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

type ParamDef struct {
	ShortName   string      `cty:"short_name" json:"name"`
	FullName    string      `cty:"full_name" json:"-"`
	Description *string     `cty:"description" json:"description"`
	RawDefault  interface{} `json:"-"`
	Default     *string     `cty:"default" json:"default"`

	// list of all block referenced by the resource
	References []ResourceReference `json:"refs"`
	// references stored as a map for easy checking
	referencesMap ResourceReferenceMap
	// list of resource names who reference this resource
	ReferencedBy []ResourceReference `json:"referenced_by"`

	parent string
}

func NewParamDef(block *hcl.Block, parent string) *ParamDef {
	return &ParamDef{
		ShortName:     block.Labels[0],
		FullName:      fmt.Sprintf("param.%s", block.Labels[0]),
		referencesMap: make(ResourceReferenceMap),
		parent:        parent,
	}
}

func (p ParamDef) String() string {
	return fmt.Sprintf("Name: %s, Description: %s, Default: %s", p.ShortName, typehelpers.SafeString(p.Description), typehelpers.SafeString(p.Default))
}

func (p ParamDef) Equals(other *ParamDef) bool {
	return p.ShortName == other.ShortName &&
		typehelpers.SafeString(p.Description) == typehelpers.SafeString(other.Description) &&
		typehelpers.SafeString(p.Default) == typehelpers.SafeString(other.Default)
}

// CtyValue implements HclResource
func (p *ParamDef) CtyValue() (cty.Value, error) {
	return getCtyValue(p)
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
	p.referencesMap.Add(ref)
}

// AddReferencedBy implements HclResource
//func (p *ParamDef)AddReferencedBy(ref []ResourceReference) {
//	p.ReferencedBy = append(p.ReferencedBy, ref...)
//}

// GetResourceReferences implements HclResource
//func (p *ParamDef) GetResourceReferences(resource HclResource) []ResourceReference {
//	return p.referencesMap[resource.Name()]
//}

// SetMod implements HclResource
//func (p *ParamDef) SetMod(mod *Mod) {}
