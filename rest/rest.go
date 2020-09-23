package rest

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/lock"
	"github.com/grupawp/tensorflow-deploy/logging"
)

var (
	logInvalidChecksumErrorCode = 1001
	errorInvalidChecksum        = exterr.NewErrorWithMessage("invalid checksum").WithComponent(app.ComponentRest).WithCode(logInvalidChecksumErrorCode)
)

// REST represents restful API methods
type REST struct {
	modelsService  ModelsService
	modulesService ModulesService

	uploadFileName     string
	uploadFileChecksum string
	listenPort         string
	version            string

	lock *lock.Lock
}

// NewREST returns new instance of REST struct
func NewREST(modelsSrv ModelsService, modulesSrv ModulesService, listenPort, version string) *REST {
	l := lock.New()

	return &REST{
		modelsService:      modelsSrv,
		modulesService:     modulesSrv,
		uploadFileName:     "archive_data",
		uploadFileChecksum: "archive_hash",
		listenPort:         listenPort,
		version:            version,
		lock:               l,
	}
}

// Mount mounts each restful endpoints into router
func (rest *REST) Mount() {
	r := chi.NewRouter()

	// logging middlewares
	r.Use(logging.HTTPCtxValuesMiddleware)
	r.Use(logging.HTTPRequestMiddleware())

	// common
	r.Get("/ping", rest.pingHandler)

	// v3: model
	r.Route("/v1/models", func(r chi.Router) {
		r.Get("/list", rest.listModelsHandler)
	})

	r.Route("/v1/models/{team}/{project}", func(r chi.Router) {
		r.Get("/config", rest.configFileHandler)
		r.Get("/list", rest.listModelsByProjectHandler)
		r.Post("/reload", rest.reloadHandler)
	})

	r.Route("/v1/models/{team}/{project}/names/{name}", func(r chi.Router) {
		r.Post("/", rest.uploadModelHandler)
		r.Get("/list", rest.listModelsByNameHandler)
		r.Put("/revert", rest.revertModelHandler)
	})

	r.Route("/v1/models/{team}/{project}/names/{name}/labels/{label}", func(r chi.Router) {
		r.Get("/", rest.downloadModelByLabelHandler)
		r.Delete("/", rest.deleteModelLabelHandler)
		r.Post("/", rest.uploadModelWithLabelHandler)
		r.Delete("/remove_version", rest.deleteModelByLabelHandler)
	})

	r.Route("/v1/models/{team}/{project}/names/{name}/versions/{version}", func(r chi.Router) {
		r.Get("/", rest.downloadModelByVersionHandler)
		r.Delete("/", rest.deleteModelByVersionHandler)
		r.Put("/labels/stable", rest.setModelLabelToStableHandler)
		r.Put("/labels/{label}", rest.setModelLabelHandler)
	})

	// v3: module
	r.Route("/v1/modules", func(r chi.Router) {
		r.Get("/list", rest.listModulesHandler)
	})
	r.Route("/v1/modules/{team}/{project}", func(r chi.Router) {
		r.Get("/list", rest.listModulesByProjectHandler)
	})
	r.Route("/v1/modules/{team}/{project}/names/{name}", func(r chi.Router) {
		r.Post("/", rest.uploadModuleHandler)
		r.Get("/list", rest.listModulesByNameHandler)
		r.Get("/versions/{version}", rest.downloadModuleByVersionHandler)
		r.Delete("/versions/{version}", rest.deleteModuleHandler)
	})

	server := &http.Server{Addr: rest.listenPort, Handler: r}
	server.ListenAndServe()
}

func (rest *REST) pingHandler(w http.ResponseWriter, r *http.Request) {
	writeJSONSuccessResponse(w, r, http.StatusOK, struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}{
		Name:    fmt.Sprintf("%s:%s", "tensorflow-deploy", rest.version),
		Version: rest.version,
	})
}
