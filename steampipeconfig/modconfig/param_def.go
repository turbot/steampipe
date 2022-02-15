package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

type ParamDef struct {
	Name        string      `cty:"name" json:"name"`
	FullName    string      `cty:"full_name" json:"-"`
	Description *string     `cty:"description" json:"description"`
	RawDefault  interface{} `json:"-"`
	Default     *string     `cty:"default" json:"default"`

	// list of all blocks referenced by the resource
	References []*ResourceReference
	DeclRange  hcl.Range
}

func NewParamDef(block *hcl.Block) *ParamDef {
	return &ParamDef{
		Name:      block.Labels[0],
		FullName:  fmt.Sprintf("param.%s", block.Labels[0]),
		DeclRange: block.DefRange,
	}
}

func (p ParamDef) String() string {
	return fmt.Sprintf("Name: %s, Description: %s, Default: %s", p.FullName, typehelpers.SafeString(p.Description), typehelpers.SafeString(p.Default))
}

func (p ParamDef) Equals(other *ParamDef) bool {
	return p.Name == other.Name &&
		typehelpers.SafeString(p.Description) == typehelpers.SafeString(other.Description) &&
		typehelpers.SafeString(p.Default) == typehelpers.SafeString(other.Default)
}
