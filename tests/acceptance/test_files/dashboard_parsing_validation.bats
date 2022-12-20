load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# TODO rename tests properly
@test "Parsing case 1 - PASS" {
  cd $FILE_PATH/test_data/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_top_level --output snapshot
  echo $output

  assert_success
}

