load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe cache functionality check" {
  cd $WORKSPACE_DIR
  run steampipe check benchmark.check_cache_benchmark --output json
  echo $output
  # store the date in the resource field in `content`
  content=$(cat $output | jq '.groups[0].controls[0].results[0].resource')
  echo $content

  run steampipe check benchmark.check_cache_benchmark --output json
  echo $output
  # store the date in the resource field in `new_content`
  new_content=$(cat $output | jq '.groups[0].controls[0].results[0].resource')
  echo $new_content

  # verify that `content` and `new_content` are the same
  assert_equal "$new_content" "$content"
}