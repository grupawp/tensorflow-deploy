package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/service/mocks"
	"github.com/stretchr/testify/mock"
)

func modelID(team, project, name, label string, version int64) app.ModelID {
	return app.ModelID{ServableID: app.ServableID{Team: team, Project: project, Name: name}, Version: version, Label: label}
}

func modelData(team, project, name, label string, version int64) *app.ModelData {
	return &app.ModelData{ModelID: modelID(team, project, name, label, version)}
}

var (
	teamProjectNameVersion1WithoutLabel     = modelData("team", "project", "name", "", 1)
	teamProjectNameVersion1WithLabelKopytko = modelData("team", "project", "name", "Kopytko", 1)

	teamProjectNameVersion2WithoutLabel    = modelData("team", "project", "name", "", 2)
	teamProjectNameVersion2WithLabelCanary = modelData("team", "project", "name", "canary", 2)

	teamProjectNameVersion3WithoutLabel    = modelData("team", "project", "name", "", 3)
	teamProjectNameVersion3WithLabelStable = modelData("team", "project", "name", "stable", 3)
)

func Test_cleanList(t *testing.T) {
	tests := []struct {
		name string
		args []*app.ModelData
		want []*app.ModelData
	}{
		{
			name: "List without any labeled models should return input list",
			args: []*app.ModelData{
				teamProjectNameVersion1WithoutLabel,
				teamProjectNameVersion2WithoutLabel,
				teamProjectNameVersion3WithoutLabel,
			},
			want: []*app.ModelData{
				teamProjectNameVersion1WithoutLabel,
				teamProjectNameVersion2WithoutLabel,
				teamProjectNameVersion3WithoutLabel,
			},
		},
		{
			name: "List with labeled models should return list without empty labels",
			args: []*app.ModelData{
				teamProjectNameVersion1WithoutLabel,
				teamProjectNameVersion1WithLabelKopytko,
				teamProjectNameVersion2WithoutLabel,
				teamProjectNameVersion2WithLabelCanary,
				teamProjectNameVersion3WithoutLabel,
				teamProjectNameVersion3WithLabelStable,
			},
			want: []*app.ModelData{
				teamProjectNameVersion1WithLabelKopytko,
				teamProjectNameVersion2WithLabelCanary,
				teamProjectNameVersion3WithLabelStable,
			},
		},
		{
			name: "List with mixed labeled or not models should return list without empty labels when label",
			args: []*app.ModelData{
				teamProjectNameVersion1WithoutLabel,
				teamProjectNameVersion1WithLabelKopytko,
				teamProjectNameVersion2WithoutLabel,
				teamProjectNameVersion3WithoutLabel,
				teamProjectNameVersion3WithLabelStable,
			},
			want: []*app.ModelData{
				teamProjectNameVersion1WithLabelKopytko,
				teamProjectNameVersion2WithoutLabel,
				teamProjectNameVersion3WithLabelStable,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanList(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cleanList() = %v, want %v", got, tt.want)
			}
		})
	}
}

var GetMetaErrorOnListResponse = func(params app.QueryParameters, err error) ModelsMetadata {
	mm := new(mocks.ModelsMetadata)
	mm.On("List", mock.Anything, params).Return(nil, err)
	return mm
}

var GetMetaListResponse = func(params app.QueryParameters, metaResponse []*app.ModelData) ModelsMetadata {
	mm := new(mocks.ModelsMetadata)
	mm.On("List", mock.Anything, params).Return(metaResponse, nil)
	return mm
}

func TestModelsService_ListModelsByProject(t *testing.T) {

	type args struct {
		metadata ModelsMetadata
		team     string
		project  string
	}
	tests := []struct {
		name    string
		args    args
		want    []*app.ModelData
		wantErr bool
	}{
		{
			name: "Error on List() should return an error",
			args: args{
				metadata: GetMetaErrorOnListResponse(map[string]interface{}{"team": "testTeam", "project": "testProject"}, errors.New("random error")),
				team:     "testTeam",
				project:  "testProject",
			},
			wantErr: true,
		},
		{
			name: "Valid List() parameters should return some results without any errors",
			args: args{
				team:    "testTeam",
				project: "testProject",
				metadata: GetMetaListResponse(map[string]interface{}{"team": "testTeam", "project": "testProject"}, []*app.ModelData{{
					ModelID: app.ModelID{
						ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
						Version:    1, Label: "testLabel"}},
				}),
			},
			want: []*app.ModelData{{
				ModelID: app.ModelID{
					ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
					Version:    1, Label: "testLabel"}},
			},
		},
		{
			name: "Valid List() parameters should return some results without any errors and without empty labels",
			args: args{
				team:    "testTeam",
				project: "testProject",
				metadata: GetMetaListResponse(map[string]interface{}{"team": "testTeam", "project": "testProject"}, []*app.ModelData{{
					ModelID: app.ModelID{
						ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
						Version:    1, Label: ""}}, {
					ModelID: app.ModelID{
						ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
						Version:    1, Label: "testLabel"}},
				}),
			},
			want: []*app.ModelData{
				{ModelID: app.ModelID{
					ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
					Version:    1, Label: "testLabel"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ModelsService{
				metadata: tt.args.metadata,
			}
			got, err := s.ListModelsByProject(context.Background(), tt.args.team, tt.args.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModelsService.ListModelsByProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModelsService.ListModelsByProject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelsService_ListModelsByName(t *testing.T) {
	type args struct {
		metadata ModelsMetadata
		id       app.ServableID
	}
	tests := []struct {
		name    string
		args    args
		want    []*app.ModelData
		wantErr bool
	}{
		{
			name: "Error on List() should return an error",
			args: args{
				metadata: GetMetaErrorOnListResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName"},
					errors.New("random error")),
				id: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
			},
			wantErr: true,
		},
		{
			name: "Valid List() parameters should return some results without any errors",
			args: args{
				id: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
				metadata: GetMetaListResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName"},
					[]*app.ModelData{{
						ModelID: app.ModelID{
							ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
							Version:    1, Label: "testLabel"}},
					}),
			},
			want: []*app.ModelData{{
				ModelID: app.ModelID{
					ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
					Version:    1, Label: "testLabel"}},
			},
		},
		{
			name: "Valid List() parameters should return some results without any errors and without empty labels",
			args: args{
				id: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
				metadata: GetMetaListResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName"},
					[]*app.ModelData{{
						ModelID: app.ModelID{
							ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
							Version:    1, Label: ""}}, {
						ModelID: app.ModelID{
							ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
							Version:    1, Label: "testLabel"}},
					}),
			},
			want: []*app.ModelData{{
				ModelID: app.ModelID{
					ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
					Version:    1, Label: "testLabel"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ModelsService{
				metadata: tt.args.metadata,
			}
			got, err := s.ListModelsByName(context.Background(), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModelsService.ListModelsByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModelsService.ListModelsByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

var GetConfigFileStreamRersponse = func(team, project string, response []byte, err error) ModelsConfig {
	sc := new(mocks.ModelsConfig)
	sc.On("ConfigFileStream", mock.Anything, team, project).Return(response, err)
	return sc
}

func TestModelsService_GetConfigStream(t *testing.T) {
	type args struct {
		servingConfig ModelsConfig
		team          string
		project       string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Error on ConfigFileStream() should return an error",
			args: args{
				servingConfig: GetConfigFileStreamRersponse("testTeam", "testProject", nil, errors.New("random error")),
				team:          "testTeam",
				project:       "testProject",
			},
			wantErr: true,
		},
		{
			name: "Valid ConfigFileStream() parameters should return valid config in bytes stream",
			args: args{
				servingConfig: GetConfigFileStreamRersponse("testTeam", "testProject", []byte("models_config_list: {}"), nil),
				team:          "testTeam",
				project:       "testProject",
			},
			want: []byte("models_config_list: {}"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ModelsService{
				servingConfig: tt.args.servingConfig,
			}
			got, err := s.GetConfigStream(context.Background(), tt.args.team, tt.args.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModelsService.GetConfigStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModelsService.GetConfigStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

var GetMetaOnGetResponse = func(params app.QueryParameters, metaResponse *app.ModelData, err error) ModelsMetadata {
	mm := new(mocks.ModelsMetadata)
	mm.On("Get", mock.Anything, params).Return(metaResponse, err)
	return mm
}

var GetStorageReadModelResponse = func(id app.ServableID, version int, storageResponse []byte, err error) ModelStorage {
	mm := new(mocks.ModelStorage)
	mm.On("ReadModel", mock.Anything, id, version).Return(storageResponse, err)
	return mm
}

func TestModelsService_ArchiveByLabel(t *testing.T) {
	type fields struct {
		metadata ModelsMetadata
		storage  ModelStorage
	}
	type args struct {
		id    app.ServableID
		label string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *app.Archive
		wantErr bool
	}{
		{
			name: "Error on metadata Get() should return an error",
			fields: fields{
				metadata: GetMetaOnGetResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName", "label": "testLabel"},
					nil, errors.New("random error")),
			},
			args: args{
				id:    app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
				label: "testLabel",
			},
			wantErr: true,
		},
		{
			name: "Empty response on metadata Get() return an error",
			fields: fields{
				metadata: GetMetaOnGetResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName", "label": "testLabel"},
					nil, nil),
			},
			args: args{
				id:    app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
				label: "testLabel",
			},
			wantErr: true,
		},
		{
			name: "Valid response on metadata Get() and error on storage ReadModel() return an error",
			fields: fields{
				metadata: GetMetaOnGetResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName", "label": "testLabel"},
					&app.ModelData{ModelID: app.ModelID{
						ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
						Version:    1, Label: "testLabel"}},
					nil),
				storage: GetStorageReadModelResponse(app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"}, 1, nil, errors.New("random error")),
			},
			args: args{
				id:    app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
				label: "testLabel",
			},
			wantErr: true,
		},
		{
			name: "Valid response on metadata Get() and storage ReadModel() should return valid archive",
			fields: fields{
				metadata: GetMetaOnGetResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName", "label": "testLabel"},
					&app.ModelData{
						ModelID: app.ModelID{
							ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
							Version:    1, Label: "testLabel"}},
					nil),
				storage: GetStorageReadModelResponse(app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"}, 1,
					[]byte("archive"), nil),
			},
			args: args{
				id:    app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
				label: "testLabel",
			},
			want: &app.Archive{Data: []byte("archive"), Name: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"}.ArchiveName((&ModelsService{}).archivePrefix(), int64(1))},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ModelsService{
				metadata: tt.fields.metadata,
				storage:  tt.fields.storage,
			}
			got, err := s.ArchiveByLabel(context.Background(), tt.args.id, tt.args.label)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModelsService.ArchiveByLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModelsService.ArchiveByLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelsService_ArchiveByVersion(t *testing.T) {
	type fields struct {
		metadata ModelsMetadata
		storage  ModelStorage
	}
	type args struct {
		id      app.ServableID
		version int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *app.Archive
		wantErr bool
	}{
		{
			name: "Error on metadata Get() should return an error",
			fields: fields{
				metadata: GetMetaOnGetResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName", "version": int64(1)},
					nil, errors.New("random error")),
			},
			args: args{
				id:      app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
				version: 1,
			},
			wantErr: true,
		},
		{
			name: "Empty response on metadata Get() return an error",
			fields: fields{
				metadata: GetMetaOnGetResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName", "version": int64(1)},
					nil, nil),
			},
			args: args{
				id:      app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
				version: 1,
			},
			wantErr: true,
		},
		{
			name: "Valid response on metadata Get() and error on storage ReadModel() return an error",
			fields: fields{
				metadata: GetMetaOnGetResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName", "version": int64(1)},
					&app.ModelData{
						ModelID: app.ModelID{
							ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
							Version:    1, Label: "testLabel"}},
					nil),
				storage: GetStorageReadModelResponse(app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"}, 1, nil, errors.New("random error")),
			},
			args: args{
				id:      app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
				version: 1,
			},
			wantErr: true,
		},
		{
			name: "Valid response on metadata Get() and storage ReadModel() should return valid archive",
			fields: fields{
				metadata: GetMetaOnGetResponse(map[string]interface{}{"team": "testTeam", "project": "testProject", "name": "testName", "version": int64(1)},
					&app.ModelData{
						ModelID: app.ModelID{
							ServableID: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
							Version:    1, Label: "testLabel"}},
					nil),
				storage: GetStorageReadModelResponse(app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"}, 1,
					[]byte("archive"), nil),
			},
			args: args{
				id:      app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"},
				version: 1,
			},
			want: &app.Archive{Data: []byte("archive"), Name: app.ServableID{Team: "testTeam", Project: "testProject", Name: "testName"}.ArchiveName((&ModelsService{}).archivePrefix(), int64(1))},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ModelsService{
				metadata: tt.fields.metadata,
				storage:  tt.fields.storage,
			}
			got, err := s.ArchiveByVersion(context.Background(), tt.args.id, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModelsService.ArchiveByVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModelsService.ArchiveByVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

var reloadConfigRersponse = func(team, project string, skipConfigWithoutLabels bool, response []app.ReloadResponse, err error) ModelsReload {
	sc := new(mocks.ModelsReload)
	sc.On("ReloadConfig", mock.Anything, team, project, skipConfigWithoutLabels).Return(response, err)
	return sc
}

func TestModelsService_ReloadModels(t *testing.T) {
	type fields struct {
		servingReload ModelsReload
	}
	type args struct {
		team                    string
		project                 string
		skipConfigWithoutLabels bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []app.ReloadResponse
		wantErr bool
	}{
		{
			name: "Error on ReloadConfig() should return an error",
			fields: fields{
				servingReload: reloadConfigRersponse("testTeam", "testProject", false, nil, errors.New("random error")),
			},
			args: args{
				team:    "testTeam",
				project: "testProject",
			},
			wantErr: true,
		},
		{
			name: "Valid ReloadConfig() parameters should return valid ReloadResponse struct",
			fields: fields{
				servingReload: reloadConfigRersponse("testTeam", "testProject", false,
					[]app.ReloadResponse{}, nil),
			},
			args: args{
				team:    "testTeam",
				project: "testProject",
			},
			want: []app.ReloadResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ModelsService{
				servingReload: tt.fields.servingReload,
			}
			got, err := s.ReloadModels(context.Background(), tt.args.team, tt.args.project, tt.args.skipConfigWithoutLabels)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModelsService.ReloadModels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModelsService.ReloadModels() = %v, want %v", got, tt.want)
			}
		})
	}
}
