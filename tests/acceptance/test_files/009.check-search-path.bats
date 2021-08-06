load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe check search_path_prefix when passed through command line" {
  cd $WORKSPACE_DIR
  run steampipe check control.search_path_test_1 --output json --search-path-prefix aws --export output.json
  assert_equal "$(cat output.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f output.json
}

@test "steampipe check search_path when passed through command line" {
  cd $WORKSPACE_DIR
  run steampipe check control.search_path_test_2 --output json --search-path a,b,c --export output.json
  assert_equal "$(cat output.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f output.json
}

@test "steampipe check search_path and search_path_prefix when passed through command line" {
  cd $WORKSPACE_DIR
  run steampipe check control.search_path_test_3 --output json --search-path a,b,c --search-path-prefix aws --export output.json
  assert_equal "$(cat output.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f output.json
}

@test "steampipe check search_path_prefix when passed in the control" {
  cd $WORKSPACE_DIR
  run steampipe check control.search_path_test_4 --output json --export output.json
  assert_equal "$(cat output.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f output.json
}

@test "steampipe check search_path when passed in the control" {
  cd $WORKSPACE_DIR
  run steampipe check control.search_path_test_5 --output json --export output.json
  assert_equal "$(cat output.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f output.json
}

@test "steampipe check search_path and search_path_prefix when passed in the control" {
  cd $WORKSPACE_DIR
  run steampipe check control.search_path_test_6 --output json --export output.json
  assert_equal "$(cat output.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f output.json
}
