package rest

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/logging"
)

var (
	logModuleBadRequestErrorCode = 1002

	errorModuleBadRequest = exterr.NewErrorWithMessage("bad request").WithComponent(app.ComponentRest).WithCode(logModuleBadRequestErrorCode)
)

// ModulesService is the interface that ...
type ModulesService interface {
	GetArchiveByVersion(ctx context.Context, id app.ServableID, version int64) (*app.Archive, error)
	ListModules(ctx context.Context, params app.QueryParameters) ([]*app.ModuleData, error)
	ListModulesByName(ctx context.Context, id app.ServableID) ([]*app.ModuleData, error)
	ListModulesByProject(ctx context.Context, team, project string) ([]*app.ModuleData, error)
	UploadModule(ctx context.Context, module app.ServableID, file io.Reader) (*app.ModuleID, error)
	RemoveByVersion(ctx context.Context, module app.ServableID, version int64) error
}

func (rest *REST) listModulesHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, true)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModuleBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	modules, err := rest.modulesService.ListModules(r.Context(), urlParams.QueryParameters())
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, modules)
}

func (rest *REST) listModulesByProjectHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModuleBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	modules, err := rest.modulesService.ListModulesByProject(r.Context(), urlParams.Team, urlParams.Project)
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, modules)
}

func (rest *REST) listModulesByNameHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModuleBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	modules, err := rest.modulesService.ListModulesByName(r.Context(), urlParams.ServableID())
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, modules)
}

func (rest *REST) deleteModuleHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName, urlVersion)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModuleBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := rest.modulesService.RemoveByVersion(r.Context(), urlParams.ServableID(), urlParams.Version); err != nil {
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, nil)
	return
}

// UploadModuleResponse holds such of information like
// moduleID or response code that are returned after
// uploading a module
type UploadModuleResponse struct {
	moduleID     *app.ModuleID
	responseCode int
}

func (rest *REST) uploadModuleHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModuleBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}
	if err := rest.lock.Lock(urlParams.ServableID()); err != nil {
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}
	defer rest.lock.UnLock(urlParams.ServableID())

	resp, err := rest.uploadModule(r, urlParams.ServableID())
	if err != nil {
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, resp.responseCode, err)
		return
	}

	writeJSONSuccessResponse(w, r, resp.responseCode, resp.moduleID)
}

func (rest *REST) uploadModule(r *http.Request, id app.ServableID) (*UploadModuleResponse, error) {
	file, _, err := r.FormFile(rest.uploadFileName)
	if err != nil {
		return &UploadModuleResponse{responseCode: http.StatusTemporaryRedirect}, exterr.WrapWithFrame(err)
	}
	defer file.Close()
	defer r.MultipartForm.RemoveAll()

	var dup bytes.Buffer
	tee := io.TeeReader(file, &dup)

	if checksum := r.FormValue(rest.uploadFileChecksum); len(checksum) > 0 {
		if !isChecksumValid(r.Context(), tee, checksum) {
			return &UploadModuleResponse{responseCode: http.StatusBadRequest}, errorInvalidChecksum
		}
		tee = bytes.NewReader(dup.Bytes())
	}

	module, err := rest.modulesService.UploadModule(r.Context(), id, tee)
	if err != nil {
		return &UploadModuleResponse{responseCode: http.StatusTemporaryRedirect}, err
	}

	return &UploadModuleResponse{moduleID: module, responseCode: http.StatusOK}, nil
}

func (rest *REST) downloadModuleByVersionHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName, urlVersion)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModuleBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	archive, err := rest.modulesService.GetArchiveByVersion(r.Context(), urlParams.ServableID(), urlParams.Version)
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}

	writeBinaryDataResponse(w, r, archive.Data, archive.Name)
}
