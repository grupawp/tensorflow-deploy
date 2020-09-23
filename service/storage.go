package service

import (
	"context"
	"io"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/storage"
)

type ModelStorage interface {
	ReadModel(ctx context.Context, modelID app.ServableID, version int) ([]byte, error)
	ReadAllModels(ctx context.Context, modelID app.ServableID) ([]byte, error)

	SaveModel(ctx context.Context, modelID app.ServableID, version int, archive io.Reader) (*storage.SaveModelResponse, error)
	RemoveModel(ctx context.Context, id app.ServableID, version int64) error
}

type ModuleStorage interface {
	ReadModule(ctx context.Context, moduleID app.ServableID, version int) ([]byte, error)
	SaveModule(ctx context.Context, moduleID app.ServableID, version int, archive io.Reader) error
	RemoveModule(ctx context.Context, id app.ServableID, version int64) error
}
