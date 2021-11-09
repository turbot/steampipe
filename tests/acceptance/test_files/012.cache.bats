load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "check cache functionality when querying same columns" {
  run steampipe plugin install chaos
  cd $FUNCTIONALITY_TEST_MOD

  run steampipe check benchmark.check_cache_same_columns_benchmark --export=output.json  --max-parallel 1

  # store the date from 1st control in `content`
  content=$(cat output.json | jq '.groups[].controls[0].results[0].resource')
  # store the date from 2nd control in `new_content`
  new_content=$(cat output.json | jq '.groups[].controls[1].results[0].resource')
  echo $content
  echo $new_content

  # verify that `content` and `new_content` are the same
  assert_equal "$new_content" "$content"
  rm -f output.json
}

@test "check cache functionality when the second query's columns is a subset of the first" {
  run steampipe plugin install chaos
  cd $FUNCTIONALITY_TEST_MOD

  run steampipe check benchmark.check_cache_subset_columns_benchmark --export=output.json  --max-parallel 1

  # store the date from 1st control in `content`
  content=$(cat output.json | jq '.groups[].controls[0].results[0].resource')
  # store the date from 2nd control in `new_content`
  new_content=$(cat output.json | jq '.groups[].controls[1].results[0].resource')
  echo $content
  echo $new_content

  # verify that `content` and `new_content` are the same
  assert_equal "$new_content" "$new_content"
  rm -f output.json
}

@test "check cache functionality multiple queries with same columns" {
  run steampipe plugin install chaos
  cd $FUNCTIONALITY_TEST_MOD

  run steampipe check benchmark.check_cache_multiple_same_columns_benchmark --export=output.json  --max-parallel 1

  # store the date from 1st control in `content`
  content=$(cat output.json | jq '.groups[].controls[0].results[0].resource')
  # store the date from 2nd control in `content2`
  content2=$(cat output.json | jq '.groups[].controls[1].results[0].resource')
  # store the date from 3rd control in `content3`
  content3=$(cat output.json | jq '.groups[].controls[2].results[0].resource')
  # store the date from 4th control in `content4`
  content4=$(cat output.json | jq '.groups[].controls[3].results[0].resource')
  echo $content
  echo $content2
  echo $content3
  echo $content4

  # verify that `content`, `content2`, `content3` and `content4` are the same
  assert_equal "$content2" "$content"
  assert_equal "$content3" "$content"
  assert_equal "$content4" "$content"

  rm -f output.json
}

@test "check cache functionality when multiple query's columns are a subset of the first" {
  run steampipe plugin install chaos
  cd $FUNCTIONALITY_TEST_MOD

  run steampipe check benchmark.check_cache_multiple_subset_columns_benchmark --export=output.json  --max-parallel 1

  # store the date from 1st control in `content`
  content=$(cat output.json | jq '.groups[].controls[0].results[0].resource')
  # store the date from 2nd control in `content2`
  content2=$(cat output.json | jq '.groups[].controls[1].results[0].resource')
  # store the date from 3rd control in `content3`
  content3=$(cat output.json | jq '.groups[].controls[2].results[0].resource')
  # store the date from 4th control in `content4`
  content4=$(cat output.json | jq '.groups[].controls[3].results[0].resource')
  echo $content
  echo $content2
  echo $content3
  echo $content4

  # verify that `content`, `content2`, `content3` and `content4` are the same
  assert_equal "$content2" "$content"
  assert_equal "$content3" "$content"
  assert_equal "$content4" "$content"

  rm -f output.json
}

@test "steampipe cache functionality check OFF" {
  run steampipe plugin install chaos
  cd $FUNCTIONALITY_TEST_MOD

  # set the env variable to false
  export STEAMPIPE_CACHE=false
  run steampipe check benchmark.check_cache_same_columns_benchmark --export=output.json  --max-parallel 1

  # store the date from 1st control in `content`
  content=$(cat output.json | jq '.groups[].controls[0].results[0].resource')
  # store the date from 2nd control in `new_content`
  new_content=$(cat output.json | jq '.groups[].controls[1].results[0].resource')
  echo $content
  echo $new_content

  # verify that `content` and `new_content` are not the same
  if [[ "$content" == "$new_content" ]]; then
    flag=1
  else
    flag=0
  fi
  assert_equal "$flag" "0"
  
  rm -f output.json
}
