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
	"github.com/zclconf/go-cty/cty"
)

type Base struct {
	Foo string `column:"foo,text"`
}

// Query is a struct representing the Query resource
type Query struct {
	Base
	ShortName string `cty:"short_name"`
	FullName  string `cty:"name"`

	Description      *string           `cty:"description" column:"description,text"`
	Documentation    *string           `cty:"documentation"  column:"documentation,text"`
	SearchPath       *string           `cty:"search_path"column:"search_path,text"`
	SearchPathPrefix *string           `cty:"search_path_prefix" column:"search_path_prefix,text"`
	SQL              *string           `cty:"sql" hcl:"sql" column:"sql,text"`
	Tags             map[string]string `cty:"tags" hcl:"tags" column:"tags,jsonb"`
	Title            *string           `cty:"title" hcl:"title" column:"title,text"`

	Params []*ParamDef `cty:"params" column:"params,jsonb"`
	// list of all blocks referenced by the resource
	References []*ResourceReference

	Mod                   *Mod `cty:"mod"`
	DeclRange             hcl.Range
	PreparedStatementName string `column:"prepared_statement_name,text"`
	metadata              *ResourceMetadata
	UnqualifiedName       string
}

func NewQuery(block *hcl.Block) *Query {
	q := &Query{
		ShortName:       block.Labels[0],
		UnqualifiedName: fmt.Sprintf("query.%s", block.Labels[0]),
		FullName:        fmt.Sprintf("query.%s", block.Labels[0]),
		DeclRange:       block.DefRange,
	}
	q.Base.Foo = "BAR2BASE"
	return q
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

func QueryFromFile(modPath, filePath string) (MappableResource, []byte, error) {
	q := &Query{}
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
	q.FullName = fmt.Sprintf("query.%s", name)
	q.SQL = &sql
	return q, sqlBytes, nil
}

// Name implements MappableResource, HclResource
func (q *Query) Name() string {
	return q.FullName
}

// GetMetadata implements ResourceWithMetadata
func (q *Query) GetMetadata() *ResourceMetadata {
	return q.metadata
}

// SetMetadata implements ResourceWithMetadata
func (q *Query) SetMetadata(metadata *ResourceMetadata) {
	q.metadata = metadata
}

// OnDecoded implements HclResource
func (q *Query) OnDecoded(*hcl.Block) hcl.Diagnostics { return nil }

// AddReference implements HclResource
func (q *Query) AddReference(ref *ResourceReference) {
	q.References = append(q.References, ref)
}

// SetMod implements HclResource
func (q *Query) SetMod(mod *Mod) {
	q.Mod = mod
	// add mod name to full name
	q.FullName = fmt.Sprintf("%s.%s", mod.ShortName, q.FullName)
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

// GetPreparedStatementName implements QueryProvider
func (q *Query) GetPreparedStatementName() string {
	// lazy load
	if q.PreparedStatementName == "" {
		q.PreparedStatementName = preparedStatementName(q)
	}
	return q.PreparedStatementName
}

// ModName implements QueryProvider
func (q *Query) ModName() string {
	return q.Mod.NameWithVersion()
}
