package service

import (
	"context"
	"io"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/logging"
)

var (
	moduleNotFoundErrorCode = 1004
	errorModuleNotFound     = exterr.NewErrorWithMessage("module not found").WithComponent(app.ComponentService).WithCode(moduleNotFoundErrorCode)
)

func (s *ModulesService) ListModules(ctx context.Context, params app.QueryParameters) ([]*app.ModuleData, error) {
	modules, err := s.metadata.List(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return modules, nil
}

func (s *ModulesService) ListModulesByProject(ctx context.Context, team, project string) ([]*app.ModuleData, error) {
	params := app.QueryParameters{"team": team, "project": project}
	modules, err := s.metadata.List(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return modules, nil
}

func (s *ModulesService) ListModulesByName(ctx context.Context, id app.ServableID) ([]*app.ModuleData, error) {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name}
	modules, err := s.metadata.List(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return modules, nil
}

func (s *ModulesService) UploadModule(ctx context.Context, id app.ServableID, file io.Reader) (*app.ModuleID, error) {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name}
	version, err := s.metadata.NextVersion(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	moduleID := app.ModuleID{ServableID: id, Version: version}

	if err := s.storage.SaveModule(ctx, id, int(version), file); err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	if _, err := s.metadata.Add(ctx, app.ModuleData{ModuleID: moduleID}); err != nil {
		errRemoveModel := s.storage.RemoveModule(ctx, app.ServableID{Team: id.Team, Project: id.Project, Name: id.Name}, version)
		if errRemoveModel != nil {
			err = exterr.WrapWithErr(err, errRemoveModel)
		}
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return &moduleID, nil
}

func (s *ModulesService) GetArchiveByVersion(ctx context.Context, id app.ServableID, version int64) (*app.Archive, error) {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name, "version": version}
	moduleMeta, err := s.metadata.Get(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	if moduleMeta == nil {
		logging.ErrorWithStack(ctx, errorModuleNotFound)
		return nil, errorModuleNotFound
	}

	archive, err := s.storage.ReadModule(ctx, id, int(version))
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return nil, err
	}

	return &app.Archive{Data: archive, Name: id.ArchiveName(s.archivePrefix(), version)}, nil
}

func (s *ModulesService) RemoveByVersion(ctx context.Context, id app.ServableID, version int64) error {
	params := app.QueryParameters{"team": id.Team, "project": id.Project, "name": id.Name, "version": version}
	moduleMeta, err := s.metadata.Get(ctx, params)
	if err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return err
	}

	if moduleMeta == nil {
		logging.ErrorWithStack(ctx, errorModuleNotFound)
		return errorModuleNotFound
	}

	if err := s.storage.RemoveModule(ctx, id, version); err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return err
	}

	if err := s.metadata.Delete(ctx, moduleMeta.ID); err != nil {
		logging.ErrorWithStack(ctx, exterr.WrapWithFrame(err))
		return err
	}

	return nil
}
