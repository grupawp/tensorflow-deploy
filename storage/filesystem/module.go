package filesystem

import (
	"context"
	"fmt"
	"io"
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
	logStorageModuleAlreadyExistsCode = 1004
	errModuleAlreadyExists            = exterr.NewErrorWithMessage("module already exists").WithComponent(app.ComponentStorage).WithCode(logStorageModuleAlreadyExistsCode)
)

type ModuleFilesystemConfig struct {
	ArchiveName         string
	BasePath            string
	IncomingArchivePath string

	DirPerm  os.FileMode
	FilePerm os.FileMode
}

func (fs *FSStorage) ReadModule(ctx context.Context, moduleID app.ServableID, version int) ([]storage.ArchiveHeader, error) {
	sourcePath := path.Join(fs.moduleConf.BasePath, moduleID.Team, moduleID.Project, moduleID.Name, strconv.Itoa(version))

	return getArchiveHeaders(ctx, sourcePath)
}

func (fs *FSStorage) SaveModule(ctx context.Context, archivePath string, moduleID app.ServableID, version int) error {
	destinationPath := path.Join(fs.moduleConf.BasePath, moduleID.Team, moduleID.Project, moduleID.Name)
	if err := os.MkdirAll(destinationPath, fs.moduleConf.DirPerm); err != nil {
		return exterr.WrapWithFrame(err)
	}

	destinationPath = path.Join(destinationPath, strconv.Itoa(version))
	_, err := os.Stat(destinationPath)
	if err == nil {
		return errModuleAlreadyExists
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

	archivePath = path.Join(destinationPath, fs.moduleConf.ArchiveName)
	if err := os.Remove(archivePath); err != nil {
		return exterr.WrapWithFrame(err)
	}

	return nil
}

func (fs *FSStorage) SaveIncomingModuleArchive(moduleID app.ServableID, archive io.Reader) (string, error) {
	incomingDir := fmt.Sprintf("%s-%s-%s-%d", moduleID.Team, moduleID.Project, moduleID.Name, time.Now().Unix())
	incomingDir = path.Join(fs.moduleConf.IncomingArchivePath, incomingDir)
	if err := os.Mkdir(incomingDir, fs.moduleConf.DirPerm); err != nil {
		return "", exterr.WrapWithFrame(err)
	}

	filePath := path.Join(incomingDir, fs.moduleConf.ArchiveName)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, fs.moduleConf.FilePerm)
	if err != nil {
		return "", exterr.WrapWithFrame(err)
	}
	defer f.Close()

	if _, err := io.Copy(f, archive); err != nil {
		return "", exterr.WrapWithFrame(err)
	}

	return filePath, nil
}

func (fs *FSStorage) RemoveModule(ctx context.Context, id app.ServableID, version int64) error {
	destinationPath := path.Join(fs.moduleConf.BasePath, id.Team, id.Project, id.Name, strconv.FormatInt(version, 10))
	err := os.RemoveAll(destinationPath)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	return nil
}
