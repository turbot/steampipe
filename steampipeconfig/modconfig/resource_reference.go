package modconfig

import (
	"fmt"
)

type ResourceReference struct {
	Name   string `json:"name"`
	Parent string `json:"parent"`
}

func NewResourceReference(reference HclResource) ResourceReference {
	return ResourceReference{
		Name:   reference.Name(),
		Parent: reference.Parent(),
	}
}

func (r ResourceReference) String() string {
	return fmt.Sprintf("%s.%s", r.Name, r.Parent)
}
