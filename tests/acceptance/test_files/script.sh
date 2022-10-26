#!/bin/bash -e

cd tests/acceptance/test_files/
tests=$(cat workspace_tests.json)
  # echo $tests

echo $tests | jq -c -r '.[]' | while read i; do
  test_name=$(echo $i | jq '.test')
  echo ">>> Running: $test_name <<<"

  # exports needed for setup
  exports=$(echo $i | jq '.setup.exports')
  echo $exports

  for exp in $(echo "${exports}" | jq -r '.[]'); do
    export "$exp"
  done

  # args to run with steampipe query command
  args=$(echo $i | jq '.setup.args')
  args=$(echo $args | tr -d '"')
  echo $args
  # $args
  diagnostics=$(steampipe query "select 1" "$args")
  echo $diagnostics

done


