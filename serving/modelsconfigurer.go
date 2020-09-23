package serving

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/logging"
	"github.com/grupawp/tensorflow-deploy/storage"

	tfsConfig "github.com/grupawp/tensorflow-deploy/serving/protobuf/tensorflow_serving/config"
	tfsStoragePath "github.com/grupawp/tensorflow-deploy/serving/protobuf/tensorflow_serving/sources/storage_path"
)

type ServableConfig struct {
	m            *sync.RWMutex
	storage      ConfigStorage
	defaultLabel string
}

// NewServableConfig returns new instance of ServableConfig
func NewServableConfig(configStorage ConfigStorage, defaultLabel string) (*ServableConfig, error) {
	return &ServableConfig{m: &sync.RWMutex{}, storage: configStorage, defaultLabel: defaultLabel}, nil
}

func (sc *ServableConfig) DefaultLabel() string {
	return sc.defaultLabel
}

// ConfigWithoutLabels return parsed ModelServerConfig struct from models.config file
func (sc *ServableConfig) ConfigWithoutLabels(ctx context.Context, team, project string) (*tfsConfig.ModelServerConfig, error) {
	msc, err := sc.Config(ctx, team, project)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	configs := msc.GetModelConfigList().GetConfig()
	for _, v := range configs {
		v.VersionLabels = nil
	}
	msc.GetModelConfigList().Config = configs
	return msc, nil
}

// ConfigFileStream returns byte stream
func (sc *ServableConfig) ConfigFileStream(ctx context.Context, team, project string) ([]byte, error) {
	msc, err := sc.Config(ctx, team, project)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	stringConf := proto.MarshalTextString(msc)
	return []byte(stringConf), nil
}

// Models returns data of models from config
func (sc *ServableConfig) Models(ctx context.Context, team, project string, msc *tfsConfig.ModelServerConfig) ([]app.ModelID, error) {
	result := []app.ModelID{}
	configs := msc.GetModelConfigList().GetConfig()
	for _, v := range configs {
		serableIDs := app.ServableID{
			Team:    team,
			Project: project,
			Name:    v.GetName(),
		}

		model := app.ModelID{
			ServableID: serableIDs,
		}
		for _, version := range v.GetModelVersionPolicy().GetSpecific().Versions {
			var versionValue int64
			versionValue = version
			model.Version = versionValue
			result = append(result, model)
		}
	}
	return result, nil
}

// AddModel adds new model to configuration file
func (sc *ServableConfig) AddModel(ctx context.Context, id app.ModelID) error {
	msc, err := sc.Config(ctx, id.Team, id.Project)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	// set default label
	if id.Label == "" {
		id.Label = sc.defaultLabel
	}

	configs := msc.GetModelConfigList().GetConfig()
	for k, v := range configs {
		if v.GetName() == id.Name {
			// add new version to model_version_policy
			availableVersions := v.GetModelVersionPolicy().GetSpecific().GetVersions()
			availableVersions = append(availableVersions, id.Version)
			v.GetModelVersionPolicy().GetSpecific().Versions = availableVersions

			// set label
			if id.Label != "" {
				vLabels := v.GetVersionLabels()
				if vLabels == nil {
					return errVersionNotFound
				}
				vLabels[id.Label] = id.Version
				configs[k].VersionLabels = vLabels
			}

			// save model and exit
			if err := sc.saveModel(ctx, msc, id.Team, id.Project); err != nil {
				return exterr.WrapWithFrame(err)
			}
			return nil
		}
	}

	// config not found, create new element
	config := &tfsConfig.ModelConfig{
		Name:          id.Name,
		BasePath:      fmt.Sprintf("/models/%s/%s/%s", id.Team, id.Project, id.Name),
		ModelPlatform: "tensorflow",
		ModelVersionPolicy: &tfsStoragePath.FileSystemStoragePathSourceConfig_ServableVersionPolicy{
			PolicyChoice: &tfsStoragePath.FileSystemStoragePathSourceConfig_ServableVersionPolicy_Specific_{
				Specific: &tfsStoragePath.FileSystemStoragePathSourceConfig_ServableVersionPolicy_Specific{
					Versions: []int64{1},
				},
			},
		},
		VersionLabels: map[string]int64{id.Label: id.Version},
	}

	configs = append(configs, config)
	msc.GetModelConfigList().Config = configs

	err = sc.saveModel(ctx, msc, id.Team, id.Project)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	return nil
}

// UpdateLabel sets specific label for model version identified by ModelID
func (sc *ServableConfig) UpdateLabel(ctx context.Context, id app.ModelID) (int64, error) {
	msc, err := sc.Config(ctx, id.Team, id.Project)
	if err != nil {
		return 0, exterr.WrapWithFrame(err)
	}

	var prevLabeledVersion int64
	configs := msc.GetModelConfigList().GetConfig()
	for k, v := range configs {
		if v.GetName() == id.Name {
			vLabels := v.GetVersionLabels()
			if vLabels == nil {
				vLabels = make(map[string]int64, 1)
			}
			prevLabeledVersion = vLabels[id.Label]
			vLabels[id.Label] = id.Version
			configs[k].VersionLabels = vLabels

			err = sc.saveModel(ctx, msc, id.Team, id.Project)
			if err != nil {
				return 0, exterr.WrapWithFrame(err)
			}

			return prevLabeledVersion, nil
		}
	}

	return 0, errUpdateLabel
}

// RemoveModel removes specific model for model version identified by ModelID
func (sc *ServableConfig) RemoveModel(ctx context.Context, id app.ModelID) error {
	msc, err := sc.Config(ctx, id.Team, id.Project)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}
	configs := msc.GetModelConfigList().GetConfig()
	for k, v := range configs {
		if v.GetName() == id.Name {
			labels := v.GetVersionLabels()
			if labels == nil {
				return errVersionNotFound
			}
			if isLabelStable(labels, id.Version) {
				return errRemovingModelHasStableLabel
			}

			availableVersions := v.GetModelVersionPolicy().GetSpecific().GetVersions()
			newVersions := []int64{}
			for _, version := range availableVersions {
				if version == id.Version {
					configs[k].VersionLabels = removeVersionLabels(labels, id.Version)
					continue
				}
				newVersions = append(newVersions, version)
			}
			if len(newVersions) != len(availableVersions) {
				v.GetModelVersionPolicy().GetSpecific().Versions = newVersions
				err = sc.saveModel(ctx, msc, id.Team, id.Project)
				if err != nil {
					return exterr.WrapWithFrame(err)
				}
				return nil
			}
			logging.Debug(ctx, fmt.Sprintf("serving: RemoveModel() exit without any changes for model: %s in version: %d", id.InstanceName(), id.Version))
		}
	}
	return nil
}

// RemoveModelLabel removes specific label of model identified by ModelID
func (sc *ServableConfig) RemoveModelLabel(ctx context.Context, id app.ModelID) error {
	msc, err := sc.Config(ctx, id.Team, id.Project)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	configs := msc.GetModelConfigList().GetConfig()
	for k, v := range configs {
		if v.GetName() == id.Name {
			labels := v.GetVersionLabels()
			if labels == nil {
				return errVersionNotFound
			}
			versionLabel, ok := versionByLabel(labels, id.Label)
			if !ok {
				return errVersionNotFound
			}

			if isLabelStable(labels, versionLabel) {
				return errRemovingModelHasStableLabel
			}

			availableVersions := v.GetModelVersionPolicy().GetSpecific().GetVersions()
			newVersions := []int64{}
			for _, version := range availableVersions {
				if version == versionLabel {
					configs[k].VersionLabels = removeLabel(labels, id.Label)
				}
				newVersions = append(newVersions, version)
			}

			v.GetModelVersionPolicy().GetSpecific().Versions = newVersions
			err = sc.saveModel(ctx, msc, id.Team, id.Project)
			if err != nil {
				return exterr.WrapWithFrame(err)
			}

			return nil
		}
	}
	return nil
}

// saveModel marshals ModelServerConfig struct into models.config file
func (sc *ServableConfig) saveModel(ctx context.Context, msc *tfsConfig.ModelServerConfig, team, project string) error {
	data, err := proto.Marshal(msc)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	sc.m.Lock()
	err = sc.storage.SaveConfig(ctx, team, project, data)
	sc.m.Unlock()
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	return nil
}

// Config returns valid ModelServerConfig struct read from storage
func (sc *ServableConfig) Config(ctx context.Context, team, project string) (*tfsConfig.ModelServerConfig, error) {
	sc.m.Lock()
	stream, err := sc.storage.ReadConfig(ctx, team, project)
	sc.m.Unlock()

	if err != nil && !errors.Is(err, storage.ErrConfigDoesNotExist) {
		return nil, exterr.WrapWithFrame(err)
	}

	msc := &tfsConfig.ModelServerConfig{}
	if err := proto.UnmarshalMerge(stream, msc); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	if msc.Config == nil {
		msc.Config = &tfsConfig.ModelServerConfig_ModelConfigList{
			ModelConfigList: &tfsConfig.ModelConfigList{},
		}
	}

	return msc, nil
}
