#!/bin/sh

CONFIG_FILE="/usr/local/etc/config.yaml"

/bootstrap --config_file $CONFIG_FILE
exec /tensorflow_deploy --config_file $CONFIG_FILE "$@"
