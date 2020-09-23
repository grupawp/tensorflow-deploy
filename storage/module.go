package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
)

var (
	moduleDirectoryLayoutRegexp = []string{
		`^module_archive.tar$`,
		`^variables$`,
		`^variables/variables\.data-[0-9]{5}-of-[0-9]{5}$`,
		`^variables/variables\.index$`,
		`^saved_model\..*$`,
		`^README\.md$`}

	logStorageInvalidDirectoryLayoutModuleCode = 1005
)

// ModulesStorage represents all interfaces used while read/write modules to a storage
type ModulesStorage struct {
	archiver Archiver
	reader   ModuleReader
	writer   ModuleWriter
	remover  ModuleRemover
}

// NewModuleStorage returns new instance of ModulesStorage
func NewModuleStorage(storageImplementation ModuleStorage) *ModulesStorage {
	return &ModulesStorage{reader: storageImplementation, writer: storageImplementation, remover: storageImplementation, archiver: storageImplementation}
}

// ModuleReader contains all read operations required by storage
type ModuleReader interface {
	ReadModule(ctx context.Context, moduleID app.ServableID, version int) ([]ArchiveHeader, error)
	DirectoryLayout(path string) ([]string, error)
}

// ReadModule gets archived bytes stream of module for given ServableID and module version
func (m *ModulesStorage) ReadModule(ctx context.Context, moduleID app.ServableID, version int) ([]byte, error) {
	headers, err := m.reader.ReadModule(ctx, moduleID, version)
	if err != nil {
		return nil, err
	}

	return createArchive(ctx, headers, m.archiver)
}

// ModuleWriter contains all write operations required by storage
type ModuleWriter interface {
	SaveModule(ctx context.Context, archivePath string, moduleID app.ServableID, version int) error
	SaveIncomingModuleArchive(moduleID app.ServableID, archive io.Reader) (string, error)
}

// ModuleRemover contains all remove operations required by storage
type ModuleRemover interface {
	RemoveModule(ctx context.Context, id app.ServableID, version int64) error
}

func (m *ModulesStorage) SaveModule(ctx context.Context, moduleID app.ServableID, version int, archive io.Reader) error {
	archiveID, err := m.writer.SaveIncomingModuleArchive(moduleID, archive)
	if err != nil {
		return err
	}

	baseArchiveIDPath, _ := filepath.Split(archiveID)
	if err := extractArchive(ctx, archiveID, m.archiver); err != nil {
		removeAllErr := os.RemoveAll(baseArchiveIDPath)
		if removeAllErr != nil {
			return exterr.WrapWithErr(err, removeAllErr)
		}
		return err
	}

	directoryLayout, err := m.reader.DirectoryLayout(baseArchiveIDPath)
	if err != nil {
		removeAllErr := os.RemoveAll(baseArchiveIDPath)
		if removeAllErr != nil {
			return exterr.WrapWithErr(err, removeAllErr)
		}
		return err
	}

	if !isDirectoryLayoutValid(directoryLayout, moduleDirectoryLayoutRegexp) {
		err = exterr.NewErrorWithErr(errInvalidDirectoryLayout).WithComponent(app.ComponentStorage).WithCode(logStorageInvalidDirectoryLayoutModuleCode)
		removeAllErr := os.RemoveAll(baseArchiveIDPath)
		if removeAllErr != nil {
			return exterr.WrapWithErr(err, removeAllErr)
		}
		return err
	}

	if err := m.writer.SaveModule(ctx, archiveID, moduleID, version); err != nil {
		removeAllErr := os.RemoveAll(baseArchiveIDPath)
		if removeAllErr != nil {
			return exterr.WrapWithErr(err, removeAllErr)
		}
		return err
	}

	return nil
}

// RemoveModule removes a module based on given ServableID and module version
func (m *ModulesStorage) RemoveModule(ctx context.Context, id app.ServableID, version int64) error {
	return m.remover.RemoveModule(ctx, id, version)
}
