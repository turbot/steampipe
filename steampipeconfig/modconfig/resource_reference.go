package modconfig

import (
	"fmt"
	"strings"
)

type ResourceReference struct {
	Ref          string `cty:"ref" column:"ref,text"`
	ReferencedBy string `cty:"referenced_by" column:"referenced_by,text"`
	BlockType    string `cty:"block_type" column:"block_type,text"`
	BlockName    string `cty:"block_name" column:"block_name,text"`
	Attribute    string `cty:"attribute" column:"attribute,text"`
	metadata     *ResourceMetadata
}

// ResourceReferenceMap is a map of references keyed by 'ref'
// This is to handle the same reference being made more than once by a resource
// for example the reference var.v1 might be referenced several times
type ResourceReferenceMap map[string][]*ResourceReference

func (m ResourceReferenceMap) Add(reference *ResourceReference) {
	refs, ok := m[reference.Ref]
	if !ok {
		// if no ref instances, create an empty array
		refs = []*ResourceReference{}
	}
	// write back the updated array
	m[reference.Ref] = append(refs, reference)
}

func (r *ResourceReference) String() string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", r.ReferencedBy, r.BlockType, r.BlockName, r.Attribute, r.Ref)
}

func (r *ResourceReference) Equals(other *ResourceReference) bool {
	return r.String() == other.String()
}

// Name implements ResourceWithMetadata
func (r *ResourceReference) Name() string {
	return fmt.Sprintf("ref.%s", strings.Replace(r.String(), ".", "_", -1))
}

// GetMetadata implements ResourceWithMetadata
func (r *ResourceReference) GetMetadata() *ResourceMetadata {
	return r.metadata
}

// SetMetadata implements ResourceWithMetadata
func (r *ResourceReference) SetMetadata(metadata *ResourceMetadata) {
	r.metadata = metadata
}
