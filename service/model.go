package service

import (
	"context"
	"fmt"
	"io"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/logging"
)

var (
	modelNotFoundErrorCode           = 1001
	stableModelNotFoundErrorCode     = 1002
	prevStableModelNotFoundErrorCode = 1003

	errorModelNotFound           = exterr.NewErrorWithMessage("model not found").WithComponent(app.ComponentService).WithCode(modelNotFoundErrorCode)
	errorStableModelNotFound     = exterr.NewErrorWithMessage("model with label 'stable' not found").WithComponent(app.ComponentService).WithCode(stableModelNotFoundErrorCode)
	errorPrevStableModelNotFound = exterr.NewErrorWithMessage("model with label 'prev_stable' not found").WithComponent(app.ComponentService).WithCode(prevStableModelNotFoundErrorCode)
)

func cleanList(models []*app.ModelData) []*app.ModelData {
	labeledModel := make(map[string]bool, 0)
	mapID := func(m *app.ModelData) string {
		return fmt.Sprintf("%s-%s-%s-%d", m.Team, m.Project, m.Name, m.Version)
	}

	for _, model := range models {
		if model.Label == "" {
			continue
		}

		labeledModel[mapID(model)] = true
	}

	var cleanModelsList []*app.ModelData
	for _, model := range models {
		if model.Label == "" && labeledModel[mapID(model)] {
			continue
		}
		cleanModelsList = append(cleanModelsList, model)
	}

	return cleanModelsList
}

func (s *ModelsService) ListModels(ctx context.Context, params app.QueryParameters) ([]*app.ModelData, error) {
	models, err := s.metadata.List(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return cleanList(models), nil
}

func (s *ModelsService) ListModelsByProject(ctx context.Context, team, project string) ([]*app.ModelData, error) {
	params := app.QueryParameters{"team": team, "project": project}
	models, err := s.metadata.List(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return cleanList(models), nil
}

func (s *ModelsService) ListModelsByName(ctx context.Context, id app.ServableID) ([]*app.ModelData, error) {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name}
	models, err := s.metadata.List(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return cleanList(models), nil
}

func (s *ModelsService) GetConfigStream(ctx context.Context, team, project string) ([]byte, error) {
	config, err := s.servingConfig.ConfigFileStream(ctx, team, project)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return config, nil
}

func (s *ModelsService) ArchiveByLabel(ctx context.Context, id app.ServableID, label string) (*app.Archive, error) {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name, "label": label}
	modelMeta, err := s.metadata.Get(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	if modelMeta == nil {
		logging.ErrorWithStack(ctx, errorModelNotFound)
		return nil, errorModelNotFound
	}

	archive, err := s.storage.ReadModel(ctx, id, int(modelMeta.Version))
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return &app.Archive{Data: archive, Name: id.ArchiveName(s.archivePrefix(), modelMeta.Version)}, nil
}

func (s *ModelsService) ArchiveByVersion(ctx context.Context, id app.ServableID, version int64) (*app.Archive, error) {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name, "version": version}
	modelMeta, err := s.metadata.Get(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	if modelMeta == nil {
		logging.ErrorWithStack(ctx, errorModelNotFound)
		return nil, errorModelNotFound
	}

	archive, err := s.storage.ReadModel(ctx, id, int(version))
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return &app.Archive{Data: archive, Name: id.ArchiveName(s.archivePrefix(), version)}, nil
}

func (s *ModelsService) ReloadModels(ctx context.Context, team, project string, skipConfigWithoutLabels bool) ([]app.ReloadResponse, error) {
	reloadStatus, err := s.servingReload.ReloadConfig(ctx, team, project, skipConfigWithoutLabels)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return reloadStatus, nil
}

func (s *ModelsService) SetLabel(ctx context.Context, model app.ModelID) (*app.LabelChanged, error) {
	params := app.QueryParameters{"team": model.Team, "project": model.Project, "name": model.Name, "version": model.Version}
	modelMeta, err := s.metadata.Get(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}
	if modelMeta == nil {
		logging.ErrorWithStack(ctx, errorModelNotFound)
		return nil, errorModelNotFound
	}

	prevVersion, err := s.servingConfig.UpdateLabel(ctx, model)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	if err := s.metadata.ChangeLabel(ctx, app.ModelData{ModelID: model, Status: app.StatusReady}); err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	if model.Label == app.StableLabel && prevVersion != 0 {
		lastStableModel := app.ModelID{ServableID: model.ServableID, Version: prevVersion, Label: app.PrevStableLabel}
		if err := s.metadata.ChangeLabel(ctx, app.ModelData{ModelID: lastStableModel, Status: app.StatusReady}); err != nil {
			logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
			return nil, err
		}
	}

	return &app.LabelChanged{ServableID: model.ServableID, Label: model.Label, PreviousVersion: prevVersion, NewVersion: model.Version}, nil
}

func (s *ModelsService) Revert(ctx context.Context, id app.ServableID) (*app.LabelChanged, error) {

	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name, "label": app.StableLabel}
	currentStableMeta, err := s.metadata.Get(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}
	if currentStableMeta == nil {
		logging.ErrorWithStack(ctx, errorStableModelNotFound)
		return nil, errorStableModelNotFound
	}

	params["label"] = app.PrevStableLabel
	newStableMeta, err := s.metadata.Get(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}
	if newStableMeta == nil {
		logging.ErrorWithStack(ctx, errorPrevStableModelNotFound)
		return nil, errorPrevStableModelNotFound
	}

	model := app.ModelID{ServableID: id, Version: newStableMeta.Version, Label: app.StableLabel}
	if _, err := s.servingConfig.UpdateLabel(ctx, model); err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	if err := s.metadata.Delete(ctx, newStableMeta.ID); err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	if err := s.metadata.ChangeLabel(ctx, app.ModelData{ModelID: model, Status: app.StatusReady}); err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return &app.LabelChanged{ServableID: model.ServableID, Label: model.Label, PreviousVersion: currentStableMeta.Version, NewVersion: model.Version}, nil
}

func (s *ModelsService) UploadModel(ctx context.Context, id app.ServableID, file io.Reader, label ...string) (*app.ModelID, error) {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name}
	version, err := s.metadata.NextVersion(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	_, err = s.storage.SaveModel(ctx, id, int(version), file)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	modelID := app.ModelID{ServableID: id, Version: version, Label: ""}

	metaID, err := s.metadata.Add(ctx, app.ModelData{ModelID: modelID, Status: app.StatusPending})
	if err != nil {
		errRemoveModel := s.storage.RemoveModel(ctx, app.ServableID{Team: id.Team, Project: id.Project, Name: id.Name}, version)
		if errRemoveModel != nil {
			err = exterr.WrapWithErr(err, errRemoveModel)
		}
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	modelIDWithLabel := modelID
	modelIDWithLabel.Label = s.servingConfig.DefaultLabel()
	if len(label) != 0 {
		modelIDWithLabel.Label = label[0]
	}

	if err := s.servingConfig.AddModel(ctx, modelIDWithLabel); err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	if err := s.metadata.UpdateStatus(ctx, metaID, app.StatusReady); err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	model := app.ModelData{ModelID: modelIDWithLabel, Status: app.StatusReady}
	if err := s.metadata.ChangeLabel(ctx, model); err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return &modelID, nil
}

func (s *ModelsService) RemoveByLabel(ctx context.Context, id app.ServableID, label string) error {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name, "label": label}
	err := s.removeModel(ctx, id, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return err
	}
	return nil
}

func (s *ModelsService) RemoveByVersion(ctx context.Context, id app.ServableID, version int64) error {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name, "version": version}
	err := s.removeModel(ctx, id, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return err
	}
	return nil
}

func (s *ModelsService) RemoveModelLabel(ctx context.Context, id app.ServableID, label string) error {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name, "label": label}
	err := s.removeModelLabel(ctx, id, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return err
	}
	return nil
}

func (s *ModelsService) removeModelLabel(ctx context.Context, id app.ServableID, params app.QueryParameters) error {
	modelMeta, err := s.metadata.Get(ctx, params)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	if modelMeta == nil {
		return errorModelNotFound
	}

	modelID := app.ModelID{ServableID: id, Label: modelMeta.Label}
	if err := s.servingConfig.RemoveModelLabel(ctx, modelID); err != nil {
		return err
	}

	if err := s.metadata.RemoveLabel(ctx, app.ModelData{ModelID: modelID}); err != nil {
		return exterr.WrapWithFrame(err)
	}

	return nil
}

func (s *ModelsService) removeModel(ctx context.Context, id app.ServableID, params app.QueryParameters) error {
	modelsMetas, err := s.metadata.List(ctx, params)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}
	if modelsMetas == nil {
		return errorModelNotFound
	}

	version := int64(-1)
	for _, modelMeta := range modelsMetas {
		if version < 0 {
			version = modelMeta.Version
		} else if version != modelMeta.Version {
			logging.Warn(ctx, fmt.Sprintf("got different versions [%d and %d] for model: %s", version, modelMeta.Version, id.InstanceName()))
		}
	}

	modelID := app.ModelID{ServableID: id, Version: version}
	if err := s.servingConfig.RemoveModel(ctx, modelID); err != nil {
		return exterr.WrapWithFrame(err)
	}

	if _, err := s.servingReload.ReloadConfig(ctx, id.Team, id.Project, true); err != nil {
		return exterr.WrapWithFrame(err)
	}

	if err := s.storage.RemoveModel(ctx, modelID.ServableID, modelID.Version); err != nil {
		return exterr.WrapWithFrame(err)
	}

	for _, modelMeta := range modelsMetas {
		if err := s.metadata.Delete(ctx, modelMeta.ID); err != nil {
			return exterr.WrapWithFrame(err)
		}
	}

	return nil
}
