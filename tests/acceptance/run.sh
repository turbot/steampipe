#!/bin/bash -e

if [[ ! ${MY_PATH} ]];
then
  MY_PATH="`dirname \"$0\"`"              # relative
  MY_PATH="`( cd \"$MY_PATH\" && pwd )`"  # absolutized and normalized
fi

if [[ ! ${TIME_TO_QUERY} ]];
then
  TIME_TO_QUERY=2
fi

# set this to the source file for development
export BATS_PATH=$MY_PATH/lib/bats/bin/bats
export LIB_BATS_ASSERT=$MY_PATH/lib/bats-assert
export LIB_BATS_SUPPORT=$MY_PATH/lib/bats-support
export TEST_DATA_DIR=$MY_PATH/test_data/templates
export SRC_DATA_DIR=$MY_PATH/test_data/source_files
export WORKSPACE_DIR=$MY_PATH/test_data/sample_workspace
export BAD_TEST_MOD_DIR=$MY_PATH/test_data/failure_test_mod
export TIME_TO_QUERY=$TIME_TO_QUERY
export SIMPLE_MOD_DIR=$MY_PATH/test_data/introspection_table_mod
export CONFIG_PARSING_TEST_MOD=$MY_PATH/test_data/config_parsing_test_mod
export FILE_PATH=$MY_PATH
export CHECK_ALL_MOD=$MY_PATH/test_data/check_all_mod
export FUNCTIONALITY_TEST_MOD=$MY_PATH/test_data/functionality_test_mod
export CONTROL_RENDERING_TEST_MOD=$MY_PATH/test_data/control_rendering_test_mod
export STEAMPIPE_CONNECTION_WATCHER=false

# Must have these commands for the test suite to run
declare -a required_commands=("jq" "sed" "steampipe" "rm" "mv" "cp" "mkdir" "cd" "head" "wc" "find" "basename" "dirname" "touch")

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

export PATH=$MY_PATH/lib/bats/bin:$PATH

if [[ ! ${STEAMPIPE_INSTALL_DIR} ]];
then
  export STEAMPIPE_INSTALL_DIR="$HOME/.steampipe"
fi

echo "Running with STEAMPIPE_INSTALL_DIR set to $STEAMPIPE_INSTALL_DIR"

bats --tap $MY_PATH/test_files

# Setting the exit_code, to use in the github workflow(This only gets set to 0 when the above bats test suite passes)
echo "::set-output name=exit_code::$(echo $?)"