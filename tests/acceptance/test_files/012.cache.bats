load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe cache functionality check" {
  cd $WORKSPACE_DIR
  run steampipe check benchmark.check_cache_benchmark --output json
  content=$(cat $output | jq '.groups[0].controls[0].results[0].resource')
  echo $content
  run steampipe check benchmark.check_cache_benchmark --output json
  new_content=$(cat $output | jq '.groups[0].controls[0].results[0].resource')
  echo $new_content
  assert_equal "$new_content" "$content"
}