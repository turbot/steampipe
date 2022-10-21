package modconfig

import "github.com/hashicorp/hcl/v2"

type ResourceWithMetadataBase struct {
	metadata  *ResourceMetadata
	anonymous bool
}

// GetMetadata implements ResourceWithMetadata
func (b *ResourceWithMetadataBase) GetMetadata() *ResourceMetadata {
	return b.metadata
}

// SetMetadata implements ResourceWithMetadata
func (b *ResourceWithMetadataBase) SetMetadata(metadata *ResourceMetadata) {
	b.metadata = metadata
	// set anonymous property on metadata
	b.metadata.Anonymous = b.anonymous
}

// SetAnonymous implements ResourceWithMetadata
func (b *ResourceWithMetadataBase) SetAnonymous(block *hcl.Block) {
	b.anonymous = len(block.Labels) == 0
}

// IsAnonymous implements ResourceWithMetadata
func (b *ResourceWithMetadataBase) IsAnonymous() bool {
	return b.anonymous
}

func (b *ResourceWithMetadataBase) AddReference(ref *ResourceReference) {
}

func (b *ResourceWithMetadataBase) GetReferences() []*ResourceReference {
	return nil
}
