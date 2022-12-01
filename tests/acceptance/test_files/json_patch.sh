#!/bin/bash -e

patch_file=$1

patch_keys=$(echo $patch_file | jq '. | keys[]')

for i in $patch_keys; do
  op=$(echo $patch_file | jq -c ".[${i}]" | jq ".op")
  path=$(echo $patch_file | jq -c ".[${i}]" | jq ".path")
  value=$(echo $patch_file | jq -c ".[${i}]" | jq ".value")

  if [[ $op != '"test"' ]] && [[ $path != '"/end_time"' ]] && [[ $path != '"/start_time"' ]] && [[ $path != *'"/search_path'* ]]; then
    if [[ $op == '"remove"' ]]; then
      echo "key: $path"
      echo "expected: $value"
    else
      echo "actual: $value"
    fi
  fi
done