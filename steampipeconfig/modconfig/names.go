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
)

type ModResourceName struct {
	Mod      string
	ItemType ModBlockType
	Name     string
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
