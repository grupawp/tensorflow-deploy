# ENVS
Description of configuration parameters for environments.

## Application
Section contains basic configuration parameters of application.

| Parameter | Description |
|:----------|:------------|
| TFD_CONFIG_FILE | Path to the config file *(default: not set)* |
| TFD_LISTEN_HOST | Listen host *(default: 0.0.0.0)* |
| TFD_LISTEN_PORT | Listen port *(default: 9500)* |
| TFD_RELOAD_INTERVAL_IN_SEC | The interval of time after which the model configurations will be reloaded on TFS instances *(default: 300)* |
| TFD_MAX_AUTO_RELOAD_DURATION_IN_SEC | Max duration auto-reload *(default: 3600)* |
| TFD_UPLOAD_TIMEOUT_IN_SEC | Timeout after which upload will be interrupted *(default: 300)* |
| TFD_DEFAULT_MODEL_LABEL | Default model label *(default: canary)* |
| TFD_TFS_ALLOW_LABELS_FOR_UNAVAILABLE_MODELS | If true, assume TFS instances accept assigning labels to models that are not available yet *(default: false)* |
| TFD_DISCOVERY | Discovery source, see section of selected Discovery Options *(default: dns)* |
| TFD_STORAGE | Storage backend, see section of selected Storage Options *(default: filesystem)* |
| TFD_METADATA | Metadata backend, see section of selected Metadata Options *(default: sqldb)* |

<br />

## Discovery
Section contains configuration parameters of chosen Discovery which is necessary to get list of available TFS instances.

### DNS
DNS based Discovery.

| Parameter | Description |
|:----------|:------------|
| TFD_DISCOVERY_DNS_SERVICE_SUFFIX | Service suffix with or without dot prefix; optional *(default: not set)* |
| TFD_DISCOVERY_DNS_DEFAULT_INSTANCE_PORT | Default TFS instance port in case SRV records are unavailable *(default: 8500)* |

### Plaintext
Text file based Discovery.

| Parameter | Description |
|:----------|:------------|
| TFD_DISCOVERY_PLAINTEXT_HOSTS_PATH | Path to the file containing configuration of the TFS instances *(default: /tfdeploy/hosts)* |

<br />

## Storage
Section contains configuration parameters of chosen kind of models/modules store.

### Filesystem
File System based storage.

#### Base

| Parameter | Description |
|:----------|:------------|
| TFD_STORAGE_FILESYSTEM_BASE_PATH | Root location *(default: /tfdeploy)* |

#### Model

| Parameter | Description |
|:----------|:------------|
| TFD_STORAGE_FILESYSTEM_MODEL_ARCHIVE_NAME | Model archive name *(default: model_archive.tar)* |
| TFD_STORAGE_FILESYSTEM_MODEL_CONFIG_NAME | Models config filename *(default: models.config)* |
| TFD_STORAGE_FILESYSTEM_MODEL_EMPTY_CONFIG_NAME | Empty config filename *(default: empty.config)* |

#### Module

| Parameter | Description |
|:----------|:------------|
| TFD_STORAGE_FILESYSTEM_MODULE_ARCHIVE_NAME | Module archive name *(default: module_archive.tar)* |

<br />

## Metadata
Section contains configuration parameters of chosen database implementation used to hold necessary metadata which is required for proper work of application.

### SQLDB
SQL based storage.

| Parameter | Description |
|:----------|:------------|
| TFD_METADATA_SQLDB_DRIVER | SQL driver *(default: sqlite3)* |
| TFD_METADATA_SQLDB_DSN | The Data Source Name in common format like e.g. PEAR DB, but without type-prefix (optional parts marked by squared brackets) *(default: /tfdeploy/tfdeploy.db)* |

<br/>

## Example ENVS With Defaults

```bash
# application
export TFD_CONFIG_FILE=
export TFD_LISTEN_HOST=0.0.0.0
export TFD_LISTEN_PORT=9500
export TFD_RELOAD_INTERVAL_IN_SEC=300
export TFD_MAX_AUTO_RELOAD_DURATION_IN_SEC=3600
export TFD_UPLOAD_TIMEOUT_IN_SEC=300
export TFD_DEFAULT_MODEL_LABEL=canary
export TFD_TFS_ALLOW_LABELS_FOR_UNAVAILABLE_MODELS=false
export TFD_DISCOVERY=dns
export TFD_STORAGE=filesystem
export TFD_METADATA=sqldb

# discovery
export TFD_DISCOVERY_PLAINTEXT_HOSTS_PATH=/tfdeploy/hosts
export TFD_DISCOVERY_DNS_SERVICE_SUFFIX=
export TFD_DISCOVERY_DNS_DEFAULT_INSTANCE_PORT=8500

# storage
export TFD_STORAGE_FILESYSTEM_BASE_PATH=/tfdeploy
export TFD_STORAGE_FILESYSTEM_MODEL_ARCHIVE_NAME=model_archive.tar
export TFD_STORAGE_FILESYSTEM_MODEL_CONFIG_NAME=models.config
export TFD_STORAGE_FILESYSTEM_MODEL_EMPTY_CONFIG_NAME=empty.config
export TFD_STORAGE_FILESYSTEM_MODULE_ARCHIVE_NAME=module_archive.tar
export TFD_STORAGE_FILESYSTEM_MODULE_BASE_PATH=/tfdeploy/modules

# metadata
export TFD_METADATA_SQLDB_DRIVER=sqlite3
export TFD_METADATA_SQLDB_DSN=/tfdeploy/tfdeploy.db
```
