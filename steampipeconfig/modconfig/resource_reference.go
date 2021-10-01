package modconfig

type ResourceReference struct {
	Name      string `cty:"name" json:"name"`
	BlockType string `cty:"block_type" json:"block_type"`
	BlockName string `cty:"block_name" json:"block_name"`
	Attribute string `cty:"attribute" json:"attribute"`
}

// ResourceReferenceMap is a map of all the reference name to usages of that reference
// for example the reference var.v1 might be referenced serveral times byt a resource
type ResourceReferenceMap map[string][]ResourceReference

func (m ResourceReferenceMap) Add(reference ResourceReference) {
	refs, ok := m[reference.Name]
	if !ok {
		// if no ref instances, create an empty array
		refs = []ResourceReference{}
	}
	// write back the updated array
	m[reference.Name] = append(refs, reference)

}

//
//type ResourceReferenceList []ResourceReference
//
//
//func (l ResourceReferenceList)ContainsResource(resource HclResource)bool{
//	for _, ref := range l{
//		if ref.MatchesResource(resource){
//			return true
//		}
//	}
//	return false
//}
