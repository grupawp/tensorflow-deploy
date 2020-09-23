package filesystem

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/storage"
)

var (
	logStorageModelAlreadyExistsCode = 1002
	errModelAlreadyExists            = exterr.NewErrorWithMessage("model already exists").WithComponent(app.ComponentStorage).WithCode(logStorageModelAlreadyExistsCode)
)

// ModelFilesystemConfig holds configuration necessary
// for valid work of filesystem
type ModelFilesystemConfig struct {
	ArchiveName         string
	BasePath            string
	ConfigName          string
	EmptyConfigName     string
	IncomingArchivePath string

	DirPerm  os.FileMode
	FilePerm os.FileMode
}

func (fs *FSStorage) ReadModel(ctx context.Context, modelID app.ServableID, version int) ([]storage.ArchiveHeader, error) {
	sourcePath := path.Join(fs.modelConf.BasePath, modelID.Team, modelID.Project, modelID.Name, strconv.Itoa(version))

	return getArchiveHeaders(ctx, sourcePath)
}

func (fs *FSStorage) ReadConfig(ctx context.Context, team, project string) ([]byte, error) {
	configPath := path.Join(fs.modelConf.BasePath, team, project, fs.modelConf.ConfigName)
	config, err := ioutil.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil, exterr.WrapWithErr(err, storage.ErrConfigDoesNotExist)
	}
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	return config, nil
}

func (fs *FSStorage) ReadAllModels(ctx context.Context, modelID app.ServableID) ([]storage.ArchiveHeader, error) {
	sourcePath := path.Join(fs.modelConf.BasePath, modelID.Team, modelID.Project, modelID.Name)

	return getArchiveHeaders(ctx, sourcePath)
}

func (fs *FSStorage) SaveConfig(ctx context.Context, team, project string, config []byte) error {
	sourcePath := path.Join(fs.modelConf.BasePath, team, project)
	if _, err := os.Stat(sourcePath); err != nil {
		return exterr.WrapWithFrame(err)
	}

	sourcePath = path.Join(sourcePath, fs.modelConf.ConfigName)

	f, err := os.OpenFile(sourcePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.modelConf.FilePerm)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}
	defer f.Close()

	if _, err := f.Write(config); err != nil {
		return exterr.WrapWithFrame(err)
	}

	return nil
}

func (fs *FSStorage) SaveModel(ctx context.Context, archivePath string, modelID app.ServableID, version int) error {
	destinationPath := path.Join(fs.modelConf.BasePath, modelID.Team, modelID.Project, modelID.Name)
	if err := os.MkdirAll(destinationPath, fs.modelConf.DirPerm); err != nil {
		return exterr.WrapWithFrame(err)
	}

	destinationPath = path.Join(destinationPath, strconv.Itoa(version))

	_, err := os.Stat(destinationPath)
	if err == nil {
		return errModelAlreadyExists
	}
	if !os.IsNotExist(err) {
		return exterr.WrapWithFrame(err)
	}

	if err := os.Rename(filepath.Dir(archivePath), destinationPath); err != nil {
		removeAllErr := os.RemoveAll(destinationPath)
		if removeAllErr != nil {
			return exterr.WrapWithErr(err, removeAllErr)
		}
		return exterr.WrapWithFrame(err)
	}

	archivePath = path.Join(destinationPath, fs.modelConf.ArchiveName)
	if err := os.Remove(archivePath); err != nil {
		return exterr.WrapWithFrame(err)
	}

	return nil
}

func (fs *FSStorage) SaveIncomingModelArchive(modelID app.ServableID, archive io.Reader) (string, error) {
	incomingDir := fmt.Sprintf("%s-%s-%s-%d", modelID.Team, modelID.Project, modelID.Name, time.Now().Unix())
	incomingDir = path.Join(fs.modelConf.IncomingArchivePath, incomingDir)
	if err := os.Mkdir(incomingDir, fs.modelConf.DirPerm); err != nil {
		return "", exterr.WrapWithFrame(err)
	}

	filePath := path.Join(incomingDir, fs.modelConf.ArchiveName)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, fs.modelConf.FilePerm)
	if err != nil {
		return "", exterr.WrapWithFrame(err)
	}
	defer f.Close()

	if _, err := io.Copy(f, archive); err != nil {
		return "", exterr.WrapWithFrame(err)
	}

	return filePath, nil
}

func (fs *FSStorage) RemoveModel(ctx context.Context, id app.ServableID, version int) error {
	destinationPath := path.Join(fs.modelConf.BasePath, id.Team, id.Project, id.Name, strconv.Itoa(version))
	err := os.RemoveAll(destinationPath)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}
	return nil
}
