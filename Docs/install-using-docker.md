# Install TensorFlow Deploy using Docker
Depending on the use of TensorFlow Deploy, two installation scenarios will be covered.

* [TensorFlow Deploy as a model repository](#TensorFlow-Deploy-as-a-model-repository)
* [TensorFlow Deploy as a deployment tool](#TensorFlow-Deploy-as-a-deployment-tool)

### Installation of TensorFlow Deploy communication tools
In order to be able to communicate with the service easily, you can install `tensorflow-deploy-utils` python  package. It's recommended, but still optional.

Package can be installed directly in python

```bash
$ pip install --user tensorflow-deploy-utils
```
or by using virtualenv

```bash
$ python3 -m venv ~/tfd-env
$ source ~/tfd-env/bin/activate
$ pip install tensorflow-deploy-utils
```

## TensorFlow Deploy as a model repository
First, let's assume we would like to use `tfd` only as a repository with no possibility of communication with TensorFlow Serving.

```bash
# Prepare an environment to run tfd
$ mkdir ~/tfdeploy
# Download docker's image
$ docker pull grupawp/tensorflow-deploy
# Run the image
$ docker run --name tfd -v ~/tfdeploy:/tfdeploy -p 9500:9500 grupawp/tensorflow-deploy --storage_filesystem_base_path=/tfdeploy
```
Entering those three easy commands gives us a ready-to-go repository that provides us with a possibility of storing and versioning models as well as modules created by using TensorFlow.

Test `tfd`:

```bash
# Ping tfd
$ curl 0:9500/ping
```

You can find more information about configuration in section [Configure TensorFlow Deploy](configuration.md) and REST API in [Communicate using REST API](api.md).

## TensorFlow Deploy as a deployment tool
However, we may sometimes want to use `tfd` full potential and pair it together with `tfs`.

Prepare enviroment and download needed files from the repository.

```bash
# Prepare an environment to run tfd
$ mkdir ~/tfdeploy

# Download empty configuration for TensorFlow Serving so that it can run without models (by default it requires a model to run and if it is missing one it results in a loop of errors impossible to stop - endpoint reload_config is not working)
$ wget -O ~/tfdeploy/empty.config https://github.com/grupawp/tensorflow-deploy/raw/master/demo/data/empty.config

# Download example model
$ wget https://github.com/grupawp/tensorflow-deploy/raw/master/demo/data/example_model.tar

# Create a dns file storing details of our tfs instance with its IP and GRPC communication port
# Make sure the tfs instance IP address is correct. It may differ depending on your Docker environment and currently running cointainers.
$ echo 'tfs-team1-project1 172.17.0.3:8500' > ~/tfdeploy/hosts
```

Install `tfd` and `tfs`.

```bash
# Download TensorFlow Deploy image
$ docker pull grupawp/tensorflow-deploy
# Run tfd. Optionaly, add -d to run container as a daemon.
$ docker run --name tfd -v ~/tfdeploy:/tfdeploy -p 9500:9500 grupawp/tensorflow-deploy --discovery=plaintext --discovery_plaintext_hosts_path=/tfdeploy/hosts --storage_filesystem_base_path=/tfdeploy

# Download TensorFlow Serving image
$ docker pull tensorflow/serving
# Run tfs. Optionaly, add -d to run container as a daemon.
$ docker run --name tfs-1 -v ~/tfdeploy/models:/models -p 8500:8500 -p 8501:8501 tensorflow/serving --model_config_file=/models/empty.config
```

Test it by using utils

```bash
# Upload model
$ tfd_deploy_model --host localhost --team team1 --project project1 --name name1 --path example_model.tar
# List models
$ tfd_list_models --host localhost
```

or by using REST API

```bash
# Upload model
$ curl -X POST -H "Content-Type: multipart/form-data" -F "archive_data=@example_model.tar" 0:9500/v1/models/team1/project1/names/name1
# Reload models within a team and project
$ curl -X POST 0:9500/v1/models/team1/project1/reload
# Check model on tfs
$ curl 0:8501/v1/models/name1
```

## TensorFlow Deploy for a mass production

### Service discovery configuration
While running multiple projects with at least one TensorFlow Serving instance assigned to each of them, it is difficult to keep DNS text file in order. Especially, when we are using tools like Kubernetes with running `tfs` together with autoscaller on it. In such case, it is worth considering using assistance of an administrator who can configure DNS service on the hardware running TensorFlow Deploy so that each `tfs` instance will have a common suffix, e.g. tfd. Then we can use DNS maintenance options embedded in `tfd`
```bash
$ docker run --name tfd -v ~/tfdeploy:/tfdeploy -p 9500:9500 grupawp/tensorflow-deploy --discovery=dns --discovery_dns_service_suffix=tfd --storage_filesystem_base_path=/tfdeploy
```

### Shared drive storage configuration
Shared drive storage of `tfd` and all instances of `tfs` may be an issue in a mass production environment. The simplest solution is to create a (NFS) network resource and mount it to all the instances.
