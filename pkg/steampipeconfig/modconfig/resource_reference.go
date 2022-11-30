package modconfig

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"
)

type ResourceReference struct {
	ResourceWithMetadataBase

	To        string `cty:"reference_to" column:"reference_to,text"`
	From      string `cty:"reference_from" column:"reference_from,text"`
	BlockType string `cty:"from_block_type" column:"from_block_type,text"`
	BlockName string `cty:"from_block_name" column:"from_block_name,text"`
	Attribute string `cty:"from_attribute" column:"from_attribute,text"`
}

// ResourceReferenceMap is a map of references keyed by 'ref'
// This is to handle the same reference being made more than once by a resource
// for example the reference var.v1 might be referenced several times
type ResourceReferenceMap map[string][]*ResourceReference

func (m ResourceReferenceMap) Add(reference *ResourceReference) {
	refs, ok := m[reference.To]
	if !ok {
		// if no ref instances, create an empty array
		refs = []*ResourceReference{}
	}
	// write back the updated array
	m[reference.To] = append(refs, reference)
}

func (r *ResourceReference) String() string {
	return fmt.Sprintf("To: %s\nFrom: %s\nBlockType: %s\nBlockName: %s\nAttribute: %s",
		r.To,
		r.From,
		r.BlockType,
		r.BlockName,
		r.Attribute)
}

func (r *ResourceReference) Equals(other *ResourceReference) bool {
	return r.String() == other.String()
}

// Name implements ResourceWithMetadata
// the name must start with the 'resource type' as we parse it and use just the 'name' segment
func (r *ResourceReference) Name() string {
	hash := helpers.GetMD5Hash(r.String())[:8]
	str := fmt.Sprintf("ref.%s", hash)
	return str
}
