package modconfig

import (
	"fmt"
	"strings"
)

type ParsedPropertyPath struct {
	Mod          string
	ItemType     string
	Name         string
	PropertyPath []string
	// optional scope of this property path ("root or parent")
	Scope    string
	original string
}

func (p *ParsedPropertyPath) PropertyPathString() string {
	return strings.Join(p.PropertyPath, ".")
}

func (p *ParsedPropertyPath) ToParsedResourceName() *ParsedResourceName {
	return &ParsedResourceName{
		Mod:      p.Mod,
		ItemType: p.ItemType,
		Name:     p.Name,
	}
}

func (p *ParsedPropertyPath) ToResourceName() string {
	return BuildModResourceName(p.ItemType, p.Name)
}

func (p *ParsedPropertyPath) String() string {
	return p.original
}

func ParseResourcePropertyPath(propertyPath string) (res *ParsedPropertyPath, err error) {
	res = &ParsedPropertyPath{original: propertyPath}

	// valid property paths:
	// <mod>.<resource>.<name>.<property path...>
	// <resource>.<name>.<property path...>
	// so either the first or second slice must be a valid resource type

	parts := strings.Split(propertyPath, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid property path '%s' passed to ParseResourcePropertyPath", propertyPath)
	}

	// special case handling for runtime dependencies which may have use the "self" qualifier
	if parts[0] == runtimeDependencyReportScope {
		res.Scope = parts[0]
		parts = parts[1:]
	}

	switch len(parts) {
	case 2:
		// no property path specified
		res.ItemType = parts[0]
		res.Name = parts[1]
	case 3:
		res.ItemType = parts[0]
		res.Name = parts[1]
		res.PropertyPath = parts[2:]
	default:
		res.Mod = parts[0]
		res.ItemType = parts[1]
		res.Name = parts[2]
		res.PropertyPath = parts[2:]
	}

	return
}
