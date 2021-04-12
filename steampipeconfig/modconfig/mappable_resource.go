package modconfig

// MappableResource :: a mod resource which can be created directly from a content file (e.g. sql, markdown)
type MappableResource interface {
	// initialise the mod resource from the file of the given path
	InitialiseFromFile(modPath, filePath string) (MappableResource, error)
}
