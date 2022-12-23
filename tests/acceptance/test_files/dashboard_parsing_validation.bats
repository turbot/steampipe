load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# TODO rename tests properly

@test "Parsing case 1 - top level query providers do not require query/sql blocks (PASS)" {
  cd $FILE_PATH/test_data/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_top_level --output snapshot
  assert_success
}

@test "Parsing case 2 - top level query providers do not require query/sql blocks except control/query (PASS)" {
  cd $FILE_PATH/test_data/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_top_level_require_sql --output snapshot
  assert_success
}

@test "Parsing case 3 - nested query providers do require query/sql blocks (PASS)" {
  cd $FILE_PATH/test_data/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_nested --output snapshot
  assert_success
}
