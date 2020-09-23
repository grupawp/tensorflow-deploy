package storage

// ModelStorage represents all interfaces used by storage
// while processing models
type ModelStorage interface {
	Archiver
	ModelReader
	ModelWriter
	ModelRemover
}

// ModuleStorage represents all interfaces used by storage
// while processing modules
type ModuleStorage interface {
	Archiver
	ModuleReader
	ModuleWriter
	ModuleRemover
}
