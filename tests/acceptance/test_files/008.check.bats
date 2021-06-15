load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe check all" {
  cd $WORKSPACE_DIR
  run steampipe check all --output=none --progress=false
  assert_success
  cd -
}

@test "steampipe check all - output csv" {
  cd $WORKSPACE_DIR
  run steampipe check all --output=csv --progress=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_csv.csv)"
  cd -
}

@test "steampipe check all - output json" {
  cd $WORKSPACE_DIR
  run steampipe check all --output=json --progress=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_json.json)"
  cd -
}

@test "steampipe check all - export csv" {
  cd $WORKSPACE_DIR
  run steampipe check all --export=csv:./test.csv --progress=false
  assert_equal "$(cat ./test.csv)" "$(cat $TEST_DATA_DIR/expected_check_csv.csv)"
  rm -f ./test.csv
  cd -
}

@test "steampipe check all - export json" {
  cd $WORKSPACE_DIR
  run steampipe check all --export=json:./test.json --progress=false
  assert_equal "$(cat ./test.json)" "$(cat $TEST_DATA_DIR/expected_check_json.json)"
  rm -f ./test.json
  cd -
}

