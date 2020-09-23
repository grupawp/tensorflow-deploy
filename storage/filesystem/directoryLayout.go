package filesystem

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/grupawp/tensorflow-deploy/exterr"
)

// DirectoryLayout returns model's directory structure
func (fs *FSStorage) DirectoryLayout(path string) ([]string, error) {
	var directoryLayout []string
	parentPath := path
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			directoryLayout = append(directoryLayout, strings.Replace(path, parentPath, "", 1))
			return nil
		})

	if err != nil {
		return directoryLayout, exterr.WrapWithFrame(err)
	}

	return directoryLayout, nil
}
