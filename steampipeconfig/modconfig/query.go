package modconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/turbot/go-kit/types"

	"github.com/turbot/steampipe/constants"
)

type Query struct {
	ShortName *string

	Description      *string   `hcl:"description" column:"description" column_type:"text"`
	Documentation    *string   `hcl:"documentation" column:"documentation" column_type:"text"`
	Labels           *[]string `hcl:"labels" column:"labels" column_type:"jsonb"`
	SQL              *string   `hcl:"sql" column:"sql" column_type:"text"`
	SearchPath       *string   `hcl:"search_path" column:"search_path" column_type:"text"`
	SearchPathPrefix *string   `hcl:"search_path_prefix" column:"search_path_prefix" column_type:"text"`
	Title            *string   `hcl:"title" column:"title" column_type:"text"`

	// resource metadata
	Metadata *ResourceMetadata
}

func (q *Query) String() string {
	return fmt.Sprintf(`
  -----
  Name: %s
  Title: %s
  Description: %s
  SQL: %s
`, types.SafeString(q.ShortName), types.SafeString(q.Title), types.SafeString(q.Description), types.SafeString(q.SQL))
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
	q.ShortName = &name
	q.SQL = &sql
	return q, sqlBytes, nil
}

// Name :: implementation of MappableResource
func (q *Query) Name() string {
	return fmt.Sprintf("query.%s", types.SafeString(q.ShortName))
}

// LongName :: name in format: '<modName>.control.<shortName>'
func (q *Query) LongName() string {
	return fmt.Sprintf("%s.%s", q.Metadata.ModShortName, q.Name())
}

// SetMetadata :: implementation of MappableResource
func (q *Query) SetMetadata(metadata *ResourceMetadata) {
	q.Metadata = metadata
}

// GetMetadata :: implementation of ResourceWithMetadata
func (q *Query) GetMetadata() *ResourceMetadata {
	return q.Metadata
}
