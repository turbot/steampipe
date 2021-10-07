load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe cache functionality check ON" {
  run steampipe plugin install chaos
  run steampipe query "select * from chaos.chaos_cache_check where id=0" --output json
  echo $output
  # store the date in `content`
  content=$(echo $output | jq '.[0].time_now')
  echo $content

  run steampipe query "select * from chaos.chaos_cache_check where id=0" --output json
  echo $output
  # store the date in `new_content`
  new_content=$(echo $output | jq '.[0].time_now')
  echo $new_content

  # verify that `content` and `new_content` are the same
  assert_equal "$new_content" "$content"
}

@test "steampipe cache functionality check OFF" {
  run steampipe plugin install chaos
  export STEAMPIPE_CACHE=false
  run steampipe query "select * from chaos.chaos_cache_check where id=0" --output json
  echo $output
  # store the date in `content`
  content=$(echo $output | jq '.[0].time_now')
  echo $content

  run steampipe query "select * from chaos.chaos_cache_check where id=0" --output json
  echo $output
  # store the date in `new_content`
  new_content=$(echo $output | jq '.[0].time_now')
  echo $new_content

  # verify that `content` and `new_content` are not the same
  assert_equal "$new_content" "$content"
}
