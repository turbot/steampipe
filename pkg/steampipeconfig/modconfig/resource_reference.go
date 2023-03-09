package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
)

type ResourceReference struct {
	ResourceWithMetadataImpl

	To        string `cty:"reference_to" column:"reference_to,text"`
	From      string `cty:"reference_from" column:"reference_from,text"`
	BlockType string `cty:"from_block_type" column:"from_block_type,text"`
	BlockName string `cty:"from_block_name" column:"from_block_name,text"`
	Attribute string `cty:"from_attribute" column:"from_attribute,text"`
	name      string
}

func NewResourceReference(resource HclResource, block *hcl.Block, referenceString string, blockName string, attr *hclsyntax.Attribute) *ResourceReference {
	ref := &ResourceReference{
		To:        referenceString,
		From:      resource.GetUnqualifiedName(),
		BlockType: block.Type,
		BlockName: blockName,
		Attribute: attr.Name,
	}
	ref.name = ref.buildName()
	return ref
}

func (r *ResourceReference) CloneWithNewFrom(from string) *ResourceReference {
	ref := &ResourceReference{
		ResourceWithMetadataImpl: r.ResourceWithMetadataImpl,
		To:                       r.To,
		From:                     from,
		BlockType:                r.BlockType,
		BlockName:                r.BlockName,
		Attribute:                r.Attribute,
	}
	ref.name = ref.buildName()
	// clone metadata so we can mutate it
	ref.ResourceWithMetadataImpl.metadata = ref.ResourceWithMetadataImpl.metadata.Clone()
	// set metadata name
	ref.ResourceWithMetadataImpl.metadata.ResourceName = ref.name
	return ref
}

func (r *ResourceReference) buildName() string {
	return helpers.GetMD5Hash(r.String())[:8]
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
	return fmt.Sprintf("ref.%s", r.name)
}
