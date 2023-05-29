load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

function setup() {
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  steampipe service "select 1"
}

@test "blank aggregator connection should throw a warning but not fail to run steampipe" {
  cp $SRC_DATA_DIR/blank_aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "select * from all_chaos.chaos_all_numeric_column"
  assert_output --partial "aggregator 'all_chaos' with pattern '*' matches no connections"
}

@test "blank aggregator connection should return empty results and not error" {
  cp $SRC_DATA_DIR/blank_aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "select * from all_chaos.chaos_all_numeric_column"
  assert_equal "$output" "null"
}

