package modconfig

type ResourceWithMetadataBase struct {
	metadata *ResourceMetadata
}

// GetMetadata implements ResourceWithMetadata
func (b *ResourceWithMetadataBase) GetMetadata() *ResourceMetadata {
	return b.metadata
}

// SetMetadata implements ResourceWithMetadata
func (b *ResourceWithMetadataBase) SetMetadata(metadata *ResourceMetadata) {
	b.metadata = metadata
}

func (b *ResourceWithMetadataBase) IsAnonymous() bool {
	if b.metadata == nil {
		return false
	}
	return b.metadata.Anonymous
}
