package modconfig

// MappableResource :: a mod resource which can be created directly from a content file (e.g. sql, markdown)
type MappableResource interface {
	// InitialiseFromFile :: initialise the mod resource from the file of the given path
	// return the created resource, and the file data
	InitialiseFromFile(modPath, filePath string) (MappableResource, []byte, error)

	Name() string

	SetReflectionData(*CoreReflectionData)
}
