load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "prepare dynamic aggregator tests" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_same_table_cols.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_same_table_cols.spc
  steampipe query "select 1"
}

# Aggregating two connections with same table and same columns defined.
@test "dynamic aggregator - same table and columns" {
  # cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_same_table_cols.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_same_table_cols.spc

  run steampipe query "select c1,c2 from dyn_agg.t1 order by c1" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_same_table_cols.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_same_tables_cols_result.json)"
}

# Aggregating two connections with different tables defined.
# Connection `con1` defines a table `t1` whereas connection `con2` defines table `t2`.
@test "dynamic aggregator - table mismatch" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_table_mismatch.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_table_mismatch.spc

  run steampipe query "select c1,c2 from dyn_agg.t1 order by c1" --output json
  echo $output
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_table_mismatch_t1.json)"

  run steampipe query "select c1,c2 from dyn_agg.t2 order by c1" --output json
  echo $output
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_table_mismatch_t2.json)"

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_table_mismatch.spc
}

# Aggregating two connections with same tables defined, but mismatching columns.
# Connection `con1` defines a table `t1` which has columns `c1` and `c2`, whereas connection `con2` also has a table `t1`
# but has columns `c1` and `c3`.
@test "dynamic aggregator - column mismatch" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_mismatch.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_mismatch.spc

  run steampipe query "select c1,c2,c3 from dyn_agg.t1 order by c1,c2,c3" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_mismatch.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_col_mismatch.json)"
}

# Aggregating two connections with same tables defined, but mismatching type of columns.
# Connection `con1` defines a table `t1` which has columns `c1(string)` and `c2(string)`, whereas connection `con2` also has a table `t1`
# but has columns `c1(string)` and `c2(int)`.
@test "dynamic aggregator - column type mismatch(string and int)" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch.spc

  run steampipe query "select c1, c2 from dyn_agg.t1 order by c2" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch.json)"
}

# Aggregating two connections with same tables defined, but mismatching type of columns.
# Connection `con1` defines a table `t1` which has columns `c1(string)` and `c2(string)`, whereas connection `con2` also has a table `t1`
# but has columns `c1(string)` and `c2(double)`.
@test "dynamic aggregator - column type mismatch(string and double)" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch_2.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_2.spc

  run steampipe query "select c1, c2 from dyn_agg.t1 order by c2" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_2.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch_2.json)"
}

# Aggregating two connections with same tables defined, but mismatching type of columns.
# Connection `con1` defines a table `t1` which has columns `c1(string)` and `c2(string)`, whereas connection `con2` also has a table `t1`
# but has columns `c1(string)` and `c2(bool)`.
@test "dynamic aggregator - column type mismatch(string and bool)" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch_3.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_3.spc

  run steampipe query "select c1, c2 from dyn_agg.t1 order by c2" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_3.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch_3.json)"
}

# Aggregating two connections with same tables defined, but mismatching type of columns.
# Connection `con1` defines a table `t1` which has columns `c1(string)` and `c2(string)`, whereas connection `con2` also has a table `t1`
# but has columns `c1(string)` and `c2(ipaddr)`.
@test "dynamic aggregator - column type mismatch(string and ipaddr)" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch_4.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_4.spc

  run steampipe query "select c1, c2 from dyn_agg.t1 order by c1,c2" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_4.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch_4.json)"
}

function setup_file() {
  export STEAMPIPE_SYNC_REFRESH=true
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
