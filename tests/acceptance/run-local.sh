#!/bin/bash -e

MY_PATH="`dirname \"$0\"`"              # relative
MY_PATH="`( cd \"$MY_PATH\" && pwd )`"  # absolutized and normalized

export STEAMPIPE_INSTALL_DIR=$(mktemp -d)
export TIME_TO_QUERY=3                  # overriding since it takes more than 2secs to run locally
export TZ=UTC
export WD=$(mktemp -d)

trap "cd -;code=$?;rm -rf $STEAMPIPE_INSTALL_DIR; exit $code" EXIT

cd $WD
echo "Working directory: $WD"
# setup a steampipe installation
echo "Install directory: $STEAMPIPE_INSTALL_DIR"
# steampipe query "select 1 as setup_complete"
echo "Installation complete at $STEAMPIPE_INSTALL_DIR"
echo "Installing CHAOS"
# steampipe plugin install chaos
echo "Installed CHAOS"

if [ $# -eq 0 ]; then
  # Run all test files
  $MY_PATH/run.sh
else
  $MY_PATH/run.sh ${1}
fi
