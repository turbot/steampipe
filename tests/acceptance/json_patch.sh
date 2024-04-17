#!/bin/bash -e

# This script accepts a patch format and evaluates the diffs if any.
patch_file=$1

patch_keys=$(echo $patch_file | jq -r '. | keys[]')

for i in $patch_keys; do
  op=$(echo $patch_file | jq -r -c ".[${i}]" | jq -r ".op")
  path=$(echo $patch_file | jq -r -c ".[${i}]" | jq -r ".path")
  value=$(echo $patch_file | jq -r -c ".[${i}]" | jq -r ".value")

  # ignore the diff of paths 'end_time', 'start_time' and 'schema_version',
  # print the rest
  if [[ $op != "test" ]] && [[ $path != "/end_time" ]] && [[ $path != "/start_time" ]] && [[ $path != "/schema_version" ]] && [[ $path != "/metadata"* ]]; then
    if [[ $op == "remove" ]]; then
      echo "key: $path"
      echo "expected: $value"
    else
      echo "actual: $value"
    fi
  fi
done