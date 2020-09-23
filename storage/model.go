package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
)

var (
	directoryLayoutRegexp = []string{
		`^model_archive.tar$`,
		`^variables$`,
		`^variables/variables\.data-[0-9]{5}-of-[0-9]{5}$`,
		`^variables/variables\.index$`,
		`^saved_model\..*$`,
		`^README\.md$`}

	logStorageConfigDoesNotExistCode          = 1001
	logStorageInvalidDirectoryLayoutModelCode = 1003

	errInvalidDirectoryLayout = errors.New("directory layout is invalid")
	// ErrConfigDoesNotExist means that config file is not available on our storage
	ErrConfigDoesNotExist = exterr.NewErrorWithMessage("config does not exist").WithComponent(app.ComponentStorage).WithCode(logStorageConfigDoesNotExistCode)
)

// ModelsStorage represents all interfaces used while read/write models to a storage
type ModelsStorage struct {
	reader   ModelReader
	writer   ModelWriter
	remover  ModelRemover
	archiver Archiver
}

// NewModelsStorage returns new instance of ModelsStorage
func NewModelsStorage(storageImplementation ModelStorage) *ModelsStorage {
	return &ModelsStorage{reader: storageImplementation, writer: storageImplementation, remover: storageImplementation, archiver: storageImplementation}
}

// ModelReader contains all read operations required by storage
type ModelReader interface {
	ReadConfig(ctx context.Context, team, project string) ([]byte, error)
	ReadModel(ctx context.Context, modelID app.ServableID, version int) ([]ArchiveHeader, error)
	ReadAllModels(ctx context.Context, modelID app.ServableID) ([]ArchiveHeader, error)
	DirectoryLayout(path string) ([]string, error)
}

// ModelWriter contains all write operations required by storage
type ModelWriter interface {
	SaveConfig(ctx context.Context, team, project string, config []byte) error
	SaveModel(ctx context.Context, archivePath string, modelID app.ServableID, version int) error
	SaveIncomingModelArchive(modelID app.ServableID, archive io.Reader) (string, error)
}

// ModelRemover contains all remove operations required by storage
type ModelRemover interface {
	RemoveModel(ctx context.Context, id app.ServableID, version int) error
}

// ReadModel gets archived bytes stream of model for given ServableID and model version
func (m *ModelsStorage) ReadModel(ctx context.Context, modelID app.ServableID, version int) ([]byte, error) {
	headers, err := m.reader.ReadModel(ctx, modelID, version)
	if err != nil {
		return nil, err
	}

	return createArchive(ctx, headers, m.archiver)
}

// ReadAllModels  gets archived bytes stream of model for given ServableID
// inside archive we got all versions of model
func (m *ModelsStorage) ReadAllModels(ctx context.Context, modelID app.ServableID) ([]byte, error) {
	headers, err := m.reader.ReadAllModels(ctx, modelID)
	if err != nil {
		return nil, err
	}

	return createArchive(ctx, headers, m.archiver)
}

// ReadConfig gets bytes stream of config file depends on given teamm and project
func (m *ModelsStorage) ReadConfig(ctx context.Context, team, project string) ([]byte, error) {
	return m.reader.ReadConfig(ctx, team, project)
}

type SaveModelResponse struct {
	Config []byte
}

func (m *ModelsStorage) SaveModel(ctx context.Context, modelID app.ServableID, version int, archive io.Reader) (*SaveModelResponse, error) {
	archiveID, err := m.writer.SaveIncomingModelArchive(modelID, archive)
	if err != nil {
		return nil, err
	}

	baseArchiveIDPath, _ := filepath.Split(archiveID)

	if err := extractArchive(ctx, archiveID, m.archiver); err != nil {
		removeAllErr := os.RemoveAll(baseArchiveIDPath)
		if removeAllErr != nil {
			return nil, exterr.WrapWithErr(err, removeAllErr)
		}
		return nil, err
	}

	directoryLayout, err := m.reader.DirectoryLayout(baseArchiveIDPath)
	if err != nil {
		removeAllErr := os.RemoveAll(baseArchiveIDPath)
		if removeAllErr != nil {
			return nil, exterr.WrapWithErr(err, removeAllErr)
		}
		return nil, err
	}

	if !isDirectoryLayoutValid(directoryLayout, directoryLayoutRegexp) {
		err := exterr.NewErrorWithErr(errInvalidDirectoryLayout).WithComponent(app.ComponentStorage).WithCode(logStorageInvalidDirectoryLayoutModelCode)
		removeAllErr := os.RemoveAll(baseArchiveIDPath)
		if removeAllErr != nil {
			return nil, exterr.WrapWithErr(err, removeAllErr)
		}
		return nil, err
	}

	if err := m.writer.SaveModel(ctx, archiveID, modelID, version); err != nil {
		removeAllErr := os.RemoveAll(baseArchiveIDPath)
		if removeAllErr != nil {
			return nil, exterr.WrapWithErr(err, removeAllErr)
		}
		return nil, err
	}

	var response SaveModelResponse
	config, err := m.ReadConfig(ctx, modelID.Team, modelID.Project)
	if err != nil && !errors.Is(err, ErrConfigDoesNotExist) {
		return nil, err
	}

	response.Config = config

	return &response, nil
}

// SaveConfig saves given config under valid location based on team and project parammeters
func (m *ModelsStorage) SaveConfig(ctx context.Context, team, project string, config []byte) error {
	return m.writer.SaveConfig(ctx, team, project, config)
}

// RemoveModel removes a model based on given ServableID and model version
func (m *ModelsStorage) RemoveModel(ctx context.Context, id app.ServableID, version int64) error {
	return m.remover.RemoveModel(ctx, id, int(version))
}
