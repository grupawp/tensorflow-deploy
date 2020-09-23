package filesystem

import (
	"archive/tar"
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/storage"
)

const emptyConfigContent = "model_config_list: {}"

// FSStorage holds filesystem configuration
type FSStorage struct {
	modelConf  *ModelFilesystemConfig
	moduleConf *ModuleFilesystemConfig
}

// NewStorager returs new instance if FilesystemStorage
// or error when soething goes wrong
func NewStorager(storageFilesystemConfig *app.ConfigStorageFilesystem) (*FSStorage, error) {

	modelDirPerm, err := storageFilesystemConfig.ConvertPerms(*storageFilesystemConfig.Model.DirectoryPermissions)
	if err != nil {
		return nil, err
	}
	modelFilePerm, err := storageFilesystemConfig.ConvertPerms(*storageFilesystemConfig.Model.FilePermissions)
	if err != nil {
		return nil, err
	}
	moduleDirPerm, err := storageFilesystemConfig.ConvertPerms(*storageFilesystemConfig.Module.DirectoryPermissions)
	if err != nil {

		return nil, err
	}
	moduleFilePerm, err := storageFilesystemConfig.ConvertPerms(*storageFilesystemConfig.Module.FilePermissions)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(*storageFilesystemConfig.Model.BasePath, modelDirPerm); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	if err := os.MkdirAll(*storageFilesystemConfig.Model.IncomingArchivePath, modelDirPerm); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	emptyModelsConfigPath := path.Join(*storageFilesystemConfig.Model.BasePath, *storageFilesystemConfig.Model.EmptyConfigName)
	if _, err := os.Stat(emptyModelsConfigPath); os.IsNotExist(err) {
		f, err := os.OpenFile(emptyModelsConfigPath, os.O_RDWR|os.O_CREATE, modelFilePerm)
		if err != nil {
			return nil, exterr.WrapWithFrame(err)
		}
		if _, err := f.Write([]byte(emptyConfigContent)); err != nil {
			return nil, exterr.WrapWithFrame(err)
		}
		f.Close()
	}

	if err := os.MkdirAll(*storageFilesystemConfig.Module.BasePath, modelDirPerm); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	if err := os.MkdirAll(*storageFilesystemConfig.Module.IncomingArchivePath, modelDirPerm); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	return &FSStorage{
		modelConf: &ModelFilesystemConfig{
			ArchiveName:         *storageFilesystemConfig.Model.ArchiveName,
			BasePath:            *storageFilesystemConfig.Model.BasePath,
			ConfigName:          *storageFilesystemConfig.Model.ConfigName,
			EmptyConfigName:     *storageFilesystemConfig.Model.EmptyConfigName,
			IncomingArchivePath: *storageFilesystemConfig.Model.IncomingArchivePath,
			DirPerm:             modelDirPerm,
			FilePerm:            modelFilePerm,
		},
		moduleConf: &ModuleFilesystemConfig{
			ArchiveName:         *storageFilesystemConfig.Module.ArchiveName,
			BasePath:            *storageFilesystemConfig.Module.BasePath,
			IncomingArchivePath: *storageFilesystemConfig.Module.IncomingArchivePath,
			DirPerm:             moduleDirPerm,
			FilePerm:            moduleFilePerm,
		},
	}, nil

}

// GetFileContent returns bytes stream of file located under given source filepath
func (fs *FSStorage) GetFileContent(ctx context.Context, source string) ([]byte, error) {
	result, err := ioutil.ReadFile(source)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}
	return result, err
}

func getArchiveHeaders(ctx context.Context, sourcePath string) ([]storage.ArchiveHeader, error) {
	var headers []storage.ArchiveHeader

	err := filepath.Walk(sourcePath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		hdr.Name = "./" + strings.TrimPrefix(strings.TrimPrefix(currentPath, sourcePath), "/")

		headers = append(headers, storage.ArchiveHeader{
			Header:      hdr,
			ContentPath: currentPath,
		})

		return nil
	})
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	return headers, nil
}
