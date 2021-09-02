package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

type ParamDef struct {
	Name        string  `cty:"name"`
	Description *string `cty:"description" column:"description,text"`
	RawDefault  interface{}
	Default     *string `cty:"default" column:"default,text"`
}

func NewParamDef(block *hcl.Block) *ParamDef {
	return &ParamDef{Name: block.Labels[0]}
}

func (d ParamDef) String() string {
	return fmt.Sprintf("Name: %s, Description: %s, Default: %s", d.Name, typehelpers.SafeString(d.Description), typehelpers.SafeString(d.Default))
}
