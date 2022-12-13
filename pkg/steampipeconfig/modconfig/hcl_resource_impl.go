package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

type HclResourceImpl struct {
	// required to allow partial decoding
	HclResourceRemain hcl.Body `hcl:",remain" json:"-"`

	FullName            string            `cty:"name" json:"-"`
	Title               *string           `cty:"title" hcl:"title" column:"title,text" json:"-"`
	ShortName           string            `cty:"short_name" hcl:"name,label" json:"name"`
	UnqualifiedName     string            `cty:"unqualified_name" json:"-"`
	Description         *string           `cty:"description" hcl:"description" column:"description,text" json:"-"`
	Documentation       *string           `cty:"documentation" hcl:"documentation" column:"documentation,text" json:"-"`
	DeclRange           hcl.Range         `json:"-"`
	Tags                map[string]string `cty:"tags" hcl:"tags,optional" column:"tags,jsonb" json:"-"`
	blockType           string
	disableCtySerialise bool
	isTopLevel          bool
}

// Name implements HclResource
// return name in format: '<blocktype>.<shortName>'
func (b *HclResourceImpl) Name() string {
	return b.FullName
}

// GetTitle implements HclResource
func (b *HclResourceImpl) GetTitle() string {
	return typehelpers.SafeString(b.Title)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (b *HclResourceImpl) GetUnqualifiedName() string {
	return b.UnqualifiedName
}

// OnDecoded implements HclResource
func (b *HclResourceImpl) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	return nil
}

// GetDeclRange implements HclResource
func (b *HclResourceImpl) GetDeclRange() *hcl.Range {
	return &b.DeclRange
}

// BlockType implements HclResource
func (b *HclResourceImpl) BlockType() string {
	return b.blockType
}

// GetDescription implements HclResource
func (b *HclResourceImpl) GetDescription() string {
	return typehelpers.SafeString(b.Description)
}

// GetDocumentation implements HclResource
func (b *HclResourceImpl) GetDocumentation() string {
	return typehelpers.SafeString(b.Documentation)
}

// GetTags implements HclResource
func (b *HclResourceImpl) GetTags() map[string]string {
	if b.Tags != nil {
		return b.Tags
	}
	return map[string]string{}
}

// GetHclResourceBase implements HclResource
func (b *HclResourceImpl) GetHclResourceImpl() *HclResourceImpl {
	return b
}

// SetTopLevel implements HclResource
func (b *HclResourceImpl) SetTopLevel(isTopLevel bool) {
	b.isTopLevel = isTopLevel
}

// IsTopLevel implements HclResource
func (b *HclResourceImpl) IsTopLevel() bool {
	return b.isTopLevel
}

// CtyValue implements CtyValueProvider
func (b *HclResourceImpl) CtyValue() (cty.Value, error) {
	if b.disableCtySerialise {
		return cty.Zero, nil
	}
	return GetCtyValue(b)
}
