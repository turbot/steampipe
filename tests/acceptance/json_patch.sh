#!/bin/bash -e

# This script accepts a patch format and evaluates the diffs if any.
patch_file=$1

patch_keys=$(echo $patch_file | jq '. | keys[]')

for i in $patch_keys; do
  op=$(echo $patch_file | jq -c ".[${i}]" | jq ".op")
  path=$(echo $patch_file | jq -c ".[${i}]" | jq ".path")
  value=$(echo $patch_file | jq -c ".[${i}]" | jq ".value")

  # ignore the diff of paths 'end_time', 'start_time' and 'schema_version',
  # print the rest
  if [[ $op != '"test"' ]] && [[ $path != '"/end_time"' ]] && [[ $path != '"/start_time"' ]] && [[ $path != '"/schema_version"' ]]; then
    if [[ $op == '"remove"' ]]; then
      echo "key: $path"
      echo "expected: $value"
    else
      echo "actual: $value"
    fi
  fi
done