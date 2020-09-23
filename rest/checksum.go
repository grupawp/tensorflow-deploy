package rest

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/grupawp/tensorflow-deploy/logging"
)

const (
	fieldChecksumIsEmpty = "field checksum is empty"
)

func calculateChecksum(r io.Reader) string {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func isChecksumValid(ctx context.Context, r io.Reader, formChecksum string) bool {
	if formChecksum == "" {
		logging.Warn(ctx, fieldChecksumIsEmpty)
		return true
	}

	if formChecksum != calculateChecksum(r) {
		return false
	}

	return true
}
