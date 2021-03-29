package modconfig

// map of file extension to factory function to create
type factoryFunc func(path string) (MappableResource, error)

var ResourceTypeMap = map[string]factoryFunc{
	".sql": func(path string) (MappableResource, error) { return QueryFromFile(path) },
}

func RegisteredFileExtensions() []string {
	var res []string
	for ext := range ResourceTypeMap {
		res = append(res, ext)
	}
	return res
}
