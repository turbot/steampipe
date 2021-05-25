package modconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/go-kit/types"

	"github.com/turbot/steampipe/constants"
)

// Query is a struct representing the Query resource
type Query struct {
	ShortName string
	FullName  string `cty:"name"`

	Description      *string            `cty:"description" hcl:"description" column:"description,text"`
	Documentation    *string            `cty:"documentation" hcl:"documentation" column:"documentation,text"`
	Tags             *map[string]string `cty:"tags" hcl:"tags" column:"tags,jsonb"`
	SQL              *string            `cty:"sql" hcl:"sql" column:"sql,text"`
	SearchPath       *string            `cty:"search_path" hcl:"search_path" column:"search_path,text"`
	SearchPathPrefix *string            `cty:"search_path_prefix" hcl:"search_path_prefix" column:"search_path_prefix,text"`
	Title            *string            `cty:"title" hcl:"title" column:"title,text"`

	// list of all block referenced by the resource
	References []string `column:"refs,jsonb"`

	DeclRange hcl.Range
	metadata  *ResourceMetadata
}

func NewQuery(block *hcl.Block) *Query {
	return &Query{
		ShortName: block.Labels[0],
		FullName:  fmt.Sprintf("query.%s", block.Labels[0]),
		DeclRange: block.DefRange,
	}
}

func (q *Query) CtyValue() (cty.Value, error) {
	return getCtyValue(q)
}

func (q *Query) String() string {
	return fmt.Sprintf(`
  -----
  Name: %s
  Title: %s
  Description: %s
  SQL: %s
`, q.FullName, types.SafeString(q.Title), types.SafeString(q.Description), types.SafeString(q.SQL))
}

// QueryFromFile :: factory function
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

	sqlBytes, err := ioutil.ReadFile(filePath)
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
	q.FullName = fmt.Sprintf("query.%s", name)
	q.SQL = &sql
	return q, sqlBytes, nil
}

// Name implements MappableResource, HclResource
func (q *Query) Name() string {
	return q.FullName
}

// QualifiedName returns the name in format: '<modName>.control.<shortName>'
func (q *Query) QualifiedName() string {
	return fmt.Sprintf("%s.%s", q.metadata.ModShortName, q.FullName)
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
func (q *Query) OnDecoded(*hcl.Block) {}

// AddReference implements HclResource
func (q *Query) AddReference(reference string) {
	q.References = append(q.References, reference)
}
