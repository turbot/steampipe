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

type Query struct {
	ShortName string
	FullName  string `hcl:"name"`

	Description      *string   `hcl:"description" hcl:"description" column:"description" column_type:"text"`
	Documentation    *string   `hcl:"documentation" hcl:"documentation" column:"documentation" column_type:"text"`
	Labels           *[]string `hcl:"labels" hcl:"labels" column:"labels" column_type:"jsonb"`
	SQL              *string   `hcl:"sql" hcl:"sql" column:"sql" column_type:"text"`
	SearchPath       *string   `hcl:"search_path" hcl:"search_path" column:"search_path" column_type:"text"`
	SearchPathPrefix *string   `hcl:"search_path_prefix" hcl:"search_path_prefix" column:"search_path_prefix" column_type:"text"`
	Title            *string   `hcl:"title" hcl:"title" column:"title" column_type:"text"`

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

// Schema :: hcl schema for control
func (q *Query) Schema() *hcl.BodySchema {
	return buildAttributeSchema(q)
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

// InitialiseFromFile :: implementation of MappableResource
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

// Name :: implementation of MappableResource, HclResource
func (q *Query) Name() string {
	return q.FullName
}

// QualifiedName :: name in format: '<modName>.control.<shortName>'
func (q *Query) QualifiedName() string {
	return fmt.Sprintf("%s.%s", q.metadata.ModShortName, q.FullName)
}

// GetMetadata :: implementation of HclResource and MappableResource
func (q *Query) GetMetadata() *ResourceMetadata {
	return q.metadata
}

// SetMetadata :: implementation of MappableResource, HclResource
func (q *Query) SetMetadata(metadata *ResourceMetadata) {
	q.metadata = metadata
}
