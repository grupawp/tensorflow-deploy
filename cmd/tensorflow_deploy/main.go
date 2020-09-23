package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/config"
	"github.com/grupawp/tensorflow-deploy/lock"
	"github.com/grupawp/tensorflow-deploy/logging"
	"github.com/grupawp/tensorflow-deploy/metadata/sqldb"
	"github.com/grupawp/tensorflow-deploy/rest"
	"github.com/grupawp/tensorflow-deploy/service"
	"github.com/grupawp/tensorflow-deploy/serving"
	"github.com/grupawp/tensorflow-deploy/storage"
	"github.com/grupawp/tensorflow-deploy/storage/filesystem"
)

// VERSION - service version initialized during build process
var VERSION string

var (
	logStoragerErrorCode  = "1001"
	logDiscoveryErrorCode = "1002"
	logConfigErrorCode    = "1003"
	logConfErrorCode      = "1004"
	logServingErrorCode   = "1004"
	logSQLDBErrorCode     = "1005"
)

func main() {

	ctx := context.Background()
	conf, err := config.NewConfig(ctx)
	if err != nil {
		if errors.Is(err, app.ErrCLIUsage) {
			return
		}
		logging.FatalErrorWithStack(ctx, err, logConfigErrorCode)
	}

	mainConfig, err := conf.Parse(ctx)
	if err != nil {
		logging.FatalErrorWithStack(ctx, err, logConfErrorCode)
	}

	storageImpl, err := filesystem.NewStorager(&mainConfig.Storage.Filesystem)
	if err != nil {
		logging.FatalErrorWithStack(ctx, err, logStoragerErrorCode)
	}

	modelsStorage := storage.NewModelsStorage(storageImpl)
	servingConf, err := serving.NewServableConfig(modelsStorage, *mainConfig.App.DefaultModelLabel)
	if err != nil {
		logging.FatalErrorWithStack(ctx, err, logServingErrorCode)
	}

	discovery, err := NewServiceDiscovery(*mainConfig.App.Discovery, mainConfig.Discovery)
	if err != nil {
		logging.FatalErrorWithStack(ctx, err, logDiscoveryErrorCode)
	}

	meta, err := sqldb.NewSQLDB(ctx, *mainConfig.Metadata.SQLDB.Driver, *mainConfig.Metadata.SQLDB.DSN)
	if err != nil {
		logging.FatalErrorWithStack(ctx, err, logSQLDBErrorCode)
	}

	defer meta.Close(ctx)

	servingReloader := serving.NewModelsReloader(discovery, meta.Model, servingConf, lock.New(), *mainConfig.App.ReloadIntervalInSec, *mainConfig.App.MaxAutoReloadDurationInSec, *mainConfig.App.AllowLabelsForUnavailableModels)

	modelsSvc := service.NewModelsService(meta.Model, servingConf, servingReloader, modelsStorage)

	modulesStorage := storage.NewModuleStorage(storageImpl)
	modulesSvc := service.NewModulesService(meta.Module, modulesStorage)

	api := rest.NewREST(modelsSvc, modulesSvc, mainConfig.App.Listen(), VERSION)

	logging.Info(context.Background(), fmt.Sprintf("%s v%s is up" /*service.ServiceName*/, "tensorflow-deploy", VERSION))
	logging.Info(context.Background(), fmt.Sprintf("REST listening on %s", mainConfig.App.Listen()))

	go servingReloader.ReloadInstancesJob(ctx)
	api.Mount()
}
