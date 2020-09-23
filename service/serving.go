package service

import (
	"context"

	"github.com/grupawp/tensorflow-deploy/app"
)

type ModelsConfig interface {
	DefaultLabel() string
	AddModel(ctx context.Context, modelID app.ModelID) error
	ConfigFileStream(ctx context.Context, team, project string) ([]byte, error)
	RemoveModel(ctx context.Context, id app.ModelID) error
	RemoveModelLabel(ctx context.Context, id app.ModelID) error
	UpdateLabel(ctx context.Context, id app.ModelID) (int64, error)
}

type ModelsReload interface {
	ReloadConfig(ctx context.Context, team, project string, skipConfigWithoutLabels bool) ([]app.ReloadResponse, error)
}
