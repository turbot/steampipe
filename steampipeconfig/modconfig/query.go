package modconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// Query is a struct representing the Query resource
type Query struct {
	ResourceWithMetadataBase
	QueryProviderBase

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain"`

	ShortName string `cty:"short_name"`
	FullName  string `cty:"name"`

	Description           *string           `cty:"description" hcl:"description" column:"description,text"`
	Documentation         *string           `cty:"documentation" hcl:"documentation" column:"documentation,text"`
	SearchPath            *string           `cty:"search_path" hcl:"search_path" column:"search_path,text"`
	SearchPathPrefix      *string           `cty:"search_path_prefix" hcl:"search_path_prefix" column:"search_path_prefix,text"`
	Tags                  map[string]string `cty:"tags" hcl:"tags,optional" column:"tags,jsonb"`
	Title                 *string           `cty:"title" hcl:"title" column:"title,text"`
	PreparedStatementName string            `column:"prepared_statement_name,text" json:"-"`
	SQL                   *string           `cty:"sql" hcl:"sql" column:"sql,text"`

	Params []*ParamDef `cty:"params" column:"params,jsonb"`
	// list of all blocks referenced by the resource
	References []*ResourceReference

	Mod       *Mod `cty:"mod"`
	DeclRange hcl.Range

	UnqualifiedName string
	Paths           []NodePath `column:"path,jsonb"`
	parents         []ModTreeItem
}

func NewQuery(block *hcl.Block, mod *Mod, shortName string) *Query {
	// queries cannot be anonymous
	q := &Query{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.query.%s", mod.ShortName, shortName),
		UnqualifiedName: fmt.Sprintf("query.%s", shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}
	return q
}

func QueryFromFile(modPath, filePath string, mod *Mod) (MappableResource, []byte, error) {
	q := &Query{
		Mod: mod,
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
		typehelpers.SafeString(q.SearchPath) == typehelpers.SafeString(other.SearchPath) &&
		typehelpers.SafeString(q.SearchPathPrefix) == typehelpers.SafeString(other.SearchPathPrefix) &&
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

func (q *Query) CtyValue() (cty.Value, error) {
	return getCtyValue(q)
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

// Name implements MappableResource, HclResource
func (q *Query) Name() string {
	return q.FullName
}

// GetUnqualifiedName implements DashboardLeafNode, ModTreeItem
func (q *Query) GetUnqualifiedName() string {
	return q.UnqualifiedName
}

// OnDecoded implements HclResource
func (q *Query) OnDecoded(*hcl.Block) hcl.Diagnostics { return nil }

// AddReference implements HclResource
func (q *Query) AddReference(ref *ResourceReference) {
	q.References = append(q.References, ref)
}

// GetReferences implements HclResource
func (q *Query) GetReferences() []*ResourceReference {
	return q.References
}

// GetMod implements HclResource
func (q *Query) GetMod() *Mod {
	return q.Mod
}

// GetDeclRange implements HclResource
func (q *Query) GetDeclRange() *hcl.Range {
	return &q.DeclRange
}

// GetParams implements QueryProvider
func (q *Query) GetParams() []*ParamDef {
	return q.Params
}

// GetArgs implements QueryProvider
func (q *Query) GetArgs() *QueryArgs {
	return nil
}

// GetQuery implements QueryProvider
func (q *Query) GetQuery() *Query {
	return nil
}

// GetSQL implements QueryProvider
func (q *Query) GetSQL() *string {
	return q.SQL
}

// SetArgs implements QueryProvider
func (q *Query) SetArgs(args *QueryArgs) {
	// nothing
}

// SetParams implements QueryProvider
func (q *Query) SetParams(params []*ParamDef) {
	q.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (q *Query) GetPreparedStatementName() string {
	if q.PreparedStatementName != "" {
		return q.PreparedStatementName
	}
	q.PreparedStatementName = q.buildPreparedStatementName(q.ShortName, q.Mod.NameWithVersion(), constants.PreparedStatementQuerySuffix)
	return q.PreparedStatementName
}

// GetPreparedStatementExecuteSQL implements QueryProvider
func (q *Query) GetPreparedStatementExecuteSQL(args *QueryArgs) (string, error) {
	// defer to base
	return q.getPreparedStatementExecuteSQL(q, args)
}

// AddParent implements ModTreeItem
func (q *Query) AddParent(parent ModTreeItem) error {
	q.parents = append(q.parents, parent)

	return nil
}

// GetParents implements ModTreeItem
func (q *Query) GetParents() []ModTreeItem {
	return q.parents
}

// GetChildren implements ModTreeItem
func (q *Query) GetChildren() []ModTreeItem {
	return nil
}

// GetDescription implements ModTreeItem
func (q *Query) GetDescription() string {
	return ""
}

// GetTitle implements ModTreeItem
func (q *Query) GetTitle() string {
	return typehelpers.SafeString(q.Title)
}

// GetTags implements ModTreeItem
func (q *Query) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (q *Query) GetPaths() []NodePath {
	// lazy load
	if len(q.Paths) == 0 {
		q.SetPaths()
	}
	return q.Paths
}

// SetPaths implements ModTreeItem
func (q *Query) SetPaths() {
	for _, parent := range q.parents {
		for _, parentPath := range parent.GetPaths() {
			q.Paths = append(q.Paths, append(parentPath, q.Name()))
		}
	}
}

func (q *Query) Diff(other *Query) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: q,
		Name: q.Name(),
	}

	if !utils.SafeStringsEqual(q.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(q.SQL, other.SQL) {
		res.AddPropertyDiff("SQL")
	}

	if !utils.SafeStringsEqual(q.SearchPath, other.SearchPath) {
		res.AddPropertyDiff("SearchPath")
	}

	res.populateChildDiffs(q, other)
	return res
}
