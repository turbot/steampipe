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
declare -a arr=("migration" "service_and_plugin" "search_path" "chaos_and_query" "dynamic_schema" "cache" "mod_install" "mod" "check" "workspace" "cloud" "performance" "exit_codes")
declare -i failure_count=0

# run test suite
for i in "${arr[@]}"
do
  echo ""
  echo ">>>>> running $i.bats"
  ./tests/acceptance/run.sh $i.bats
  failure_count+=$?
done

# check if all tests passed
echo $failure_count
if [[ $failure_count -eq 0 ]]; then
  echo "test run successful"
  exit 0
else
  echo "test run failed"
  exit 1
fi
