#!/usr/bin/env bash

# check version
steampipe -v

# clone the repo, to run the test suite
git clone https://github.com/turbot/steampipe.git
cd steampipe

# initialize git along with bats submodules
git init
git submodule update --init
git submodule update --recursive
git checkout $1
git branch

# declare the test file names
declare -a arr=("migration" "service_and_plugin" "search_path" "chaos_and_query" "dynamic_schema" "cache" "mod_install" "mod" "check" "performance" "exit_codes" "force_stop")

# run test suite
for i in "${arr[@]}"
do
  echo ">>>>> running $i.bats"
  ./tests/acceptance/run.sh $i.bats
done

echo "test run complete"
