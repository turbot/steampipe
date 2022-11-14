package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

type HclResourceBase struct {
	FullName        string  `cty:"name" json:"-"`
	Title           *string `cty:"title" hcl:"title" column:"title,text" json:"-"`
	ShortName       string  `cty:"short_name" hcl:"name,label" json:"-"`
	UnqualifiedName string  `json:"-"`
	Description     *string `cty:"description" hcl:"description" column:"description,text" json:"-"`

	DeclRange hcl.Range         `json:"-"`
	Tags      map[string]string `cty:"tags" hcl:"tags,optional" column:"tags,jsonb" json:"-"`
	blockType string
}

// Name implements HclResource
// return name in format: '<blocktype>.<shortName>'
func (i *HclResourceBase) Name() string {
	return i.FullName
}

// GetTitle implements HclResource
func (i *HclResourceBase) GetTitle() string {
	return typehelpers.SafeString(i.Title)
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (i *HclResourceBase) GetUnqualifiedName() string {
	return i.UnqualifiedName
}

// OnDecoded implements HclResource
func (*HclResourceBase) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	return nil
}

// CtyValue implements HclResource
func (i *HclResourceBase) CtyValue() (cty.Value, error) {
	return getCtyValue(i)
}

// GetDeclRange implements HclResource
func (i *HclResourceBase) GetDeclRange() *hcl.Range {
	return &i.DeclRange
}

// BlockType implements HclResource
func (i *HclResourceBase) BlockType() string {
	return i.blockType
}

// GetDescription implements HclResource
func (i *HclResourceBase) GetDescription() string {
	return typehelpers.SafeString(i.Description)
}

// GetTags implements HclResource
func (i *HclResourceBase) GetTags() map[string]string {
	return map[string]string{}
}
