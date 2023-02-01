package modconfig

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

// DashboardTable is a struct representing a leaf dashboard node
type DashboardTable struct {
	ResourceWithMetadataImpl
	QueryProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	// TODO remove - check introspection tables
	Width      *int                             `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type       *string                          `cty:"type" hcl:"type" column:"type,text" json:"-"`
	ColumnList DashboardTableColumnList         `cty:"column_list" hcl:"column,block" column:"columns,jsonb" json:"-"`
	Columns    map[string]*DashboardTableColumn `cty:"columns" json:"columns,omitempty"`
	Display    *string                          `cty:"display" hcl:"display" json:"display,omitempty"`
	Base       *DashboardTable                  `hcl:"base" json:"-"`
}

func NewDashboardTable(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	t := &DashboardTable{
		QueryProviderImpl: QueryProviderImpl{
			RuntimeDependencyProviderImpl: RuntimeDependencyProviderImpl{
				ModTreeItemImpl: ModTreeItemImpl{
					HclResourceImpl: HclResourceImpl{
						ShortName:       shortName,
						FullName:        fullName,
						UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
						DeclRange:       BlockRange(block),
						blockType:       block.Type,
					},
					Mod: mod,
				},
			},
		},
	}
	t.SetAnonymous(block)
	return t
}

// NewQueryDashboardTable creates a Table to wrap a query.
// This is used in order to execute queries as dashboards
func NewQueryDashboardTable(q ModTreeItem) (*DashboardTable, error) {
	parsedName, err := ParseResourceName(q.Name())
	if err != nil {
		return nil, err
	}

	queryProvider, ok := q.(QueryProvider)
	if !ok {
		return nil, fmt.Errorf("rersource passed to NewQueryDashboardTable must implement QueryProvider")
	}

	tableName := BuildFullResourceName(q.GetMod().ShortName, BlockTypeTable, parsedName.Name)
	c := &DashboardTable{
		ResourceWithMetadataImpl: ResourceWithMetadataImpl{
			metadata: &ResourceMetadata{},
		},
		QueryProviderImpl: QueryProviderImpl{
			RuntimeDependencyProviderImpl: RuntimeDependencyProviderImpl{
				ModTreeItemImpl: ModTreeItemImpl{
					HclResourceImpl: HclResourceImpl{
						ShortName:       parsedName.Name,
						FullName:        tableName,
						UnqualifiedName: fmt.Sprintf("%s.%s", BlockTypeTable, parsedName),
						Title:           utils.ToStringPointer(q.GetTitle()),
						blockType:       BlockTypeTable,
					},
					Mod: q.GetMod(),
				},
			},
			Query: queryProvider.GetQuery(),
			SQL:   queryProvider.GetSQL(),
		},
	}
	return c, nil
}

func (t *DashboardTable) Equals(other *DashboardTable) bool {
	diff := t.Diff(other)
	return !diff.HasChanges()
}

// OnDecoded implements HclResource
func (t *DashboardTable) OnDecoded(_ *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	t.setBaseProperties()
	// populate columns map
	if len(t.ColumnList) > 0 {
		t.Columns = make(map[string]*DashboardTableColumn, len(t.ColumnList))
		for _, c := range t.ColumnList {
			t.Columns[c.Name] = c
		}
	}
	return nil
}

func (t *DashboardTable) Diff(other *DashboardTable) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: t,
		Name: t.Name(),
	}

	if !utils.SafeStringsEqual(t.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if len(t.ColumnList) != len(other.ColumnList) {
		res.AddPropertyDiff("Columns")
	} else {
		for i, c := range t.Columns {
			if !c.Equals(other.Columns[i]) {
				res.AddPropertyDiff("Columns")
			}
		}
	}

	res.populateChildDiffs(t, other)
	res.queryProviderDiff(t, other)
	res.dashboardLeafNodeDiff(t, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (t *DashboardTable) GetWidth() int {
	if t.Width == nil {
		return 0
	}
	return *t.Width
}

// GetDisplay implements DashboardLeafNode
func (t *DashboardTable) GetDisplay() string {
	return typehelpers.SafeString(t.Display)
}

// GetDocumentation implements DashboardLeafNode, ModTreeItem
func (*DashboardTable) GetDocumentation() string {
	return ""
}

// GetType implements DashboardLeafNode
func (t *DashboardTable) GetType() string {
	return typehelpers.SafeString(t.Type)
}

// CtyValue implements CtyValueProvider
func (t *DashboardTable) CtyValue() (cty.Value, error) {
	return GetCtyValue(t)
}

func (t *DashboardTable) setBaseProperties() {
	if t.Base == nil {
		return
	}
	// copy base into the HclResourceImpl 'base' property so it is accessible to all nested structs
	t.base = t.Base
	// call into parent nested struct setBaseProperties
	t.QueryProviderImpl.setBaseProperties()

	if t.Width == nil {
		t.Width = t.Base.Width
	}

	if t.Type == nil {
		t.Type = t.Base.Type
	}

	if t.Display == nil {
		t.Display = t.Base.Display
	}

	if t.ColumnList == nil {
		t.ColumnList = t.Base.ColumnList
	} else {
		t.ColumnList.Merge(t.Base.ColumnList)
	}
}
