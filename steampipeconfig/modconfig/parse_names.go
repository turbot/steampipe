package modconfig

import (
	"fmt"
	"strings"
)

const (
	BlockTypeMod       = "mod"
	BlockTypeQuery     = "query"
	BlockTypeControl   = "control"
	BlockTypeBenchmark = "benchmark"
	BlockTypeReport    = "report"
	BlockTypeContainer = "container"
	BlockTypePanel     = "panel"
	BlockTypeLocals    = "locals"
	BlockTypeVariable  = "variable"
	BlockTypeParam     = "param"
)

type ParsedResourceName struct {
	Mod      string
	ItemType string
	Name     string
}

type ParsedPropertyPath struct {
	Mod          string
	ItemType     string
	Name         string
	PropertyPath []string
}

func ParseResourceName(fullName string) (res *ParsedResourceName, err error) {
	if fullName == "" {
		return &ParsedResourceName{}, nil
	}
	res = &ParsedResourceName{}

	parts := strings.Split(fullName, ".")

	switch len(parts) {
	case 0:
		err = fmt.Errorf("empty name passed to ParseResourceName")
	case 1:
		res.Name = parts[0]
	case 2:
		res.ItemType = parts[0]
		res.Name = parts[1]
	case 3:
		res.Mod = parts[0]
		res.ItemType = parts[1]
		res.Name = parts[2]
	default:
		err = fmt.Errorf("invalid name '%s' passed to ParseResourceName", fullName)
	}

	return
}

// UnqualifiedResourceName removes the mod prefix from the given name
func UnqualifiedResourceName(fullName string) string {
	parts := strings.Split(fullName, ".")
	switch len(parts) {
	case 3:
		return strings.Join(parts[1:], ".")
	default:
		return fullName
	}
}

func ParseResourcePropertyPath(propertyPath string) (res *ParsedPropertyPath, err error) {
	res = &ParsedPropertyPath{}

	// valid property paths:
	// <mod>.<resource>.<name>.<property path...>
	// <resource>.<name>.<property path...>
	// so either the first or second slice must be a valid resource type

	parts := strings.Split(propertyPath, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid property path '%s' passed to ParseResourcePropertyPath", propertyPath)
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

func PropertyPathToResourceName(propertyPath string) (string, error) {
	parsedPropertyPath, err := ParseResourcePropertyPath(propertyPath)
	if err != nil {
		return "", err
	}
	return BuildModResourceName(parsedPropertyPath.ItemType, parsedPropertyPath.Name), nil
}

func BuildModResourceName(blockType string, name string) string {
	return fmt.Sprintf("%s.%s", blockType, name)
}
