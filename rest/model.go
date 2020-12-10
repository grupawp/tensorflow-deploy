package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/logging"
)

const labelStable = "stable"

var (
	logModelBadRequestErrorCode = 1001

	errorModelBadRequest = exterr.NewErrorWithMessage("bad request").WithComponent(app.ComponentRest).WithCode(logModelBadRequestErrorCode)
)

// UploadModelResponse holds such of information like
// modelID or response code that are returned after
// uploading a model
type UploadModelResponse struct {
	modelID      *app.ModelID
	responseCode int
}

// ModelsService is the interface that ...
type ModelsService interface {
	ArchiveByLabel(ctx context.Context, id app.ServableID, label string) (*app.Archive, error)
	ArchiveByVersion(ctx context.Context, id app.ServableID, version int64) (*app.Archive, error)
	GetConfigStream(ctx context.Context, team, project string) ([]byte, error)

	ListModels(ctx context.Context, params app.QueryParameters) ([]*app.ModelData, error)
	ListModelsByProject(ctx context.Context, team, project string) ([]*app.ModelData, error)
	ListModelsByName(ctx context.Context, id app.ServableID) ([]*app.ModelData, error)
	ReloadModels(ctx context.Context, team, project string, skipConfigWithoutLabels bool) ([]app.ReloadResponse, error)
	UploadModel(ctx context.Context, model app.ServableID, file io.Reader, label ...string) (*app.ModelID, error)

	RemoveByLabel(ctx context.Context, id app.ServableID, label string) error
	RemoveByVersion(ctx context.Context, id app.ServableID, version int64) error
	RemoveModelLabel(ctx context.Context, id app.ServableID, label string) error

	Revert(ctx context.Context, id app.ServableID) (*app.LabelChanged, error)
	SetLabel(ctx context.Context, model app.ModelID) (*app.LabelChanged, error)
}

func (rest *REST) listModelsHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, true)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	models, err := rest.modelsService.ListModels(r.Context(), urlParams.QueryParameters())
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, models)
}

func (rest *REST) listModelsByProjectHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	models, err := rest.modelsService.ListModelsByProject(r.Context(), urlParams.Team, urlParams.Project)
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, models)
}

func (rest *REST) listModelsByNameHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	id := app.ServableID{
		Team:    urlParams.Team,
		Project: urlParams.Project,
		Name:    urlParams.Name}

	models, err := rest.modelsService.ListModelsByName(r.Context(), id)
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, models)
}

func (rest *REST) revertModelHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	revertResp, err := rest.modelsService.Revert(r.Context(), urlParams.ServableID())
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, fmt.Sprintf("model[%s-%s-%s] label '%s' changed from version [%d] to [%d]",
		revertResp.Team, revertResp.Project, revertResp.Name, revertResp.Label, revertResp.PreviousVersion, revertResp.NewVersion))
}

func (rest *REST) configFileHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	config, err := rest.modelsService.GetConfigStream(r.Context(), urlParams.Team, urlParams.Project)
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	writeBinaryDataResponse(w, r, config, "models.config")
}

func (rest *REST) reloadHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	reloadStatus, err := rest.modelsService.ReloadModels(r.Context(), urlParams.Team, urlParams.Project, urlParams.SkipShortConfig)
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, reloadStatus)
}

func (rest *REST) downloadModelByLabelHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName, urlLabel)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	id := app.ServableID{Team: urlParams.Team, Project: urlParams.Project, Name: urlParams.Name}
	archive, err := rest.modelsService.ArchiveByLabel(r.Context(), id, urlParams.Label)
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}

	writeBinaryDataResponse(w, r, archive.Data, archive.Name)
}

func (rest *REST) deleteModelByLabelHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName, urlLabel)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := rest.modelsService.RemoveByLabel(r.Context(), urlParams.ServableID(), urlParams.Label); err != nil {
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, nil)
	return
}

func (rest *REST) deleteModelByVersionHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName, urlVersion)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := rest.modelsService.RemoveByVersion(r.Context(), urlParams.ServableID(), urlParams.Version); err != nil {
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, nil)
	return
}

func (rest *REST) deleteModelLabelHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName, urlLabel)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := rest.modelsService.RemoveModelLabel(r.Context(), urlParams.ServableID(), urlParams.Label); err != nil {
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, nil)
	return
}

func (rest *REST) downloadModelByVersionHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName, urlVersion)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	archive, err := rest.modelsService.ArchiveByVersion(r.Context(), urlParams.ServableID(), urlParams.Version)
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}

	writeBinaryDataResponse(w, r, archive.Data, archive.Name)
}

func (rest *REST) setModelLabelToStableHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName, urlVersion)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	modelID := app.ModelID{ServableID: urlParams.ServableID(), Version: urlParams.Version, Label: labelStable}
	lChangedResp, err := rest.modelsService.SetLabel(r.Context(), modelID)
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}

	if lChangedResp.PreviousVersion != 0 {
		writeJSONSuccessResponse(w, r, http.StatusOK, fmt.Sprintf("model[%s-%s-%s] label '%s' changed from version [%d] to [%d]",
			lChangedResp.Team, lChangedResp.Project, lChangedResp.Name, lChangedResp.Label, lChangedResp.PreviousVersion, lChangedResp.NewVersion))
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, fmt.Sprintf("model[%s-%s-%s] label '%s' version set to [%d]",
		lChangedResp.Team, lChangedResp.Project, lChangedResp.Name, lChangedResp.Label, lChangedResp.NewVersion))

}

func (rest *REST) setModelLabelHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName, urlVersion, urlLabel)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	modelID := app.ModelID{ServableID: urlParams.ServableID(), Version: urlParams.Version, Label: urlParams.Label}
	lChangedResp, err := rest.modelsService.SetLabel(r.Context(), modelID)
	if err != nil {
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}

	writeJSONSuccessResponse(w, r, http.StatusOK, fmt.Sprintf("model[%s-%s-%s] label '%s' changed from version [%d] to [%d]",
		lChangedResp.Team, lChangedResp.Project, lChangedResp.Name, lChangedResp.Label, lChangedResp.PreviousVersion, lChangedResp.NewVersion))
}

func (rest *REST) uploadModelHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	modelID := app.ServableID{Team: urlParams.Team, Project: urlParams.Project, Name: urlParams.Name}
	if err := rest.lock.Lock(modelID); err != nil {
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}
	defer rest.lock.UnLock(modelID)

	resp, err := rest.uploadModel(r, modelID)
	if err != nil {
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, resp.responseCode, err)
		return
	}
	writeJSONSuccessResponse(w, r, resp.responseCode, resp.modelID)
}

func (rest *REST) uploadModelWithLabelHandler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseAndValidateParamsFromRequest(r, false, urlTeam, urlProject, urlName, urlLabel)
	if err != nil {
		err = exterr.WrapWithErr(err, errorModelBadRequest)
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	modelID := app.ServableID{Team: urlParams.Team, Project: urlParams.Project, Name: urlParams.Name}
	if err := rest.lock.Lock(modelID); err != nil {
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, http.StatusTemporaryRedirect, err)
		return
	}
	defer rest.lock.UnLock(modelID)

	resp, err := rest.uploadModel(r, modelID, urlParams.Label)
	if err != nil {
		logging.ErrorWithStack(r.Context(), err)
		writeJSONErrorResponse(w, r, resp.responseCode, err)
		return
	}

	writeJSONSuccessResponse(w, r, resp.responseCode, resp.modelID)
}

func (rest *REST) uploadModel(r *http.Request, id app.ServableID, label ...string) (*UploadModelResponse, error) {
	file, _, err := r.FormFile(rest.uploadFileName)
	if err != nil {
		return &UploadModelResponse{responseCode: http.StatusTemporaryRedirect}, exterr.WrapWithFrame(err)
	}
	defer file.Close()
	defer r.MultipartForm.RemoveAll()

	var dup bytes.Buffer
	tee := io.TeeReader(file, &dup)

	if checksum := r.FormValue(rest.uploadFileChecksum); len(checksum) > 0 {
		if !isChecksumValid(r.Context(), tee, checksum) {
			return &UploadModelResponse{responseCode: http.StatusBadRequest}, errorInvalidChecksum
		}
		tee = bytes.NewReader(dup.Bytes())
	}

	model, err := rest.modelsService.UploadModel(r.Context(), id, tee, label...)
	if err != nil {
		return &UploadModelResponse{responseCode: http.StatusTemporaryRedirect}, exterr.WrapWithFrame(err)
	}

	return &UploadModelResponse{modelID: model, responseCode: http.StatusOK}, nil
}
