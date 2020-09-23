# CLI
CLI description of configuration parameters.

Help command is available under:
* `tfd -h`
* `tfd --help`

## Application
Section contains basic configuration parameters of application.

| Parameter | Description |
|:----------|:------------|
| --config_file | Path to the config file *(default: not set)* |
| --listen_host | Listen host *(default: 0.0.0.0)* |
| --listen_port | Listen port *(default: 9500)* |
| --reload_interval_in_sec | The interval of time after which the model configurations will be reloaded on TFS instances *(default: 300)* |
| --max_auto_reload_duration_in_sec | Max duration auto-reload *(default: 3600)* |
| --upload_timeout_in_sec | Timeout after which upload will be interrupted *(default: 300)* |
| --default_model_label | Default model label *(default: canary)* |
| --tfs_allow_labels_for_unavailable_models | If true, assume TFS instances accept assigning labels to models that are not available yet *(default: false)* |
| --discovery | Discovery source, see section of selected Discovery Options *(default: dns)* |
| --storage | Storage backend, see section of selected Storage Options *(default: filesystem)* |
| --metadata | Metadata backend, see section of selected Metadata Options *(default: sqldb)* |

<br />

## Discovery
Section contains configuration parameters of chosen Discovery which is necessary to get list of available TFS instances.

### DNS
DNS based Discovery.

| Parameter | Description |
|:----------|:------------|
| --discovery_dns_service_suffix | Service suffix with or without dot prefix; optional *(default: not set)* |
| --discovery_dns_default_instance_port | Default TFS instance port in case SRV records are unavailable *(default: 8500)* |

### Plaintext
Text file based Discovery.

| Parameter | Description |
|:----------|:------------|
| --discovery_plaintext_hosts_path | Path to the file containing configuration of the TFS instances *(default: /tfdeploy/hosts)* |

<br />

## Storage
Section contains configuration parameters of chosen kind of models/modules store.

### Filesystem
File System based storage.

#### Base

| Parameter | Description |
|:----------|:------------|
| --storage_filesystem_base_path | Root location (default: /tfdeploy)* |

#### Model

| Parameter | Description |
|:----------|:------------|
| --storage_filesystem_model_archive_name | Model archive name *(default: model_archive.tar)* |
| --storage_filesystem_model_config_name | Models config filename *(default: models.config)* |
| --storage_filesystem_model_empty_config_name | Empty config filename *(default: empty.config)* |

#### Module

| Parameter | Description |
|:----------|:------------|
| --storage_filesystem_module_archive_name | Module archive name *(default: module_archive.tar)* |

<br />

## Metadata
Section contains configuration parameters of chosen database implementation used to hold necessary metadata which is required for proper work of application.

### SQLDB
SQL based storage.

| Parameter | Description |
|:----------|:------------|
| --metadata_sqldb_driver | SQL driver *(default: sqlite3)* |
| --metadata_sqldb_dsn | The Data Source Name in common format like e.g. PEAR DB, but without type-prefix (optional parts marked by squared brackets) *(default: /tfdeploy/tfdeploy.db)* |
