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
	ShortName   *string
	Title       *string `hcl:"title"`
	Description *string `hcl:"description"`
	SQL         *string `hcl:"sql"`

	// reflection data
	ReflectionData *ReflectionData
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

//
//func (q *Query) Equals(other *Query) bool {
//	return types.SafeString(q.Name) == types.SafeString(other.Name) &&
//		types.SafeString(q.Title) == types.SafeString(other.Title) &&
//		types.SafeString(q.Description) == types.SafeString(other.Description) &&
//		types.SafeString(q.SQL) == types.SafeString(other.SQL)
//}

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

// SetReflectionData :: implementation of MappableResource
func (q *Query) SetReflectionData(reflectionData *ReflectionData) {
	q.ReflectionData = reflectionData
}
