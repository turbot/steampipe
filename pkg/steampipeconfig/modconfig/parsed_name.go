package modconfig

import (
	"fmt"
	"strings"
)

type ParsedResourceName struct {
	Mod      string
	ItemType string
	Name     string
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
	if !IsValidResourceItemType(res.ItemType) {
		err = fmt.Errorf("invalid name '%s' passed to ParseResourceName", fullName)
	}
	return
}

func (p *ParsedResourceName) ToResourceName() string {
	return BuildModResourceName(p.ItemType, p.Name)
}

func (p *ParsedResourceName) ToFullName() (string, error) {
	fullName, err := BuildFullResourceName(p.Mod, p.ItemType, p.Name)
	if err != nil {
		return "", err
	}
	return fullName, nil
}

func (p *ParsedResourceName) ToFullNameWithMod(mod string) (string, error) {
	if p.Mod != "" {
		fullName, err := p.ToFullName()
		if err != nil {
			return "", err
		}
		return fullName, nil
	}
	fullResourceName, err := BuildFullResourceName(mod, p.ItemType, p.Name)
	if err != nil {
		return "", err
	}
	return fullResourceName, nil
}

// BuildFullResourceName generates a fully qualified name from the given components
// e.g: aws_compliance.benchmark.cis_v150_1
// TODO: validate argument are not empty strings [https://github.com/turbot/steampipe/issues/3601]
func BuildFullResourceName(mod, blockType, name string) (string, error) {
	if mod == "" {
		return "", fmt.Errorf("mod name not provided")
	}
	if blockType == "" {
		return "", fmt.Errorf("block type not provided")
	}
	if name == "" {
		return "", fmt.Errorf("resource name not provided")
	}
	return fmt.Sprintf("%s.%s.%s", mod, blockType, name), nil
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

func BuildModResourceName(blockType, name string) string {
	return fmt.Sprintf("%s.%s", blockType, name)
}
