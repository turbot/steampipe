load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# testing the check summary output feature in steampipe
@test "check summary output" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check benchmark.control_summary_benchmark --theme plain

  echo $output

  # TODO: Find a way to store the output in a file and match it with the 
  # expected file. For now the work-around is to check whether the output
  # contains `summary`
  assert_output --partial 'Summary'
}
