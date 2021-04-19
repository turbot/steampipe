#!/bin/bash -e

MY_PATH="`dirname \"$0\"`"              # relative
MY_PATH="`( cd \"$MY_PATH\" && pwd )`"  # absolutized and normalized

export STEAMPIPE_INSTALL_DIR=$(mktemp -d)
trap "code=$?;rm -rf $STEAMPIPE_INSTALL_DIR; exit $code" EXIT

source $MY_PATH/run.sh
