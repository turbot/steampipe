package modconfig

import (
	"fmt"
	"strings"
)

type ModBlockType string

const (
	BlockTypeMod       ModBlockType = "mod"
	BlockTypeQuery                  = "query"
	BlockTypeControl                = "control"
	BlockTypeBenchmark              = "benchmark"
	BlockTypeReport                 = "report"
	BlockTypePanel                  = "panel"
	BlockTypeLocals                 = "locals"
)

type ParsedResourceName struct {
	Mod      string
	ItemType ModBlockType
	Name     string
}

func (m *ParsedResourceName) TypeString() string {
	return string(m.ItemType)
}

type ParsedPropertyPath struct {
	Mod          string
	ItemType     ModBlockType
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
		res.ItemType = ModBlockType(parts[0])
		res.Name = parts[1]
	case 3:
		res.Mod = parts[0]
		res.ItemType = ModBlockType(parts[1])
		res.Name = parts[2]
	default:
		err = fmt.Errorf("invalid name '%s' passed to ParseResourceName", fullName)
	}

	return
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
		res.ItemType = ModBlockType(parts[0])
		res.Name = parts[1]
	case 3:
		res.ItemType = ModBlockType(parts[0])
		res.Name = parts[1]
		res.PropertyPath = parts[2:]
	default:
		res.Mod = parts[0]
		res.ItemType = ModBlockType(parts[1])
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

func BuildModResourceName(blockType ModBlockType, name string) string {
	return fmt.Sprintf("%s.%s", string(blockType), name)
}
