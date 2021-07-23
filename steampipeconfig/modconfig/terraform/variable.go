package terraform

//
//// Variable is a struct representing a Variable resource
//type Variable struct {
//	ShortName string
//	FullName  string `cty:"name"`
//
//	Type cty.S
//	Value     cty.Value
//	DeclRange hcl.Range
//
//	metadata *ResourceMetadata
//}
//
//func NewVariable(name string, val cty.Value, declRange hcl.Range) *Variable {
//	return &Variable{
//		ShortName: name,
//		FullName:  fmt.Sprintf("variable.%s", name),
//		Value:     val,
//		DeclRange: declRange,
//	}
//}
//
//// Name implements HclResource, ResourceWithMetadata
//func (l *Variable) Name() string {
//	return l.FullName
//}
//
//// GetMetadata implements ResourceWithMetadata
//func (l *Variable) GetMetadata() *ResourceMetadata {
//	return l.metadata
//}
//
//// SetMetadata implements ResourceWithMetadata
//func (l *Variable) SetMetadata(metadata *ResourceMetadata) {
//	l.metadata = metadata
//}
//
//// OnDecoded implements HclResource
//func (l *Variable) OnDecoded(*hcl.Block) {}
//
//// AddReference implements HclResource
//func (l *Variable) AddReference(string) {}
//
//// CtyValue implements HclResource
//func (l *Variable) CtyValue() (cty.Value, error) {
//	return l.Value, nil
//}
