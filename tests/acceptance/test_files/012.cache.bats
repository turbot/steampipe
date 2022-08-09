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

  steampipe query "select case when mod(id,2)=0 then 'alarm' when mod(id,2)=1 then 'ok' end status, unique_col as resource, id as reason from chaos.chaos_cache_check where id=2" --output json > query_result1.json

  steampipe query "select case when mod(id,2)=0 then 'alarm' when mod(id,2)=1 then 'ok' end status, unique_col as resource, id as reason from chaos.chaos_cache_check where id=2" --output json > query_result2.json

  # stop the service
  steampipe service stop

  # both the results should be same
  assert_equal "$(cat query_result1.json)" "$(cat query_result2.json)"

  rm -f query_result1.json
  rm -f query_result2.json
}
