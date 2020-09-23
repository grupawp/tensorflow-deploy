package filesystem

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/grupawp/tensorflow-deploy/exterr"
)

// SaveArchiveFile saves an archive under given archivePath location
func (fs *FSStorage) SaveArchiveFile(ctx context.Context, header *tar.Header, archive *tar.Reader, archivePath string) error {
	target := path.Join(path.Dir(archivePath), header.Name)

	switch header.Typeflag {
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(target), fs.modelConf.DirPerm); err != nil {
			return exterr.WrapWithFrame(err)
		}

		f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE, fs.modelConf.FilePerm)
		if err != nil {
			return exterr.WrapWithFrame(err)
		}

		if _, err := io.Copy(f, archive); err != nil {
			return exterr.WrapWithFrame(err)
		}

		f.Close()
	}

	return nil
}
