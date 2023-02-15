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

@test "query snapshot mode" {
  cd $FILE_PATH/test_data/functionality_test_mod

  steampipe query query.static_query_2 --snapshot --output csv --cloud-token $SPIPETOOLS_TOKEN --snapshot-location spipetools/toolstest > output.csv
  cat output.csv

  assert_equal "1" "0"
}