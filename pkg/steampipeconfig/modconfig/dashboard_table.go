package modconfig

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
)

// DashboardTable is a struct representing a leaf dashboard node
type DashboardTable struct {
	ResourceWithMetadataBase
	QueryProviderBase
	ModTreeItemBase

	// required to allow partial decoding
	Remain     hcl.Body                         `hcl:",remain" json:"-"`
	Width      *int                             `cty:"width" hcl:"width" column:"width,text" json:"-"`
	Type       *string                          `cty:"type" hcl:"type" column:"type,text" json:"-"`
	ColumnList DashboardTableColumnList         `cty:"column_list" hcl:"column,block" column:"columns,jsonb" json:"-"`
	Columns    map[string]*DashboardTableColumn `cty:"columns" json:"columns,omitempty"`
	Display    *string                          `cty:"display" hcl:"display" json:"display,omitempty"`
	Base       *DashboardTable                  `hcl:"base" json:"-"`
	References []*ResourceReference             `json:"-"`
}

func NewDashboardTable(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)

	t := &DashboardTable{
		QueryProviderBase: QueryProviderBase{
			modNameWithVersion: mod.NameWithVersion(),
			HclResourceBase: HclResourceBase{
				ShortName:       shortName,
				FullName:        fullName,
				UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
		},
		ModTreeItemBase: ModTreeItemBase{
			Mod:      mod,
			fullName: fullName,
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
		ResourceWithMetadataBase: ResourceWithMetadataBase{
			metadata: &ResourceMetadata{},
		},
		QueryProviderBase: QueryProviderBase{
			modNameWithVersion: q.GetMod().NameWithVersion(),
			Query:              queryProvider.GetQuery(),
			SQL:                queryProvider.GetSQL(),
			HclResourceBase: HclResourceBase{
				ShortName:       parsedName.Name,
				FullName:        tableName,
				UnqualifiedName: fmt.Sprintf("%s.%s", BlockTypeTable, parsedName),
				Title:           utils.ToStringPointer(q.GetTitle()),
				blockType:       BlockTypeTable,
			},
		},
		ModTreeItemBase: ModTreeItemBase{
			Mod:      q.GetMod(),
			fullName: tableName,
		},
	}
	return c, nil
}

func (t *DashboardTable) Equals(other *DashboardTable) bool {
	diff := t.Diff(other)
	return !diff.HasChanges()
}

// AddReference implements ResourceWithMetadata
func (t *DashboardTable) AddReference(ref *ResourceReference) {
	t.References = append(t.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (t *DashboardTable) GetReferences() []*ResourceReference {
	return t.References
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

func (t *DashboardTable) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	// not all base properties are stored in the evalContext
	// (e.g. resource metadata and runtime dependencies are not stores)
	//  so resolve base from the resource map provider (which is the RunContext)
	if base, resolved := resolveBase(t.Base, resourceMapProvider); !resolved {
		return
	} else {
		t.Base = base.(*DashboardTable)
	}

	if t.Title == nil {
		t.Title = t.Base.Title
	}

	if t.Width == nil {
		t.Width = t.Base.Width
	}

	if t.SQL == nil {
		t.SQL = t.Base.SQL
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
