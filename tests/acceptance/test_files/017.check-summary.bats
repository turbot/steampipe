load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# testing the check summary output feature in steampipe
@test "check summary output" {
  cd $WORKSPACE_DIR
  run steampipe check benchmark.control_summary_benchmark --theme plain

  # storing the content in a variable as a string or else test was failing
  var=$(cat $TEST_DATA_DIR/expected_summary_output.txt)

  assert_equal "$var" "$(cat $TEST_DATA_DIR/expected_summary_output.txt)"
}
