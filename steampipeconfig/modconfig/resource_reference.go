package modconfig

import (
	"fmt"
)

type ResourceReference struct {
	Name      string `cty:"name" json:"name"`
	BlockType string `cty:"block_type" json:"block_type"`
	BlockName string `cty:"block_name" json:"block_name"`
	Attribute string `cty:"attribute" json:"attribute"`
}

//
//func NewResourceReference(reference HclResource) ResourceReference {
//	// special case code for param - set the param parent as the reference name and the param name as the child
//	if paramDef, ok := reference.(*ParamDef); ok {
//		return ResourceReference{
//			Name:  paramDef.parent,
//			Child: reference.Name(),
//		}
//	}
//
//	return ResourceReference{
//		Name: reference.Name(),
//	}
//
//}

func (r ResourceReference) String() string {
	return fmt.Sprintf("%s.%s.%s", r.Name, r.BlockType, r.BlockName)
}
