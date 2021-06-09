load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe check all" {
    run steampipe check all --workspace $WORKSPACE_DIR
    assert_success
}

@test "steampipe check all - output csv" {
  run steampipe check all --workspace $WORKSPACE_DIR --output=csv --progress=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_csv.csv)"
}

@test "steampipe check all - output json" {
  run steampipe check all --workspace $WORKSPACE_DIR --output=json --progress=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_json.json)"
}

@test "steampipe check all - export csv" {
  run steampipe check all --workspace $WORKSPACE_DIR --export=csv:./test.csv --progress=false
  assert_equal "$(cat ./test.csv)" "$(cat $TEST_DATA_DIR/expected_check_csv.csv)"
  rm -f ./test.csv
}

@test "steampipe check all - export json" {
  run steampipe check all --workspace $WORKSPACE_DIR --export=json:./test.json --progress=false
  assert_equal "$(cat ./test.json)" "$(cat $TEST_DATA_DIR/expected_check_json.json)"
  rm -f ./test.json
}

