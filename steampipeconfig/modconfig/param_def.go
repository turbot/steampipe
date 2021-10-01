package modconfig

import (
	"fmt"

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
	referencesMap ResourceReferenceMap
	// list of resource names who reference this resource
	ReferencedBy []ResourceReference `json:"referenced_by"`

	DeclRange hcl.Range

	parent string
}

func NewParamDef(block *hcl.Block, parent string) *ParamDef {
	return &ParamDef{
		ShortName:     block.Labels[0],
		FullName:      fmt.Sprintf("param.%s", block.Labels[0]),
		referencesMap: make(ResourceReferenceMap),
		parent:        parent,
		DeclRange:     block.DefRange,
	}
}

func (p ParamDef) String() string {
	return fmt.Sprintf("Name: %s, Description: %s, Default: %s", p.FullName, typehelpers.SafeString(p.Description), typehelpers.SafeString(p.Default))
}

func (p ParamDef) Equals(other *ParamDef) bool {
	return p.ShortName == other.ShortName &&
		typehelpers.SafeString(p.Description) == typehelpers.SafeString(other.Description) &&
		typehelpers.SafeString(p.Default) == typehelpers.SafeString(other.Default)
}
