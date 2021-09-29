package modconfig

import (
	"fmt"
)

type ResourceReference struct {
	Name   string `json:"name"`
	Parent string `json:"parent,omitempty"`
}

func NewResourceReference(reference HclResource) ResourceReference {
	res := ResourceReference{
		Name: reference.Name(),
	}
	// special case code for param - set the parent
	if paramDef, ok := reference.(*ParamDef); ok {
		res.Parent = paramDef.parent
	}
	return res
}

func (r ResourceReference) String() string {
	return fmt.Sprintf("%s.%s", r.Name, r.Parent)
}
