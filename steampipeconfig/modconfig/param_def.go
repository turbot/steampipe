package modconfig

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

type ParamDef struct {
	ShortName   string      `cty:"name" json:"name"`
	FullName    string      `cty:"name"`
	Description *string     `cty:"description" json:"description"`
	RawDefault  interface{} `json:"-"`
	Default     *string     `cty:"default" json:"default"`

	// list of all block referenced by the resource
	References []string `json:"refs"`
	// references stored as a map for easy checking
	referencesMap map[string]bool
	// list of resource names who reference this resource
	ReferencedBy []string `json:"referenced_by"`
}

func NewParamDef(block *hcl.Block) *ParamDef {
	return &ParamDef{
		ShortName: block.Labels[0],
		FullName:  fmt.Sprintf("query.%s", block.Labels[0]),
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
func (p *ParamDef) AddReference(reference string) {
	p.References = append(p.References, reference)
	p.referencesMap[reference] = true
}

// AddReferencedBy implements HclResource
func (p *ParamDef) AddReferencedBy(reference string) {
	p.ReferencedBy = append(p.ReferencedBy, reference)
}

// ReferencesResource implements HclResource
func (p *ParamDef) ReferencesResource(name string) bool {
	return p.referencesMap[name]
}

// SetMod implements HclResource
func (p *ParamDef) SetMod(mod *Mod) {}
