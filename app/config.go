package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"

	"github.com/grupawp/tensorflow-deploy/app/defaults"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/logging"
)

const (
	models   = "models"
	modules  = "modules"
	incoming = "incoming"
)

var (
	logUnsupportedDiscoverySourceErrorCode = 1001
	logUnsupportedStorageBackendErrorCode  = 1002
	logUnsupportedMetadataBackendErrorCode = 1003

	errUnsupportedDiscoverySource = exterr.NewErrorWithMessage("unsupported discovery source").WithComponent(ComponentAPP).WithCode(logUnsupportedDiscoverySourceErrorCode)
	errUnsupportedStorageBackend  = exterr.NewErrorWithMessage("unsupported storage backend").WithComponent(ComponentAPP).WithCode(logUnsupportedStorageBackendErrorCode)
	errUnsupportedMetadataBackend = exterr.NewErrorWithMessage("unsupported metadata backend").WithComponent(ComponentAPP).WithCode(logUnsupportedMetadataBackendErrorCode)

	ErrCLIUsage error = errors.New("cli usage")
)

type (
	// Config holds all configuration parameters
	Config struct {
		App       ConfigApp       `yaml:"application"`
		Discovery ConfigDiscovery `yaml:"discovery" group:"Discovery Options"`
		Storage   ConfigStorage   `yaml:"storage" group:"Storage Options"`
		Metadata  ConfigMetadata  `yaml:"metadata" group:"Metadata Options"`
	}

	// ConfigApp holds configuration parameters common to the application
	ConfigApp struct {
		ConfigFile                      *string `envconfig:"TFD_CONFIG_FILE" long:"config_file" description:"Path to the config file" default-mask:"not set"` // core parameter
		ListenHost                      *string `validate:"ip" defaults:"0.0.0.0" yaml:"listenHost" envconfig:"TFD_LISTEN_HOST" long:"listen_host" description:"Listen host" default-mask:"0.0.0.0"`
		ListenPort                      *uint16 `defaults:"9500" yaml:"listenPort" envconfig:"TFD_LISTEN_PORT" long:"listen_port" description:"Listen port" default-mask:"9500"`
		ReloadIntervalInSec             *int    `validate:"min=1" defaults:"300" yaml:"reloadIntervalInSec" envconfig:"TFD_RELOAD_INTERVAL_IN_SEC" long:"reload_interval_in_sec" description:"The interval of time after which the model configurations will be reloaded on TFS instances" default-mask:"300"`
		MaxAutoReloadDurationInSec      *int    `validate:"min=900" defaults:"900" yaml:"maxAutoReloadDurationInSec" envconfig:"TFD_MAX_AUTO_RELOAD_DURATION_IN_SEC" long:"max_auto_reload_duration_in_sec" description:"Max auto-reload duration" default-mask:"3600"`
		UploadTimeoutInSec              *int    `validate:"min=1" defaults:"300" yaml:"uploadTimeoutInSec" envconfig:"TFD_UPLOAD_TIMEOUT_IN_SEC" long:"upload_timeout_in_sec" description:"Timeout after which upload will be interrupted" default-mask:"300"`
		DefaultModelLabel               *string `defaults:"canary" yaml:"defaultModelLabel" envconfig:"TFD_DEFAULT_MODEL_LABEL" long:"default_model_label" description:"Default model label" default-mask:"canary"`
		AllowLabelsForUnavailableModels *bool   `defaults:"false" yaml:"tfsAllowsLabelsForUnavailableModels" envconfig:"TFD_TFS_ALLOWS_LABELS_FOR_UNAVAILABLE_MODELS" long:"tfs_allows_labels_for_unavailable_models" description:"If true, assume TFS instances allow assigning labels to models that are not available yet" default-mask:"false"`
		Discovery                       *string `validate:"oneof=plaintext dns" defaults:"dns" yaml:"discovery" envconfig:"TFD_DISCOVERY" long:"discovery" description:"Discovery source, see section of selected Discovery Options" choice:"plaintext" choice:"dns" default-mask:"dns"`
		Storage                         *string `validate:"oneof=filesystem" defaults:"filesystem" yaml:"storage" envconfig:"TFD_STORAGE" long:"storage" description:"Storage backend, see section of selected Storage Options" choice:"filesystem" default-mask:"filesystem"`
		Metadata                        *string `validate:"oneof=sqldb" defaults:"sqldb" yaml:"metadata" envconfig:"TFD_METADATA" long:"metadata" description:"Metadata backend, see section of selected Metadata Options" choice:"sqldb" default-mask:"sqldb"`
	}

	// ConfigDiscovery holds discovery package configuration parameters
	ConfigDiscovery struct {
		Plaintext ConfigDiscoveryPlaintext `yaml:"plaintext" group:"Plaintext Discovery Options"`
		DNS       ConfigDiscoveryDNS       `yaml:"dns" group:"DNS Discovery Options"`
	}
	// ConfigDiscoveryPlaintext holds Plaintext package configuration parameters
	ConfigDiscoveryPlaintext struct {
		HostsPath *string `validate:"file" defaults:"hosts" yaml:"hostsPath" envconfig:"TFD_DISCOVERY_PLAINTEXT_HOSTS_PATH" long:"discovery_plaintext_hosts_path" description:"Path to the file containing configuration of the TFS instances" default-mask:"hosts"`
	}
	// ConfigDiscoveryDNS holds DNS package configuration parameters
	ConfigDiscoveryDNS struct {
		ServiceSuffix       *string `defaults:"" yaml:"serviceSuffix" envconfig:"TFD_DISCOVERY_DNS_SERVICE_SUFFIX" long:"discovery_dns_service_suffix" description:"Service suffix with or without dot prefix; optional" default-mask:"not set"` // allowed empty string
		DefaultInstancePort *uint16 `validate:"min=1000,max=65535" defaults:"8500" yaml:"defaultInstancePort" envconfig:"TFD_DISCOVERY_DNS_DEFAULT_INSTANCE_PORT" long:"discovery_dns_default_instance_port" description:"Default TFS instance port in case SRV records are unavailable" default-mask:"8500"`
	}

	// ConfigStorage holds storage package configuration parameters
	ConfigStorage struct {
		Filesystem ConfigStorageFilesystem `yaml:"filesystem" group:"Filesystem Storage Options"`
	}
	// ConfigStorageFilesystem holds filesystem configuration parameters
	ConfigStorageFilesystem struct {
		Base   ConfigStorageFilesystemBase   `yaml:"base"`
		Model  ConfigStorageFilesystemModel  `yaml:"model"`
		Module ConfigStorageFilesystemModule `yaml:"module"`
	}
	// ConfigStorageFilesystemModel holds model configuration parameters
	ConfigStorageFilesystemModel struct {
		ArchiveName          *string `defaults:"model_archive.tar" yaml:"archiveName" envconfig:"TFD_STORAGE_FILESYSTEM_MODEL_ARCHIVE_NAME" long:"storage_filesystem_model_archive_name" description:"Model archive name" default-mask:"model_archive.tar"`
		BasePath             *string `defaults:"/tfdeploy/models" yaml:"basePath" envconfig:"TFD_STORAGE_FILESYSTEM_MODEL_BASE_PATH" long:"storage_filesystem_model_base_path" description:"models base path" default-mask:"/tfdeploy/models" hidden:"true"`
		ConfigName           *string `defaults:"models.config" yaml:"configName" envconfig:"TFD_STORAGE_FILESYSTEM_MODEL_CONFIG_NAME" long:"storage_filesystem_model_config_name" description:"Models config filename" default-mask:"models.config"`
		EmptyConfigName      *string `defaults:"empty.config" yaml:"emptyConfigName" envconfig:"TFD_STORAGE_FILESYSTEM_MODEL_EMPTY_CONFIG_NAME" long:"storage_filesystem_model_empty_config_name" description:"Empty config filename" default-mask:"empty.config"`
		IncomingArchivePath  *string `defaults:"/tfdeploy/incoming/models" yaml:"incomingArchivePath" envconfig:"TFD_STORAGE_FILESYSTEM_MODEL_INCOMING_ARCHIVE_PATH" long:"storage_filesystem_model_incoming_archive_path" description:"Incoming model archive path" default-mask:"/tfdeploy/incoming/models" hidden:"true"`
		DirectoryPermissions *string `defaults:"0755" yaml:"directoryPermissions" envconfig:"TFD_STORAGE_FILESYSTEM_MODEL_DIRECTORY_PERMISSIONS" long:"storage_filesystem_model_directory_permissions" description:"Model directory permissions" default-mask:"0755" hidden:"true"`
		FilePermissions      *string `defaults:"0644" yaml:"filePermissions" envconfig:"TFD_STORAGE_FILESYSTEM_MODEL_FILE_PERMISSIONS" long:"storage_filesystem_model_file_permissions" description:"Model file permissions" default-mask:"0644" hidden:"true"`
	}
	// ConfigStorageFilesystemModule holds module configuration parameters
	ConfigStorageFilesystemModule struct {
		ArchiveName          *string `defaults:"module_archive.tar" yaml:"archiveName" envconfig:"TFD_STORAGE_FILESYSTEM_MODULE_ARCHIVE_NAME" long:"storage_filesystem_module_archive_name" description:"Module archive name" default-mask:"module_archive.tar"`
		BasePath             *string `defaults:"/tfdeploy/modules" yaml:"basePath" envconfig:"TFD_STORAGE_FILESYSTEM_MODULE_BASE_PATH" long:"storage_filesystem_module_base_path" description:"Modules base path" default-mask:"/tfdeploy/modules" hidden:"true"`
		IncomingArchivePath  *string `defaults:"/tfdeploy/incoming/modules" yaml:"incomingArchivePath" envconfig:"TFD_STORAGE_FILESYSTEM_MODULE_INCOMING_ARCHIVE_PATH" long:"storage_filesystem_module_incoming_archive_path" description:"Incoming module archive path" default-mask:"/tfdeploy/incoming/modules" hidden:"true"`
		DirectoryPermissions *string `defaults:"0755" yaml:"directoryPermissions" envconfig:"TFD_STORAGE_FILESYSTEM_MODULE_DIRECTORY_PERMISSIONS" long:"storage_filesystem_module_directory_permissions" description:"Module directory permissions" default-mask:"0755" hidden:"true"`
		FilePermissions      *string `defaults:"0644" yaml:"filePermissions" envconfig:"TFD_STORAGE_FILESYSTEM_MODULE_FILE_PERMISSIONS" long:"storage_filesystem_module_file_permissions" description:"Module file permissions" default-mask:"0644" hidden:"true"`
	}

	ConfigStorageFilesystemBase struct {
		BasePath *string `defaults:"/tfdeploy" yaml:"basePath" envconfig:"TFD_STORAGE_FILESYSTEM_BASE_PATH" long:"storage_filesystem_base_path" description:"Base path sets: incoming model/module archive path, model/module base path if these paths aren't set" default-mask:"/tfdeploy"`
	}

	// ConfigMetadata holds metadata package configuration parameters
	ConfigMetadata struct {
		SQLDB ConfigMetadataSQLDB `yaml:"sqldb" group:"SQLDB Metadata Options"`
	}
	// ConfigMetadataSQLDB holds sqldb package configuration parameters
	ConfigMetadataSQLDB struct {
		Driver *string `validate:"oneof=sqlite3" defaults:"sqlite3" yaml:"driver" envconfig:"TFD_METADATA_SQLDB_DRIVER" long:"metadata_sqldb_driver" description:"SQL driver" choice:"sqlite3" default-mask:"sqlite3"`
		DSN    *string `validate:"min=1" defaults:"metadata.db" yaml:"dsn" envconfig:"TFD_METADATA_SQLDB_DSN" long:"metadata_sqldb_dsn" description:"The Data Source Name in common format like e.g. PEAR DB, but without type-prefix (optional parts marked by squared brackets)" default-mask:"metadata.db"`
	}
)

// Listen returns joined host with port
func (a *ConfigApp) Listen() string {
	return fmt.Sprintf("%s:%d", *a.ListenHost, *a.ListenPort)
}

// ConvertPerms converts string permission to FileMode
func (ConfigStorageFilesystem) ConvertPerms(perms string) (os.FileMode, error) {
	p, err := strconv.ParseUint(perms, 0, 32)
	if err != nil {
		return os.FileMode(0), exterr.WrapWithFrame(err)
	}

	return os.FileMode(p), nil
}

// Validate validates configuration parameters
func (c *Config) Validate(ctx context.Context) error {
	validate := validator.New()

	if err := validate.StructCtx(ctx, c.App); err != nil {
		return exterr.WrapWithFrame(err)
	}

	switch *c.App.Discovery {
	case "plaintext":
		if err := validate.StructCtx(ctx, c.Discovery.Plaintext); err != nil {
			return exterr.WrapWithFrame(err)
		}
	case "dns":
		if err := validate.StructCtx(ctx, c.Discovery.DNS); err != nil {
			return exterr.WrapWithFrame(err)
		}
	default:
		return errUnsupportedDiscoverySource
	}

	switch *c.App.Storage {
	case "filesystem":
		if err := validate.StructCtx(ctx, c.Storage.Filesystem); err != nil {
			return exterr.WrapWithFrame(err)
		}
	default:
		return errUnsupportedStorageBackend
	}

	switch *c.App.Metadata {
	case "sqldb":
		if err := validate.StructCtx(ctx, c.Metadata.SQLDB); err != nil {
			return exterr.WrapWithFrame(err)
		}
	default:
		return errUnsupportedMetadataBackend
	}

	return nil
}

// NewConfig creates unfilled config
func NewConfig() *Config {
	return &Config{}
}

// NewConfigDefaults loads configuration parameters from defaults
func NewConfigDefaults(ctx context.Context) (*Config, error) {
	params := &Config{}
	if err := defaults.Set(params); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	var empty string

	// allowed empty value for config file
	if params.App.ConfigFile == nil {
		params.App.ConfigFile = &empty
	}

	// allowed empty value for dns discovery service suffix
	if params.Discovery.DNS.ServiceSuffix == nil {
		params.Discovery.DNS.ServiceSuffix = &empty
	}
	setNeededPaths(params)

	return params, nil
}

// NewConfigENV loads configuration parameters from enviroment
func NewConfigENV(ctx context.Context, prefix string) (*Config, error) {
	params := &Config{}
	if err := envconfig.Process(prefix, params); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}
	setNeededPaths(params)

	return params, nil
}

// NewConfigCLI loads configuration parameters from command-line
func NewConfigCLI(ctx context.Context) (*Config, error) {
	params := &Config{}
	parser := flags.NewParser(params, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if strings.HasPrefix(err.Error(), "Usage") {
			return nil, ErrCLIUsage
		}
		return nil, exterr.WrapWithFrame(err)
	}
	setNeededPaths(params)

	return params, nil
}

// NewConfigYAML loads configuration parameters from YAML file
func NewConfigYAML(ctx context.Context, file string) (*Config, error) {
	params := &Config{}

	if file == "" {
		logging.Info(ctx, "YAML file config skipped")
		return params, nil
	}

	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	d := yaml.NewDecoder(f)
	if err := d.Decode(params); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}
	setNeededPaths(params)

	return params, nil
}

// setNeededPaths sets incoming model/module archive path, model/module base path if these paths aren't set
func setNeededPaths(config *Config) {
	if config.Storage.Filesystem.Base.BasePath != nil {
		// models
		if config.Storage.Filesystem.Model.IncomingArchivePath == nil {
			path := path.Join(*config.Storage.Filesystem.Base.BasePath, incoming, models)
			config.Storage.Filesystem.Model.IncomingArchivePath = &path
		}
		if config.Storage.Filesystem.Model.BasePath == nil {
			path := path.Join(*config.Storage.Filesystem.Base.BasePath, models)
			config.Storage.Filesystem.Model.BasePath = &path
		}

		// modules
		if config.Storage.Filesystem.Module.IncomingArchivePath == nil {
			path := path.Join(*config.Storage.Filesystem.Base.BasePath, incoming, modules)
			config.Storage.Filesystem.Module.IncomingArchivePath = &path
		}
		if config.Storage.Filesystem.Module.BasePath == nil {
			path := path.Join(*config.Storage.Filesystem.Base.BasePath, modules)
			config.Storage.Filesystem.Module.BasePath = &path
		}
	}
}
