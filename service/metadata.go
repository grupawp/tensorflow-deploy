package service

import (
	"context"

	"github.com/grupawp/tensorflow-deploy/app"
)

// ModelsMetadata is an interface that contains necessary methods required to
// manage models medatada from a service level
type ModelsMetadata interface {
	Add(ctx context.Context, model app.ModelData) (int64, error)
	ChangeLabel(ctx context.Context, model app.ModelData) error
	Delete(ctx context.Context, id int64) error
	RemoveLabel(ctx context.Context, model app.ModelData) error
	Get(ctx context.Context, parameters app.QueryParameters) (*app.ModelData, error)
	List(ctx context.Context, parameters app.QueryParameters) ([]*app.ModelData, error)
	ListUniqueTeamProject(ctx context.Context) ([]*app.ServableID, error)
	NextVersion(ctx context.Context, parameters app.QueryParameters) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	IsStatusPending(ctx context.Context, servableID app.ServableID) (bool, error)
}

// ModulesMetadata is an interface that contains necessary methods required to
// manage modules medatada from a service level
type ModulesMetadata interface {
	Add(ctx context.Context, module app.ModuleData) (int64, error)
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, parameters app.QueryParameters) (*app.ModuleData, error)
	List(ctx context.Context, parameters app.QueryParameters) ([]*app.ModuleData, error)
	NextVersion(ctx context.Context, parameters app.QueryParameters) (int64, error)
}
