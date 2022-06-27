load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe check search_path_prefix when passed through command line" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.search_path_test_1 --output json --search-path-prefix aws --export json
  assert_equal "$(cat control.*.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f control.*.json
}

@test "steampipe check search_path when passed through command line" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.search_path_test_2 --output json --search-path chaos,b,c --export json
  assert_equal "$(cat control.*.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f control.*.json
}

@test "steampipe check search_path and search_path_prefix when passed through command line" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.search_path_test_3 --output json --search-path chaos,b,c --search-path-prefix aws --export json
  assert_equal "$(cat control.*.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f control.*.json
}

@test "steampipe check search_path_prefix when passed in the control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.search_path_test_4 --output json --export json
  assert_equal "$(cat control.*.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f control.*.json
}

@test "steampipe check search_path when passed in the control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.search_path_test_5 --output json --export json
  assert_equal "$(cat control.*.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f control.*.json
}

@test "steampipe check search_path and search_path_prefix when passed in the control" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check control.search_path_test_6 --output json --export json
  assert_equal "$(cat control.*.json | jq '.controls[0].results[0].status')" '"ok"'
  rm -f control.*.json
}
