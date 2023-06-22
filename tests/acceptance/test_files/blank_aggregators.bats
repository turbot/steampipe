load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# function setup() {
#   rm -f $STEAMPIPE_INSTALL_DIR/config/chaos.spc
#   steampipe service "select 1"
# }

@test "blank aggregator connection should throw a warning but not fail to run steampipe" {
  skip
  cp $SRC_DATA_DIR/blank_aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "select * from all_chaos.chaos_all_numeric_column"
  echo $output
  assert_output --partial "aggregator 'all_chaos' with pattern '*' matches no connections"
}

@test "blank aggregator connection should return empty results and not error" {
  skip
  cp $SRC_DATA_DIR/blank_aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "select * from all_chaos.chaos_all_numeric_column"
  echo $output
  assert_equal "$output" "null"
}

@test "blank aggregator connection schema not created issue" {
  skip
  # for blank aggregator connections, schema was not getting created while service was running
  # https://github.com/turbot/steampipe/issues/3488
  run steampipe service start
  cp $SRC_DATA_DIR/blank_aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "select * from all_chaos.chaos_all_numeric_column"
  echo $output
  steampipe service stop
  assert_equal "$output" "null"
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
