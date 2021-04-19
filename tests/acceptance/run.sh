#!/bin/bash -e

if [[ ! ${MY_PATH} ]];
then
  MY_PATH="`dirname \"$0\"`"              # relative
  MY_PATH="`( cd \"$MY_PATH\" && pwd )`"  # absolutized and normalized  
fi

# set this to the source file for development
export BATS_PATH=$MY_PATH/lib/bats/bin/bats
export LIB_BATS_ASSERT=$MY_PATH/lib/bats-assert
export LIB_BATS_SUPPORT=$MY_PATH/lib/bats-support
export TEST_DATA_DIR=$MY_PATH/test_data/templates
export TEST_SRC_DIR=$MY_PATH/test_data/source_files

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

export PATH=$PATH:$MY_PATH/lib/bats/bin

if [[ ! ${STEAMPIPE_INSTALL_DIR} ]];
then
  export STEAMPIPE_INSTALL_DIR="~/.steampipe"
fi

echo "Running with STEAMPIPE_INSTALL_DIR set to $STEAMPIPE_INSTALL_DIR"

bats --tap $MY_PATH/test_files