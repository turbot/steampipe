package parse

import (
	"github.com/hashicorp/hcl/v2"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ParseModFlag uint32

const (
	CreateDefaultMod ParseModFlag = 1 << iota
	CreatePseudoResources
)

type ParseModOptions struct {
	Flags       ParseModFlag
	ListOptions *filehelpers.ListOptions
	// if set, only decode these blocks
	BlockTypes []string
	// if set, exclude these block types
	BlockTypeExclusions []string
	Variables           map[string]*modconfig.Variable
}

func (o *ParseModOptions) CreateDefaultMod() bool {
	return o.Flags&CreateDefaultMod == CreateDefaultMod
}

func (o *ParseModOptions) CreatePseudoResources() bool {
	return o.Flags&CreatePseudoResources == CreatePseudoResources
}

func (o *ParseModOptions) ShouldIncludeBlock(block *hcl.Block) bool {
	if len(o.BlockTypes) > 0 && !helpers.StringSliceContains(o.BlockTypes, block.Type) {
		return false
	}
	if len(o.BlockTypeExclusions) > 0 && helpers.StringSliceContains(o.BlockTypes, block.Type) {
		return false
	}
	return true
}
