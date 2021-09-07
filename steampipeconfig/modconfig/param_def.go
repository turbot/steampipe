package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

type ParamDef struct {
	Name        string      `cty:"name"`
	Description *string     `cty:"description"`
	RawDefault  interface{} `json:"-"`
	Default     *string     `cty:"default"`
}

func NewParamDef(block *hcl.Block) *ParamDef {
	return &ParamDef{Name: block.Labels[0]}
}

func (d ParamDef) String() string {
	return fmt.Sprintf("Name: %s, Description: %s, Default: %s", d.Name, typehelpers.SafeString(d.Description), typehelpers.SafeString(d.Default))
}

func (d ParamDef) Equals(other *ParamDef) bool {
	return d.Name == other.Name &&
		typehelpers.SafeString(d.Description) == typehelpers.SafeString(other.Description) &&
		typehelpers.SafeString(d.Default) == typehelpers.SafeString(other.Default)
}
