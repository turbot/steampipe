package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

type HclResourceBase struct {
	HclResourceRemain hcl.Body `hcl:",remain" json:"-"`

	FullName            string            `cty:"name" json:"-"`
	Title               *string           `cty:"title" hcl:"title" column:"title,text" json:"-"`
	ShortName           string            `cty:"short_name" hcl:"name,label" json:"-"`
	UnqualifiedName     string            `cty:"unqualified_name" json:"unqualified_name"`
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
func (b *HclResourceBase) Name() string {
	return b.FullName
}

// GetTitle implements HclResource
func (b *HclResourceBase) GetTitle() string {
	return typehelpers.SafeString(b.Title)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (b *HclResourceBase) GetUnqualifiedName() string {
	return b.UnqualifiedName
}

// OnDecoded implements HclResource
func (b *HclResourceBase) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	return nil
}

// GetDeclRange implements HclResource
func (b *HclResourceBase) GetDeclRange() *hcl.Range {
	return &b.DeclRange
}

// BlockType implements HclResource
func (b *HclResourceBase) BlockType() string {
	return b.blockType
}

// GetDescription implements HclResource
func (b *HclResourceBase) GetDescription() string {
	return typehelpers.SafeString(b.Description)
}

// GetDocumentation implements HclResource
func (b *HclResourceBase) GetDocumentation() string {
	return typehelpers.SafeString(b.Documentation)
}

// GetTags implements HclResource
func (b *HclResourceBase) GetTags() map[string]string {
	if b.Tags != nil {
		return b.Tags
	}
	return map[string]string{}
}

// GetHclResourceBase implements HclResource
func (b *HclResourceBase) GetHclResourceBase() *HclResourceBase {
	return b
}

// SetTopLevel implements HclResource
func (b *HclResourceBase) SetTopLevel(isTopLevel bool) {
	b.isTopLevel = isTopLevel
}

// IsTopLevel implements HclResource
func (b *HclResourceBase) IsTopLevel() bool {
	return b.isTopLevel
}

// CtyValue implements CtyValueProvider
func (b *HclResourceBase) CtyValue() (cty.Value, error) {
	if b.disableCtySerialise {
		return cty.Zero, nil
	}
	return GetCtyValue(b)
}
