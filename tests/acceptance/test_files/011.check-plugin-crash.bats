load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "check whether the plugin is crashing or not" {
  skip
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check benchmark.check_plugin_crash_benchmark
  echo $output
  [ $(echo $output | grep "ERROR: context canceled" | wc -l | tr -d ' ') -eq 0 ]
}
