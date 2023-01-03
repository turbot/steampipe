package modconfig

import "github.com/hashicorp/hcl/v2"

type ResourceWithMetadataImpl struct {
	// required to allow partial decoding
	ResourceWithMetadataBaseRemain hcl.Body             `hcl:",remain" json:"-"`
	References                     []*ResourceReference `json:"-"`

	metadata  *ResourceMetadata
	anonymous bool
}

// GetMetadata implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) GetMetadata() *ResourceMetadata {
	return b.metadata
}

// SetMetadata implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) SetMetadata(metadata *ResourceMetadata) {
	b.metadata = metadata
	// set anonymous property on metadata
	b.metadata.Anonymous = b.anonymous
}

// SetAnonymous implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) SetAnonymous(block *hcl.Block) {
	b.anonymous = len(block.Labels) == 0
}

// IsAnonymous implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) IsAnonymous() bool {
	return b.anonymous
}

// AddReference implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) AddReference(ref *ResourceReference) {
	b.References = append(b.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) GetReferences() []*ResourceReference {
	return b.References
}
