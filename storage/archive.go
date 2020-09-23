package storage

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"

	"github.com/grupawp/tensorflow-deploy/exterr"
)

// Archiver contains methods required while creating/saving an archive
type Archiver interface {
	GetFileContent(ctx context.Context, source string) ([]byte, error)
	SaveArchiveFile(ctx context.Context, header *tar.Header, archive *tar.Reader, archivePath string) error
}

type ArchiveHeader struct {
	Header      *tar.Header
	ContentPath string
}

func createArchive(ctx context.Context, headers []ArchiveHeader, storage Archiver) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, header := range headers {
		if err := tw.WriteHeader(header.Header); err != nil {
			return nil, exterr.WrapWithFrame(err)
		}

		if header.Header.Typeflag != tar.TypeReg {
			continue
		}

		fileContent, err := storage.GetFileContent(ctx, header.ContentPath)
		if err != nil {
			return nil, err
		}

		if _, err := tw.Write(fileContent); err != nil {
			return nil, exterr.WrapWithFrame(err)
		}
	}

	if err := tw.Close(); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	return buf.Bytes(), nil
}

func extractArchive(ctx context.Context, archivePath string, storage Archiver) error {
	a, err := os.Open(archivePath)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}
	defer a.Close()
	tr := tar.NewReader(a)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return exterr.WrapWithFrame(err)
		}

		if err := storage.SaveArchiveFile(ctx, hdr, tr, archivePath); err != nil {
			return err
		}
	}

	return nil
}
