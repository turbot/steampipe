package modconfig

type ResourceReference struct {
	Ref          string `cty:"ref" json:"ref"`
	ReferencedBy string `cty:"referenced_by" json:"referenced_by"`
	BlockType    string `cty:"block_type" json:"block_type"`
	BlockName    string `cty:"block_name" json:"block_name"`
	Attribute    string `cty:"attribute" json:"attribute"`
}

// ResourceReferenceMap is a map of references keyed by 'ref'
// THis is to handle the same same reference being made more than once by a resource
// for example the reference var.v1 might be referenced several times
type ResourceReferenceMap map[string][]ResourceReference

func (m ResourceReferenceMap) Add(reference ResourceReference) {
	refs, ok := m[reference.Ref]
	if !ok {
		// if no ref instances, create an empty array
		refs = []ResourceReference{}
	}
	// write back the updated array
	m[reference.Ref] = append(refs, reference)

}
