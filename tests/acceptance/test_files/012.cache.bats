load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe cache functionality check ON" {
  run steampipe plugin install chaos
  cd $FUNCTIONALITY_TEST_MOD

  run steampipe check benchmark.check_cache_benchmark --export json  --max-parallel 1

  # store the date from 1st control in `content`
  content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
  # store the date from 2nd control in `new_content`
  new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
  echo $content
  echo $new_content

  # verify that `content` and `new_content` are the same
  assert_equal "$new_content" "$content"
  rm -f benchmark.*.json
}

@test "steampipe cache functionality check OFF" {
  run steampipe plugin install chaos
  cd $FUNCTIONALITY_TEST_MOD

  # set the env variable to false
  export STEAMPIPE_CACHE=false
  run steampipe check benchmark.check_cache_benchmark --export json  --max-parallel 1

  # store the date from 1st control in `content`
  content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
  # store the date from 2nd control in `new_content`
  new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
  echo $content
  echo $new_content

  # verify that `content` and `new_content` are not the same
  if [[ "$content" == "$new_content" ]]; then
    flag=1
  else
    flag=0
  fi
  assert_equal "$flag" "0"
  rm -f benchmark.*.json
}
