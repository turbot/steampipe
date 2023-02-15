load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# These set of tests are skipped locally
# To run these tests locally set the SPIPETOOLS_TOKEN env var.
# These tests will be skipped locally unless the above env var is set.

function setup() {
  if [[ -z "${SPIPETOOLS_TOKEN}" ]]; then
    skip
  else
    echo "SPIPETOOLS_TOKEN is set..."
  fi
}

# These set of tests check the different types of output in query snapshot mode and not snapshot creation/upload
# Related to https://github.com/turbot/steampipe/issues/3112

@test "snapshot mode - query output csv" {
  cd $FILE_PATH/test_data/functionality_test_mod

  steampipe query query.static_query_2 --snapshot --output csv --cloud-token $SPIPETOOLS_TOKEN --snapshot-location spipetools/toolstest > output.csv

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 15th line, since it contains snapshot upload link, which will be different in each run
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".csv" "15d" output.csv
  else
    run sed -i "15d" output.csv
  fi
  cat output.csv

  assert_equal "$(cat output.csv)" "$(cat $TEST_DATA_DIR/expected_static_query_csv_snapshot_mode.csv)"
}

@test "snapshot mode - query output json" {
  cd $FILE_PATH/test_data/functionality_test_mod

  steampipe query query.static_query_2 --snapshot --output json --cloud-token $SPIPETOOLS_TOKEN --snapshot-location spipetools/toolstest > output.json

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 64th line, since it contains snapshot upload link, which will be different in each run
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".csv" "64d" output.json
  else
    run sed -i "64d" output.json
  fi
  cat output.json

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_static_query_json_snapshot_mode.json)"
}

@test "snapshot mode - query output table" {
  cd $FILE_PATH/test_data/functionality_test_mod

  steampipe query query.static_query_2 --snapshot --output table --cloud-token $SPIPETOOLS_TOKEN --snapshot-location spipetools/toolstest > output.txt

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 18th line, since it contains snapshot upload link, which will be different in each run
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".csv" "18d" output.txt
  else
    run sed -i "18d" output.txt
  fi
  cat output.txt

  assert_equal "$(cat output.txt)" "$(cat $TEST_DATA_DIR/expected_static_query_table_snapshot_mode.txt)"
}

function teardown() {
  rm -f output.* 
  cd -
}
