package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

type HclResourceBase struct {
	HclResourceRemain hcl.Body `hcl:",remain" json:"-"`

	FullName            string            `cty:"name" json:"-"`
	Title               *string           `cty:"title" hcl:"title" column:"title,text" json:"-"`
	ShortName           string            `cty:"short_name" hcl:"name,label" json:"-"`
	UnqualifiedName     string            `json:"-"`
	Description         *string           `cty:"description" hcl:"description" column:"description,text" json:"-"`
	Documentation       *string           `cty:"documentation" hcl:"documentation" column:"documentation,text"`
	DeclRange           hcl.Range         `json:"-"`
	Tags                map[string]string `cty:"tags" hcl:"tags,optional" column:"tags,jsonb" json:"-"`
	blockType           string
	disableCtySerialise bool
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
func (*HclResourceBase) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
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
	return map[string]string{}
}

// GetHclResourceBase implements HclResource
func (b *HclResourceBase) GetHclResourceBase() *HclResourceBase {
	return b
}

// ShouldCtySerialise implements ModTreeItem
// allows disabling of base class serialization, used for Local
func (b *HclResourceBase) ShouldCtySerialise() bool {
	return !b.disableCtySerialise
}
