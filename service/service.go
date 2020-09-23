package service

type ModelsService struct {
	metadata      ModelsMetadata
	servingConfig ModelsConfig
	servingReload ModelsReload
	storage       ModelStorage
}

func (s *ModelsService) archivePrefix() string {
	return "model"
}

// NewModelsService returns new instance of ModelsService
func NewModelsService(meta ModelsMetadata, servingConfig ModelsConfig, servingReload ModelsReload, storage ModelStorage) *ModelsService {
	return &ModelsService{
		metadata:      meta,
		servingConfig: servingConfig,
		servingReload: servingReload,
		storage:       storage,
	}
}

type ModulesService struct {
	metadata ModulesMetadata
	storage  ModuleStorage
}

func (s *ModulesService) archivePrefix() string {
	return "module"
}

// NewModulesService returns new instance of ModulesService
func NewModulesService(meta ModulesMetadata, storage ModuleStorage) *ModulesService {
	return &ModulesService{
		metadata: meta,
		storage:  storage,
	}
}
