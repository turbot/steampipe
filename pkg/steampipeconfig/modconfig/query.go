package modconfig

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
)

// Query is a struct representing the Query resource
type Query struct {
	ResourceWithMetadataImpl
	QueryProviderImpl

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	// only here as otherwise gocty.ImpliedType panics
	Unused string `cty:"unused" json:"-"`
}

func NewQuery(block *hcl.Block, mod *Mod, shortName string) HclResource {
	fullName := fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName)
	// queries cannot be anonymous
	q := &Query{
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
	return q
}

func QueryFromFile(modPath, filePath string, mod *Mod) (MappableResource, []byte, error) {
	q := &Query{
		QueryProviderImpl: QueryProviderImpl{
			RuntimeDependencyProviderImpl: RuntimeDependencyProviderImpl{
				ModTreeItemImpl: ModTreeItemImpl{
					Mod: mod,
				},
			},
		},
	}
	return q.InitialiseFromFile(modPath, filePath)
}

// InitialiseFromFile implements MappableResource
func (q *Query) InitialiseFromFile(modPath, filePath string) (MappableResource, []byte, error) {
	// only valid for sql files
	if filepath.Ext(filePath) != constants.SqlExtension {
		return nil, nil, fmt.Errorf("Query.InitialiseFromFile must be called with .sql files only - filepath: '%s'", filePath)
	}

	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	sql := string(sqlBytes)
	if sql == "" {
		log.Printf("[TRACE] SQL file %s contains no query", filePath)
		return nil, nil, nil
	}
	// get a sluggified version of the filename
	name, err := PseudoResourceNameFromPath(modPath, filePath)
	if err != nil {
		return nil, nil, err
	}
	q.ShortName = name
	q.UnqualifiedName = fmt.Sprintf("query.%s", name)
	q.FullName = fmt.Sprintf("%s.query.%s", q.Mod.ShortName, name)
	q.SQL = &sql
	q.DeclRange = hcl.Range{
		Filename: filePath,
		Start: hcl.Pos{
			Line:   0,
			Column: 0,
			Byte:   0,
		},
		End: hcl.Pos{
			Line: len(sql),
		},
	}

	return q, sqlBytes, nil
}

func (q *Query) Equals(other *Query) bool {
	res := q.ShortName == other.ShortName &&
		q.FullName == other.FullName &&
		typehelpers.SafeString(q.Description) == typehelpers.SafeString(other.Description) &&
		typehelpers.SafeString(q.Documentation) == typehelpers.SafeString(other.Documentation) &&
		typehelpers.SafeString(q.SQL) == typehelpers.SafeString(other.SQL) &&
		typehelpers.SafeString(q.Title) == typehelpers.SafeString(other.Title)
	if !res {
		return res
	}

	// tags
	if q.Tags == nil {
		if other.Tags != nil {
			return false
		}
	} else {
		// we have tags
		if other.Tags == nil {
			return false
		}
		for k, v := range q.Tags {
			if otherVal, ok := (other.Tags)[k]; !ok && v != otherVal {
				return false
			}
		}
	}

	// params
	if len(q.Params) != len(other.Params) {
		return false
	}
	for i, p := range q.Params {
		if !p.Equals(other.Params[i]) {
			return false
		}
	}

	return true
}

func (q *Query) String() string {
	res := fmt.Sprintf(`
  -----
  Name: %s
  Title: %s
  Description: %s
  SQL: %s
`, q.FullName, types.SafeString(q.Title), types.SafeString(q.Description), types.SafeString(q.SQL))

	// add param defs if there are any
	if len(q.Params) > 0 {
		var paramDefsStr = make([]string, len(q.Params))
		for i, def := range q.Params {
			paramDefsStr[i] = def.String()
		}
		res += fmt.Sprintf("Params:\n\t%s\n  ", strings.Join(paramDefsStr, "\n\t"))
	}
	return res
}

// OnDecoded implements HclResource
func (q *Query) OnDecoded(*hcl.Block, ResourceMapsProvider) hcl.Diagnostics {
	return nil
}

// CtyValue implements CtyValueProvider
func (q *Query) CtyValue() (cty.Value, error) {
	return GetCtyValue(q)
}

func (q *Query) Diff(other *Query) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: q,
		Name: q.Name(),
	}

	if !utils.SafeStringsEqual(q.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	res.populateChildDiffs(q, other)
	res.queryProviderDiff(q, other)

	return res
}
