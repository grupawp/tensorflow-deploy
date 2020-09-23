package serving

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/grupawp/tensorflow-deploy/exterr"
	tfsApis "github.com/grupawp/tensorflow-deploy/serving/protobuf/tensorflow_serving/apis"
	tfsConfig "github.com/grupawp/tensorflow-deploy/serving/protobuf/tensorflow_serving/config"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/lock"
	"github.com/grupawp/tensorflow-deploy/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

const (
	autoReloadLockID        = "ReloadInstancesIfNecessary"
	logDelimiter            = " "
	maxReloadAttempts       = 2
	oneUnit                 = 1
	timeToWaitForNextReload = 150 * time.Millisecond
)

var (
	infoNumberServablesToReload                = "number of servables to reload:"
	infoServableInstancesToReload              = "servable instances which should be reloaded"
	infoAddedInvalidServableInstancesToReload  = "added invalid servable instances which should be reloaded"
	infoInvalidServableInstancesToReload       = "invalid servable instances which should be reloaded"
	infoNumberInvalidServableInstancesToReload = "number invalid servable instances"
	infoReloadInstancesJob                     = "reload instances job start"
	infoReloadInstancesJobEnd                  = "reload instances job end"
	infoAutoReloadEnd                          = "auto-reload end"
	infoReloadConfigInstances                  = "reload config instances"
	infoConfig                                 = "config doesn't exist for team project"
	infoNextReload                             = "next reload"
	infoReloadSuccess                          = "reload success"
	infoSkipUnlockAutoReloadAction             = "skip unlock auto-reload action"
	infoUnlockAutoReloadAction                 = "unlock auto-reload action"

	messageResponseStatusError = "response status is invalid"

	logUpdateLabelErrorCode                 = 1001
	logRemovingModelHasStableLabelErrorCode = 1002
	logResponseStatusErrorCode              = 1003
	logVersionNotFoundErrorCode             = 1006
	logDialErrorCode                        = "1004"
	logReloadConfigRequestErrorCode         = "1005"

	errUpdateLabel                 = exterr.NewErrorWithMessage("name not found").WithComponent(app.ComponentServing).WithCode(logUpdateLabelErrorCode)
	errRemovingModelHasStableLabel = exterr.NewErrorWithMessage("model has stable label").WithComponent(app.ComponentServing).WithCode(logRemovingModelHasStableLabelErrorCode)
	errVersionNotFound             = exterr.NewErrorWithMessage("version not found").WithComponent(app.ComponentServing).WithCode(logVersionNotFoundErrorCode)
)

type ModelsMetadata interface {
	ListUniqueTeamProject(ctx context.Context) ([]*app.ServableID, error)
}

type Reloader interface {
	ReloadInstances(ctx context.Context, id app.ServableID) (instanceErrorList *[]string, err error)
	ReloadInstancesJob(ctx context.Context)
}

type Discoverer interface {
	Discover(ctx context.Context, model app.ServableID) ([]string, error)
}

type ServableConfigurer interface {
	Config(ctx context.Context, team, project string) (*tfsConfig.ModelServerConfig, error)
	ConfigWithoutLabels(ctx context.Context, team, project string) (*tfsConfig.ModelServerConfig, error)
	Models(ctx context.Context, team, project string, msc *tfsConfig.ModelServerConfig) ([]app.ModelID, error)
}

type ModelsReloader struct {
	serviceDiscovery                Discoverer
	servableConfigurer              ServableConfigurer
	lock                            *lock.Lock
	reloadInterval                  int
	maxDurationAutoReload           int
	modelsMetadata                  ModelsMetadata
	lastStateInstancesOfModel       map[string]app.ServableInstances
	allowLabelsForUnavailableModels bool
}

// NewModelsReloader returns new instance of ModelsReloader
func NewModelsReloader(serviceDiscovery Discoverer, modelMetadata ModelsMetadata, modelsConfig ServableConfigurer, lock *lock.Lock, reloadInterval, maxDurationAutoReload int, allowLabelsForUnavailableModels bool) *ModelsReloader {
	return &ModelsReloader{serviceDiscovery: serviceDiscovery, modelsMetadata: modelMetadata, servableConfigurer: modelsConfig, lock: lock, reloadInterval: reloadInterval, maxDurationAutoReload: maxDurationAutoReload, allowLabelsForUnavailableModels: allowLabelsForUnavailableModels}
}

// ReloadConfig  reloads all instances
func (r *ModelsReloader) ReloadConfig(ctx context.Context, team, project string, skipConfigWithoutLabels bool) ([]app.ReloadResponse, error) {
	if _, err := r.reloadConfig(ctx, app.ServableID{Team: team, Project: project}, r.allowLabelsForUnavailableModels && skipConfigWithoutLabels); err != nil {
		return nil, err
	}
	return nil, nil
}

// ReloadInstancesJob reloads TFS instances
func (r *ModelsReloader) ReloadInstancesJob(ctx context.Context) {
	for {
		logging.Info(ctx, infoReloadInstancesJob)
		r.ReloadInstancesIfIsNecessary(ctx)
		logging.Info(ctx, infoReloadInstancesJobEnd)
		time.Sleep(time.Duration(r.reloadInterval) * time.Second)
	}
}

// ReloadInstancesIfIsNecessary reloads instances of model
func (r *ModelsReloader) ReloadInstancesIfIsNecessary(ctx context.Context) {
	err := r.lock.LockID(autoReloadLockID)
	if err != nil {
		logging.ErrorWithStackWithoutRequestID(ctx, err)
		return
	}
	defer r.lock.UnLockID(autoReloadLockID)

	go r.reloadInstancesIfIsNecessary(ctx)

	for counter := 0; counter <= r.maxDurationAutoReload; counter++ {
		if !r.lock.IsLockedID(autoReloadLockID) {
			logging.Info(ctx, fmt.Sprintf("%s %d", infoSkipUnlockAutoReloadAction, counter))
			return
		}
		time.Sleep(oneUnit * time.Second)
	}
	logging.Info(ctx, fmt.Sprintf("%s", infoUnlockAutoReloadAction))
	return
}

func (r *ModelsReloader) reloadInstancesIfIsNecessary(ctx context.Context) {
	defer r.lock.UnLockID(autoReloadLockID)

	models, err := r.modelsMetadata.ListUniqueTeamProject(ctx)
	if err != nil {
		logging.ErrorWithStackWithoutRequestID(ctx, exterr.WrapWithFrame(err))
		return
	}

	servables := make([]app.ServableID, 0)

	for _, v := range models {
		m := app.ServableID{}
		m.Team = v.Team
		m.Project = v.Project
		servables = append(servables, m)
	}

	servablesToReload, err := r.instancesOfServablesToReload(ctx, servables)
	if err != nil {
		logging.ErrorWithStackWithoutRequestID(ctx, exterr.WrapWithFrame(err))
		return
	}

	logging.Info(ctx, fmt.Sprintf("%s %d", infoNumberServablesToReload, len(servablesToReload)))

	for _, v := range servablesToReload {
		logging.Info(ctx, fmt.Sprintf("%s %s %s %v", infoServableInstancesToReload, v.Team, v.Project, strings.Join(v.Instances, logDelimiter)))
		if _, err := r.reloadConfig(ctx, v.ServableID, r.allowLabelsForUnavailableModels, v.Instances...); err != nil {
			if err != nil {
				r.removeServableInstance(v)
				logging.ErrorWithStackWithoutRequestID(ctx, err)
			}
		}
	}
	logging.Info(ctx, fmt.Sprintf("%s", infoAutoReloadEnd))
}

func (r *ModelsReloader) removeServableInstance(servable app.ServableInstances) {
	newLastStateInstancesOfModel := make(map[string]app.ServableInstances, 0)
	for _, v := range r.lastStateInstancesOfModel {
		model := app.ServableInstances{ServableID: v.ServableID}
		if servable.InstanceName() != v.InstanceName() {
			model.Instances = v.Instances
		}
		newLastStateInstancesOfModel[v.InstanceName()] = model
	}
	r.lastStateInstancesOfModel = newLastStateInstancesOfModel
}

// InstancesOfServablesToReload returns list of instances which should reload
func (r *ModelsReloader) instancesOfServablesToReload(ctx context.Context, models []app.ServableID) ([]app.ServableInstances, error) {
	currentStateModels := make(map[string]app.ServableInstances, 0)
	invalidInstancesModel := []app.ServableInstances{}

	result := []app.ServableInstances{}
	for _, v := range models {
		currentInstances, _ := r.serviceDiscovery.Discover(ctx, v)
		modelWithInstances := app.ServableInstances{
			ServableID: v,
			Instances:  currentInstances,
		}
		currentStateModels[v.InstanceName()] = modelWithInstances

		config, err := r.servableConfigurer.Config(ctx, v.Team, v.Project)
		if err != nil {
			return result, err
		}

		versions := []int64{}

		// check only the first name and max version
		for _, configValue := range config.GetModelConfigList().GetConfig() {
			versions = configValue.ModelVersionPolicy.GetSpecific().GetVersions()
			numberVersions := len(versions)
			if numberVersions > 0 {
				invalidInstances := r.invalidInstancesModel(ctx, configValue.GetName(), versions[numberVersions-1], currentStateModels[v.InstanceName()].Instances)
				if len(invalidInstances) > 0 {
					modelWithInstances = app.ServableInstances{
						ServableID: v,
						Instances:  invalidInstances,
					}
					invalidInstancesModel = append(invalidInstancesModel, modelWithInstances)
					logging.Info(ctx, fmt.Sprintf("%s %s %s %v", infoInvalidServableInstancesToReload, v.Team, v.Project, strings.Join(invalidInstances, logDelimiter)))
				}
			}
			break
		}

	}

	if len(r.lastStateInstancesOfModel) == 0 {
		r.lastStateInstancesOfModel = currentStateModels
		return result, nil
	}

	for k, v := range currentStateModels {
		itemLastInstance, ok := r.lastStateInstancesOfModel[k]
		if !ok {
			continue
		}

		if !equal(itemLastInstance.Instances, v.Instances) {
			tmpModel := app.ServableInstances{
				ServableID: v.ServableID,
				Instances:  v.Instances,
			}
			result = append(result, tmpModel)
		}

	}
	r.lastStateInstancesOfModel = currentStateModels

	invalidInstancesToReload := r.invalidInstancesToReload(ctx, &result, &invalidInstancesModel)
	numberInvalidInstancesToReload := len(*invalidInstancesToReload)
	if numberInvalidInstancesToReload > 0 {
		result = append(result, *invalidInstancesToReload...)
	}
	logging.Info(ctx, fmt.Sprintf("%s %d", infoNumberInvalidServableInstancesToReload, numberInvalidInstancesToReload))

	return result, nil
}

// reloadTFSInstances sends requests with new configuration via gRPC to given TFS instances
// It returns list of TFS instances with successfully reloaded configuration
// And list with errors for another instances
func (r *ModelsReloader) reloadTFSInstances(ctx context.Context, config *tfsConfig.ModelServerConfig, instances []string) ([]string, []error) {
	var reloaderErrors []error
	var validInstances []string

	// prepare request with same configuration for each instance
	configRequest := &tfsApis.ReloadConfigRequest{Config: config}

	for _, instance := range instances {
		conn, err := grpc.Dial(fmt.Sprintf("dns:///%s", instance), grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
		if err != nil {
			wrappedDialError := exterr.WrapWithFrame(err)
			reloaderErrors = append(reloaderErrors, wrappedDialError)
			logging.Error(ctx, fmt.Sprintf("%s %v", instance, wrappedDialError), logDialErrorCode)
			continue
		}

		resp, err := tfsApis.NewModelServiceClient(conn).HandleReloadConfigRequest(ctx, configRequest)
		if err != nil {
			wrappedReloadConfigRequestError := exterr.WrapWithFrame(err)
			reloaderErrors = append(reloaderErrors, wrappedReloadConfigRequestError)
			logging.Error(ctx, fmt.Sprintf("%s %v", instance, wrappedReloadConfigRequestError), logReloadConfigRequestErrorCode)
			conn.Close()
			continue
		}

		if errMsg := resp.GetStatus().GetErrorMessage(); errMsg != "" {
			errResponseStatus := exterr.NewErrorWithMessage(fmt.Sprintf("%s status: %s", messageResponseStatusError, errMsg)).
				WithComponent(app.ComponentServing).WithCode(logResponseStatusErrorCode)
			reloaderErrors = append(reloaderErrors, errResponseStatus)
			logging.Error(ctx, fmt.Sprintf("%s %v", instance, errResponseStatus), strconv.Itoa(logResponseStatusErrorCode))
			conn.Close()
			continue
		}
		conn.Close()
		validInstances = append(validInstances, instance)
	}

	return validInstances, reloaderErrors
}

// ReloadConfig  reloads all instances
func (r *ModelsReloader) reloadConfig(ctx context.Context, id app.ServableID, labelsOnly bool, instances ...string) (*[]string, error) {
	var err error
	var retErrors []error
	if len(instances) == 0 {
		instances, err = r.serviceDiscovery.Discover(ctx, id)
		if err != nil {
			return nil, exterr.WrapWithFrame(err)
		}
	}

	logging.Info(ctx, fmt.Sprintf("%s %v", infoReloadConfigInstances, instances))

	// 1 step: get full config
	config, err := r.servableConfigurer.Config(ctx, id.Team, id.Project)
	if err != nil {
		if os.IsNotExist(err) {
			// there's no config yet, so return without any errors
			logging.Info(ctx, fmt.Sprintf("%s %s %s", infoConfig, id.Team, id.Project))
			return nil, nil
		}
		return nil, exterr.WrapWithFrame(err)
	}

	// 2 step: get config without labels
	if !labelsOnly {
		configWithoutLabels, err := r.servableConfigurer.ConfigWithoutLabels(ctx, id.Team, id.Project)
		if err != nil {
			return nil, exterr.WrapWithFrame(err)

		}
		instances, retErrors = r.reloadTFSInstances(ctx, configWithoutLabels, instances)
	} else {
		logging.Debug(ctx, fmt.Sprintf("skipping reload config without labels for: %s", id.InstanceName()))
	}

	// 3 step: reload config
	_, reloadedErrors := r.reloadTFSInstances(ctx, config, instances)

	// try to reload again config if errors occured
	attemptReload := 1
	for len(reloadedErrors) > 0 && attemptReload <= maxReloadAttempts {
		wait(timeToWaitForNextReload, attemptReload)
		attemptLogInfo := ""
		for k, v := range reloadedErrors {
			attemptLogInfo += fmt.Sprintf("attempt: %d, error: %d %s; ", attemptReload, k, v.Error())
		}

		logging.Info(ctx, fmt.Sprintf("%s %s", infoNextReload, attemptLogInfo))
		_, reloadedErrors = r.reloadTFSInstances(ctx, config, instances)
		attemptReload++
	}
	retErrors = append(retErrors, reloadedErrors...)

	if len(retErrors) == 0 {
		logging.Info(ctx, fmt.Sprintf("%s", infoReloadSuccess))
		return nil, nil
	}
	var instanceErrorList []string

	for _, v := range retErrors {
		instanceErrorList = append(instanceErrorList, v.Error())
	}
	return &instanceErrorList, exterr.WrapWithFrame(fmt.Errorf("%v", retErrors))
}

func (r *ModelsReloader) invalidInstancesToReload(ctx context.Context, instances *[]app.ServableInstances, invalidInstances *[]app.ServableInstances) *[]app.ServableInstances {
	result := &[]app.ServableInstances{}
	if len(*invalidInstances) == 0 {
		return result
	}

	if len(*instances) == 0 {
		for _, invalidInstance := range *invalidInstances {
			if len(invalidInstance.Instances) > 0 {
				invalidInstancesModel := app.ServableInstances{
					ServableID: invalidInstance.ServableID,
					Instances:  invalidInstance.Instances,
				}
				logging.Info(ctx, fmt.Sprintf("%s %s %s %v 00", infoAddedInvalidServableInstancesToReload, invalidInstancesModel.Team, invalidInstancesModel.Project, strings.Join(invalidInstancesModel.Instances, logDelimiter)))
				*result = append(*result, invalidInstancesModel)
			}
		}

		return result
	}

	for _, invalidInstance := range *invalidInstances {
		for _, vInstance := range *instances {
			if vInstance.InstanceName() == invalidInstance.InstanceName() {
				instancesToAdd := []string{}
				for _, invalidInstanceIP := range invalidInstance.Instances {
					found := false
					for _, instanceToReload := range vInstance.Instances {
						if invalidInstanceIP == instanceToReload {
							found = true
						}
					}
					if !found {
						instancesToAdd = append(instancesToAdd, invalidInstanceIP)
					}
				}
				if len(instancesToAdd) > 0 {
					invalidInstancesModel := app.ServableInstances{
						ServableID: invalidInstance.ServableID,
						Instances:  instancesToAdd,
					}
					logging.Info(ctx, fmt.Sprintf("%s %s %s %v", infoAddedInvalidServableInstancesToReload, invalidInstancesModel.Team, invalidInstancesModel.Project, strings.Join(instancesToAdd, logDelimiter)))
					*result = append(*result, invalidInstancesModel)
				}
			}
		}
	}

	return result
}

func (r *ModelsReloader) invalidInstancesModel(ctx context.Context, name string, version int64, instances []string) []string {
	modelStatusRequest := &tfsApis.GetModelStatusRequest{
		ModelSpec: &tfsApis.ModelSpec{
			Name: name,
			VersionChoice: &tfsApis.ModelSpec_Version{
				Version: &wrappers.Int64Value{
					Value: version,
				},
			},
		},
	}

	invalidInstances := []string{}
	for _, instance := range instances {
		conn, err := grpc.Dial(fmt.Sprintf("dns:///%s", instance), grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
		if err != nil {
			invalidInstances = append(invalidInstances, instance)
			wrappedDialError := exterr.WrapWithFrame(err)
			logging.Error(ctx, fmt.Sprintf("%s %v", instance, wrappedDialError), logDialErrorCode)
			continue
		}

		resp, err := tfsApis.NewModelServiceClient(conn).GetModelStatus(ctx, modelStatusRequest)
		if err != nil {
			invalidInstances = append(invalidInstances, instance)
			conn.Close()
			continue
		}

		if resp != nil {
			modelStatus := resp.GetModelVersionStatus()[0].GetState()
			if modelStatus != tfsApis.ModelVersionStatus_AVAILABLE {
				invalidInstances = append(invalidInstances, instance)
				conn.Close()
				continue
			}
		}
		conn.Close()
	}

	return invalidInstances
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for _, aItem := range a {
		match := false
		for _, bItem := range b {
			if aItem == bItem {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	return true
}

func wait(unitWaitFor time.Duration, multiplier int) {
	time.Sleep(unitWaitFor * time.Duration(multiplier))
}
