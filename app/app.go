package app

import (
	"fmt"
	"strings"
	"time"
)

const (
	StableLabel     = "stable"
	PrevStableLabel = "last_stable"

	StatusRunning = "running"
	StatusReady   = "ready"
	StatusPending = "pending"

	RequestFieldStatus = "status"

	// MaxTeamLength model team max length
	MaxTeamLength = 32
	// MaxProjectLength model project max length
	MaxProjectLength = 32
	// MaxModelNameLength model name max length
	MaxModelNameLength = 32

	ComponentServing   = "SERVING"
	ComponentStorage   = "STORAGE"
	ComponentLock      = "LOCK"
	ComponentDiscovery = "DISCOVERY"
	ComponentMetadata  = "METADATA"
	ComponentService   = "SERVICE"
	ComponentRest      = "REST"
	ComponentAPP       = "APP"
)

type ServableID struct {
	Team    string `json:"team"`
	Project string `json:"project"`
	Name    string `json:"name"`
}

func (s ServableID) InstanceName() string {
	return fmt.Sprintf("tfs-%s-%s", s.Team, s.Project)
}

func (s ServableID) InstanceHost(suffix string) string {
	if suffix == "" {
		return s.InstanceName()
	}

	if !strings.HasPrefix(suffix, ".") {
		return s.InstanceName() + "." + suffix
	}

	return s.InstanceName() + suffix
}

func (s ServableID) ArchiveName(prefix string, version int64) string {
	return fmt.Sprintf("%s_%s-%s-%s-%d_%d.tar", prefix, s.Team, s.Project, s.Name, version, time.Now().Unix())
}

// ModelID is struct to simply hold basic informations
// using in most application endpoints and functions
type ModelID struct {
	ServableID
	Version int64  `json:"version"`
	Label   string `json:"label,omitempty"`
}

// IsVersionSet checks if version is set
func (m ModelID) IsVersionSet() bool {
	if m.Version < 0 {
		return false
	}

	return true
}

// ModelData
type ModelData struct {
	ModelID
	ID      int64  `json:"id"`
	Status  string `json:"status"`
	Created string `json:"created,omitempty"`
	Updated string `json:"updated,omitempty"`
}

// ModuleID is struct to simply hold basic informations
// using in most application endpoints and functions
type ModuleID struct {
	ServableID
	Version int64 `json:"version"`
}

// IsVersionSet checks if version is set
func (m ModuleID) IsVersionSet() bool {
	if m.Version < 0 {
		return false
	}

	return true
}

// ModuleData
type ModuleData struct {
	ModuleID
	ID      int64  `json:"id"`
	Created string `json:"created,omitempty"`
	Updated string `json:"updated,omitempty"`
}

// QueryParameters
type QueryParameters map[string]interface{}

type AnswerStatusInfo struct {
	ResponseStatus
	ErrorToLog string
	Endpoint   string
}

type ResponseStatus struct {
	Code              int       `json:"code"`
	Status            string    `json:"status"`
	Error             string    `json:"error"`
	InstanceErrorList *[]string `json:"errorinstancelist,omitempty"`
}

type Response struct {
	ResponseStatus
	ResponseData interface{} `json:"output"`
}

type ResponseModel struct {
	Model   string `json:"model"`
	Label   string `json:"label"`
	Version int64  `json:"version"`
}

func (r ResponseModel) setModel(model ServableID) string {
	return fmt.Sprintf("%s/%s/%s", model.Team, model.Project, model.Name)
}

type ResponseModule struct {
	Module  string `json:"module"`
	Version int64  `json:"version"`
}

func (r ResponseModule) setModule(module ServableID) string {
	return fmt.Sprintf("%s/%s/%s", module.Team, module.Project, module.Name)
}

type ServableInstances struct {
	ServableID
	Instances []string
}

type ReloadResponse struct {
}

type LabelChanged struct {
	ServableID
	Label string

	PreviousVersion int64
	NewVersion      int64
}

type Archive struct {
	Data []byte
	Name string
}

type ErrorDetails struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type ErrorBody struct {
	Error ErrorDetails `json:"error_details"`
}
