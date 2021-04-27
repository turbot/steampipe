package modconfig

import (
	"fmt"
	"strings"
)

type ModBlockType string

const (
	BlockTypeMod          ModBlockType = "mod"
	BlockTypeQuery                     = "query"
	BlockTypeControl                   = "control"
	BlockTypeControlGroup              = "control_group"
	BlockTypeLocals                    = "locals"
)

type ModResourceName struct {
	Mod      string
	ItemType ModBlockType
	Name     string
}

type ModResourcePropertyPath struct {
	Mod          string
	ItemType     ModBlockType
	Name         string
	PropertyPath []string
}

func ParseModResourceName(fullName string) (res *ModResourceName, err error) {
	if fullName == "" {
		return nil, fmt.Errorf("empty name passed to ParseModResourceName")
	}
	res = &ModResourceName{}

	parts := strings.Split(fullName, ".")

	switch len(parts) {
	case 0:
		err = fmt.Errorf("empty name passed to ParseModResourceName")
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
		err = fmt.Errorf("invalid name '%s' passed to ParseModResourceName", fullName)
	}

	return
}

func ParseModResourcePropertyPath(propertyPath string) (res *ModResourcePropertyPath, err error) {
	res = &ModResourcePropertyPath{}

	// valid property paths:
	// <mod>.<resource>.<name>.<property path...>
	// <resource>.<name>.<property path...>
	// so either the first or second slice must be a valid resource type
	// and len must be at least 3
	parts := strings.Split(propertyPath, ".")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid property path '%s' passed to ParseModResourcePropertyPath", propertyPath)
	}

	switch len(parts) {
	case 0:
		err = fmt.Errorf("empty name passed to ParseModResourceName")
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

func BuildModResourceName(blockType ModBlockType, name string) string {
	return fmt.Sprintf("%s.%s", string(blockType), name)
}
