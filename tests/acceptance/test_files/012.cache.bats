load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe cache functionality check ON" {
  run steampipe plugin install chaos
  cd $FUNCTIONALITY_TEST_MOD

  run steampipe check benchmark.check_cache_benchmark --export json  --max-parallel 1

  # store the unique number from 1st control in `content`
  content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
  # store the unique number from 2nd control in `new_content`
  new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
  echo $content
  echo $new_content

  # verify that `content` and `new_content` are the same
  assert_equal "$new_content" "$content"
  rm -f benchmark.*.json
}

@test "steampipe cache functionality check OFF" {
  cd $FUNCTIONALITY_TEST_MOD

  # set the env variable to false
  export STEAMPIPE_CACHE=false
  run steampipe check benchmark.check_cache_benchmark --export json  --max-parallel 1

  # store the unique number from 1st control in `content`
  content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
  # store the unique number from 2nd control in `new_content`
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

@test "steampipe cache functionality check ON(check content of results, not just the unique column)" {
  # start service to turn on caching
  steampipe service start

  steampipe query "select unique_col, a, b from chaos_cache_check" --output json &> output1.json
  # store the result from 1st query in `content`
  content=$(cat output1.json)

  steampipe query "select unique_col, a, b from chaos_cache_check" --output json &> output2.json
  # store the result from 2nd query in `new_content`
  new_content=$(cat output2.json)

  echo $content
  echo $new_content

  # stop service
  steampipe service stop

  # verify that `content` and `new_content` are the same
  assert_equal "$new_content" "$content"

  rm -f output1.json
  rm -f output2.json
}
