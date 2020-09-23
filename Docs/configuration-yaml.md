# YAML File
Description of configuration parameters for YAML file.

File could be used from both [CLI](configuration-cli.md#Application) or [ENVS](configuration-envs.md#Application), using parameters below:
* for CLI: `--config_file`, eg. `./tfd --config_file=config.yaml`
* for ENV: `TFD_CONFIG_FILE`, eg. `export TFD_CONFIG_FILE=config.yaml ./tfd`

## Application
Section contains basic configuration parameters of application.

| Parameter | Description |
|:----------|:------------|
| listenHost | Listen host *(default: 0.0.0.0)* |
| listenPort | Listen port *(default: 9500)* |
| reloadIntervalInSec | The interval of time after which the model configurations will be reloaded on TFS instances *(default: 300)* |
| maxAutoReloadDurationInSec | Max duration auto-reload *(default: 3600)* |
| uploadTimeoutInSec | Timeout after which upload will be interrupted *(default: 300)* |
| defaultModelLabel | Default model label *(default: canary)* |
| tfsAllowLabelsForUnavailableModels | If true, assume TFS instances accept assigning labels to models that are not available yet *(default: false)* |
| discovery | Discovery source, see section of selected Discovery Options *(default: dns)* |
| storage | Storage backend, see section of selected Storage Options *(default: filesystem)* |
| metadata | Metadata backend, see section of selected Metadata Options *(default: sqldb)* |

<br />

## Discovery
Section contains configuration parameters of chosen Discovery which is necessary to get list of available TFS instances.

### DNS
DNS based Discovery.

| Parameter | Description |
|:----------|:------------|
| serviceSuffix | Service suffix with or without dot prefix; optional *(default: not set)* |
| defaultInstancePort | Default TFS instance port in case SRV records are unavailable *(default: 8500)* |

### Plaintext
Text file based Discovery.

| Parameter | Description |
|:----------|:------------|
| hostsPath | Path to the file containing configuration of the TFS instances *(default: /tfdeploy/hosts)* |

<br />

## Storage
Section contains configuration parameters of chosen kind of models/modules store.

### Filesystem
File System based storage.

#### Base

| Parameter | Description |
|:----------|:------------|
| basePath | Root location *(default: /tfdeploy)* |

#### Model

| Parameter | Description |
|:----------|:------------|
| archiveName | Model archive name *(default: model_archive.tar)* |
| configName | Models config filename *(default: models.config)* |
| emptyConfigName | Empty config filename *(default: empty.config)* |

#### Module

| Parameter | Description |
|:----------|:------------|
| archiveName | Module archive name *(default: module_archive.tar)* |

<br />

## Metadata
Section contains configuration parameters of chosen database implementation used to hold necessary metadata which is required for proper work of application.

### SQLDB
SQL based storage.

| Parameter | Description |
|:----------|:------------|
| driver | SQL driver *(default: sqlite3)* |
| dsn | The Data Source Name in common format like e.g. PEAR DB, but without type-prefix (optional parts marked by squared brackets) *(default: /tfdeploy/tfdeploy.db)* |

<br />

## Example Configuration File

```yaml
application:
    listenHost: '0.0.0.0'
    listenPort: 9500
    reloadIntervalInSec: 300
    maxAutoReloadDurationInSec: 900
    uploadTimeoutInSec: 300
    defaultModelLabel: 'canary'
    tfsAllowLabelsForUnavailableModels: false
    discovery: 'plaintext'
    storage: 'filesystem'
    metadata: 'sqldb'

discovery:
    dns:
        serviceSuffix: ''
        defaultInstancePort: 8500
    plaintext:
        hostsPath: '/tfdeploy/hosts'

storage:
    filesystem:
        base:
            basePath: '/tfdeploy'
        model:
            archiveName: 'model_archive.tar'
            configName: 'models.config'
            emptyConfigName: 'empty.config'
        module:
            archiveName: 'module_archive.tar'

metadata:
    sqldb:
        driver: 'sqlite3'
        dsn: '/tfdeploy/tfdeploy.db'
```
