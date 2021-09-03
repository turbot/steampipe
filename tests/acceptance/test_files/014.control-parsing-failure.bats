load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"


@test "control with neither query property nor sql property" {
  cd $BAD_TEST_MOD_DIR
  run steampipe check control.control_fail_with_no_query_no_sql --output json

  # store the results field in `content`
  content=$(cat output.json | jq '.controls[0].results')

  # should return an error `must define either a 'sql' property or a 'query' property`,
  # so the results should be empty
  assert_equal "$content" ""
}

@test "control with both query property and sql property" {
  cd $BAD_TEST_MOD_DIR
  run steampipe check control.control_fail_with_both_query_and_sql --output json

  # store the results field in `content`
  content=$(cat output.json | jq '.controls[0].results')

  # should return an error `must define either a 'sql' property or a 'query' property`,
  # so the results should be empty
  assert_equal "$content" ""
}

@test "control with both params property and query property" {
  cd $BAD_TEST_MOD_DIR
  run steampipe check control.control_fail_with_params_and_query --output json

  # store the results field in `content`
  content=$(cat output.json | jq '.controls[0].results')

  # should return an error `has 'query' property set so cannot define param blocks`,
  # so the results should be empty
  assert_equal "$content" ""
}