load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# These set of tests are skipped locally
# To run these tests locally set the SPIPETOOLS_TOKEN env var.
# These tests will be skipped locally unless the below env var is set.

function setup() {
  if [[ -z "${SPIPETOOLS_TOKEN}" ]]; then
    skip
  fi
}

# These set of tests check the different types of output in query snapshot mode and not snapshot creation/upload
# Related to https://github.com/turbot/steampipe/issues/3112

@test "snapshot mode - query output csv" {

  steampipe query $FILE_PATH/test_data/mods/functionality_test_mod/query/static_query_2.sql --snapshot --output csv --pipes-token $SPIPETOOLS_TOKEN --snapshot-location turbot-ops/clitesting > output.csv

  # extract the snapshot url from the output
  url=$(grep -o 'http[^"]*' output.csv)
  echo $url

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 15th line, since it contains snapshot upload link, which will be different in each run
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".csv" "2d" output.csv
  else
    run sed -i "2d" output.csv
  fi
  cat output.csv

  # create the snapshot DELETE Request URL
  req_url=$($FILE_PATH/url_parse.sh $url)
  echo $req_url

  assert_equal "$(cat output.csv)" "$(cat $TEST_DATA_DIR/expected_static_query_csv_snapshot_mode.csv)"
  rm -f output.*

  # delete the snapshot from cloud workspace to avoid exceeding quota
  curl -X DELETE "$req_url" -H "Authorization: Bearer $SPIPETOOLS_TOKEN"
}

@test "snapshot mode - query output json" {
  skip
  steampipe query $FILE_PATH/test_data/mods/functionality_test_mod/query/static_query_2.sql --snapshot --output json --pipes-token $SPIPETOOLS_TOKEN --snapshot-location turbot-ops/clitesting > output.json

  # extract the snapshot url from the output
  url=$(grep -o 'http[^"]*' output.json)
  echo $url

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 64th line, since it contains snapshot upload link, which will be different in each run
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".csv" "2d" output.json
  else
    run sed -i "2d" output.json
  fi
  cat output.json

  # create the snapshot DELETE Request URL
  req_url=$($FILE_PATH/url_parse.sh $url)
  echo $req_url

  assert_equal "$(cat output.json)" "$(cat $TEST_DATA_DIR/expected_static_query_json_snapshot_mode.json)"
  rm -f output.*

  # delete the snapshot from cloud workspace to avoid exceeding quota
  curl -X DELETE "$req_url" -H "Authorization: Bearer $SPIPETOOLS_TOKEN"
}

@test "snapshot mode - query output table" {

  steampipe query $FILE_PATH/test_data/mods/functionality_test_mod/query/static_query_2.sql --snapshot --output table --pipes-token $SPIPETOOLS_TOKEN --snapshot-location turbot-ops/clitesting > output.txt

  # extract the snapshot url from the output
  url=$(grep -o 'http[^"]*' output.txt)
  echo $url

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 18th line, since it contains snapshot upload link, which will be different in each run
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".csv" "2d" output.txt
  else
    run sed -i "2d" output.txt
  fi
  cat output.txt

  # create the snapshot DELETE Request URL
  req_url=$($FILE_PATH/url_parse.sh $url)
  echo $req_url

  assert_equal "$(cat output.txt)" "$(cat $TEST_DATA_DIR/expected_static_query_table_snapshot_mode.txt)"
  rm -f output.*

  # delete the snapshot from cloud workspace to avoid exceeding quota
  curl -X DELETE "$req_url" -H "Authorization: Bearer $SPIPETOOLS_TOKEN"
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
