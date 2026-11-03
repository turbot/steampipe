#!/bin/bash -e

if [[ ! ${MY_PATH} ]];
then
  MY_PATH="`dirname \"$0\"`"              # relative
  MY_PATH="`( cd \"$MY_PATH\" && pwd )`"  # absolutized and normalized
fi

if [[ ! ${TIME_TO_QUERY} ]];
then
  TIME_TO_QUERY=4
fi

# set this to the source file for development
export BATS_PATH=$MY_PATH/lib/bats-core/bin/bats
export LIB=$MY_PATH/lib
export LIB_BATS_ASSERT=$LIB/bats-assert
export LIB_BATS_SUPPORT=$LIB/bats-support
export TEST_DATA_DIR=$MY_PATH/test_data/templates
export SRC_DATA_DIR=$MY_PATH/test_data/source_files
export TIME_TO_QUERY=$TIME_TO_QUERY
export FILE_PATH=$MY_PATH
export STEAMPIPE_CONNECTION_WATCHER=false
export STEAMPIPE_INTROSPECTION=info
# from GH action env variables
export SPIPETOOLS_PG_CONN_STRING=$SPIPETOOLS_PG_CONN_STRING
export SPIPETOOLS_TOKEN=$SPIPETOOLS_TOKEN
# Disable parallelisation only within test file(for steampipe plugin manager processes to shutdown properly)
export BATS_NO_PARALLELIZE_WITHIN_FILE=true
export BATS_TEST_TIMEOUT=180

# Must have these commands for the test suite to run
declare -a required_commands=("jq" "sed" "steampipe" "rm" "mv" "cp" "mkdir" "cd" "head" "wc" "find" "basename" "dirname" "touch" "jd" "openssl" "cksum")

for required_command in "${required_commands[@]}"
do
  if [[ $(command -v $required_command | head -c1 | wc -c) -eq 0 ]]; then
    echo "$required_command is required for this test suite to run."
    exit -1
  fi
done

echo " ____  _             _   _               _____         _       "
echo "/ ___|| |_ __ _ _ __| |_(_)_ __   __ _  |_   _|__  ___| |_ ___ "
echo "\___ \| __/ _\` | '__| __| | '_ \ / _\` |   | |/ _ \/ __| __/ __|"
echo " ___) | || (_| | |  | |_| | | | | (_| |   | |  __/\__ \ |_\__ \\"
echo "|____/ \__\__,_|_|   \__|_|_| |_|\__, |   |_|\___||___/\__|___/"
echo "                                 |___/                         "

export PATH=$MY_PATH/lib/bats-core/bin:$PATH

if [[ ! ${STEAMPIPE_INSTALL_DIR} ]];
then
  export STEAMPIPE_INSTALL_DIR="$HOME/.steampipe"
fi

batversion=$(bats --version)
echo $batversion
echo "Running with STEAMPIPE_INSTALL_DIR set to: $STEAMPIPE_INSTALL_DIR"
echo "Running with binary from: $(which steampipe)"

if [ $# -eq 0 ]; then
  # Run all test files
  bats --tap $MY_PATH/test_files
else
  # Run a single test file
  bats --tap $MY_PATH/test_files/${1}
fi
