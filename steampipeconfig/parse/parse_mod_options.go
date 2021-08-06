package parse

import (
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/zclconf/go-cty/cty"
)

type ParseModFlag uint32

const (
	CreateDefaultMod ParseModFlag = 1 << iota
	CreatePseudoResources
)

type ParseModOptions struct {
	Flags       ParseModFlag
	ListOptions *filehelpers.ListOptions
	Variables   map[string]cty.Value
	// if set, only decode these blocks
	BlockTypes []string
}

func (o *ParseModOptions) CreateDefaultMod() bool {
	return o.Flags&CreateDefaultMod == CreateDefaultMod
}

func (o *ParseModOptions) CreatePseudoResources() bool {
	return o.Flags&CreatePseudoResources == CreatePseudoResources
}
