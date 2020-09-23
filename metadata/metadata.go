package metadata

import (
	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
)

const (
	StartVersion   = 1
	InvalidVersion = 0
	InvalidStatus  = 0
	InvalidID      = 0

	StatusRunning = 1
	StatusReady   = 2
	StatusPending = 3
)

var (
	// statuses is a helper to map status from name to number
	statuses = map[string]uint8{
		app.StatusRunning: StatusRunning,
		app.StatusReady:   StatusReady,
		app.StatusPending: StatusPending,
	}

	invalidStatusErrorCode  = 1001
	invalidVersionErrorCode = 1002

	errorInvalidStatus  = exterr.NewErrorWithMessage("invalid status").WithComponent(app.ComponentMetadata).WithCode(invalidStatusErrorCode)
	errorInvalidVersion = exterr.NewErrorWithMessage("invalid version").WithComponent(app.ComponentMetadata).WithCode(invalidVersionErrorCode)
)

// StatusToName converts status numeric id to name
func StatusToName(id uint8) (string, error) {
	for k, v := range statuses {
		if v == id {
			return k, nil
		}
	}

	return "", errorInvalidStatus
}

// StatusToID converts status name to numeric id
func StatusToID(name string) (uint8, error) {
	for k, v := range statuses {
		if k == name {
			return v, nil
		}
	}

	return InvalidStatus, errorInvalidStatus
}

// NextVersion increments version
func NextVersion(version int64) (int64, error) {
	nextVersion := version + 1
	if nextVersion <= version {
		return InvalidVersion, errorInvalidVersion
	}

	return nextVersion, nil
}
