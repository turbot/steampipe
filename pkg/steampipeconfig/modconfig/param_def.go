package modconfig

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

type ParamDef struct {
	ShortName       string  `cty:"name" json:"name"`
	UnqualifiedName string  `cty:"full_name" json:"-"`
	Description     *string `cty:"description" json:"description"`
	Default         *string `cty:"default" json:"default"`
	// tactical - is the raw value a string
	IsString bool `cty:"is_string" json:"-"`

	// list of all blocks referenced by the resource
	References []*ResourceReference `json:"-"`
	DeclRange  hcl.Range            `json:"-"`
}

func NewParamDef(block *hcl.Block) *ParamDef {
	return &ParamDef{
		ShortName:       block.Labels[0],
		UnqualifiedName: fmt.Sprintf("param.%s", block.Labels[0]),
		DeclRange:       BlockRange(block),
	}
}

func (p *ParamDef) String() string {
	return fmt.Sprintf("Name: %s, Description: %s, Default: %s", p.UnqualifiedName, typehelpers.SafeString(p.Description), typehelpers.SafeString(p.Default))
}

func (p *ParamDef) Equals(other *ParamDef) bool {
	return p.ShortName == other.ShortName &&
		typehelpers.SafeString(p.Description) == typehelpers.SafeString(other.Description) &&
		typehelpers.SafeString(p.Default) == typehelpers.SafeString(other.Default)
}

// SetDefault sets the default as a atring points, marshalling to json is the underlying value is NOT a string
func (p *ParamDef) SetDefault(value interface{}) error {
	strVal, ok := value.(string)
	if ok {
		p.IsString = true
		// no need to convert to string
		p.Default = &strVal
		return nil
	}
	// format the arg value as a JSON string
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	def := string(jsonBytes)
	p.Default = &def
	return nil
}

// GetDefault returns the default as an interface{}, unmarshalling json is the underlying value was NOT a string
func (p *ParamDef) GetDefault() (any, error) {
	if p.Default == nil {
		return nil, nil
	}
	if p.IsString {
		return *p.Default, nil
	}
	var val any
	err := json.Unmarshal([]byte(*p.Default), &val)
	return val, err
}
