package modconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/go-kit/types"

	"github.com/turbot/steampipe/constants"
)

type Query struct {
	ShortName string `json:"name"`

	Description      *string   `json:"description" hcl:"description" column:"description" column_type:"text"`
	Documentation    *string   `json:"documentation" hcl:"documentation" column:"documentation" column_type:"text"`
	Labels           *[]string `json:"labels" hcl:"labels" column:"labels" column_type:"jsonb"`
	SQL              *string   `json:"sql" hcl:"sql" column:"sql" column_type:"text"`
	SearchPath       *string   `json:"search_path" hcl:"search_path" column:"search_path" column_type:"text"`
	SearchPathPrefix *string   `json:"search_path_prefix" hcl:"search_path_prefix" column:"search_path_prefix" column_type:"text"`
	Title            *string   `json:"title" hcl:"title" column:"title" column_type:"text"`

	DeclRange hcl.Range `json:"-"`
	metadata  *ResourceMetadata
}

// Schema :: hcl schema for control
func (q *Query) Schema() *hcl.BodySchema {
	var attributes []hcl.AttributeSchema
	for attribute := range HclProperties(q) {
		attributes = append(attributes, hcl.AttributeSchema{Name: attribute})
	}
	return &hcl.BodySchema{Attributes: attributes}
}

func (q *Query) CtyValue() (cty.Value, error) {
	return getCtyValue(q, queryBlock)
}

// queryBlock :: return the block schema of a hydrated Query
// used to convert a query into a cty type for block evaluation
// TODO autogenerate from Query struct by reflection?
var queryBlock = configschema.Block{
	Attributes: map[string]*configschema.Attribute{
		"name":               {Optional: true, Type: cty.String},
		"description":        {Optional: true, Type: cty.String},
		"documentation":      {Optional: true, Type: cty.String},
		"labels":             {Optional: true, Type: cty.List(cty.String)},
		"sql":                {Optional: true, Type: cty.String},
		"search_path":        {Optional: true, Type: cty.String},
		"search_path_prefix": {Optional: true, Type: cty.String},
		"title":              {Optional: true, Type: cty.String},
	},
}

func queryCtyType() cty.Type {
	spec := queryBlock.DecoderSpec()
	return hcldec.ImpliedType(spec)
}
func (q *Query) String() string {
	return fmt.Sprintf(`
  -----
  Name: %s
  Title: %s
  Description: %s
  SQL: %s
`, q.ShortName, types.SafeString(q.Title), types.SafeString(q.Description), types.SafeString(q.SQL))
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
	q.SQL = &sql
	return q, sqlBytes, nil
}

// Name :: implementation of MappableResource, HclResource
func (q *Query) Name() string {
	return fmt.Sprintf("query.%s", q.ShortName)
}

// LongName :: name in format: '<modName>.control.<shortName>'
func (q *Query) LongName() string {
	return fmt.Sprintf("%s.%s", q.metadata.ModShortName, q.Name())
}

// GetMetadata :: implementation of HclResource and MappableResource
func (q *Query) GetMetadata() *ResourceMetadata {
	return q.metadata
}

// SetMetadata :: implementation of MappableResource, HclResource
func (q *Query) SetMetadata(metadata *ResourceMetadata) {
	q.metadata = metadata
}
