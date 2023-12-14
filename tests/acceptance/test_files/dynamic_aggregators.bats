load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# all tests are skipped - https://github.com/turbot/steampipe/issues/3742

# Aggregating two connections with same table and same columns defined.
@test "dynamic aggregator - same table and columns" {
  skip "currently does not pass due to bug - https://github.com/turbot/steampipe/issues/3743"
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_same_table_cols.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_same_table_cols.spc

  steampipe query "select c1,c2 from dyn_agg.t1 order by c1" --output json > output.json
  run jd "$TEST_DATA_DIR/dynamic_aggregators_same_tables_cols_result.json" output.json
  echo $output
  assert_success
  rm -f output.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_same_table_cols.spc
}

# Aggregating two connections with different tables defined.
# Connection `con1` defines a table `t1` whereas connection `con2` defines table `t2`.
@test "dynamic aggregator - table mismatch" {
  skip "currently does not pass due to bug - https://github.com/turbot/steampipe/issues/3743"
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_table_mismatch.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_table_mismatch.spc

  steampipe query "select c1,c2 from dyn_agg.t1 order by c1" --output json > output.json
  run jd "$TEST_DATA_DIR/dynamic_aggregators_table_mismatch_t1.json" output.json
  echo $output
  assert_success
  rm -f output.json

  steampipe query "select c1,c2 from dyn_agg.t2 order by c1" --output json > output.json
  run jd "$TEST_DATA_DIR/dynamic_aggregators_table_mismatch_t2.json" output.json
  echo $output
  assert_success
  rm -f output.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_table_mismatch.spc
}

# Aggregating two connections with same tables defined, but mismatching columns.
# Connection `con1` defines a table `t1` which has columns `c1` and `c2`, whereas connection `con2` also has a table `t1`
# but has columns `c1` and `c3`.
@test "dynamic aggregator - column mismatch" {
  skip "currently does not pass due to bug - https://github.com/turbot/steampipe/issues/3743"
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_mismatch.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_mismatch.spc

  steampipe query "select c1,c2,c3 from dyn_agg.t1 order by c1,c2,c3" --output json > output.json
  run jd "$TEST_DATA_DIR/dynamic_aggregators_col_mismatch.json" output.json
  echo $output
  assert_success
  rm -f output.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_mismatch.spc
}

# Aggregating two connections with same tables defined, but mismatching type of columns.
# Connection `con1` defines a table `t1` which has columns `c1(string)` and `c2(string)`, whereas connection `con2` also has a table `t1`
# but has columns `c1(string)` and `c2(int)`.
@test "dynamic aggregator - column type mismatch(string and int)" {
  skip "currently does not pass due to bug - https://github.com/turbot/steampipe/issues/3743"
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch.spc

  steampipe query "select c1, c2 from dyn_agg.t1 order by c2" --output json > output.json
  run jd "$TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch.json" output.json
  echo $output
  assert_success
  rm -f output.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch.spc
}

# Aggregating two connections with same tables defined, but mismatching type of columns.
# Connection `con1` defines a table `t1` which has columns `c1(string)` and `c2(string)`, whereas connection `con2` also has a table `t1`
# but has columns `c1(string)` and `c2(double)`.
@test "dynamic aggregator - column type mismatch(string and double)" {
  skip "currently does not pass due to bug - https://github.com/turbot/steampipe/issues/3743"
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch_2.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_2.spc

  steampipe query "select c1, c2 from dyn_agg.t1 order by c2" --output json > output.json
  run jd "$TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch_2.json" output.json
  echo $output
  assert_success
  rm -f output.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_2.spc
}

# Aggregating two connections with same tables defined, but mismatching type of columns.
# Connection `con1` defines a table `t1` which has columns `c1(string)` and `c2(string)`, whereas connection `con2` also has a table `t1`
# but has columns `c1(string)` and `c2(bool)`.
@test "dynamic aggregator - column type mismatch(string and bool)" {
  skip "currently does not pass due to bug - https://github.com/turbot/steampipe/issues/3743"
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch_3.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_3.spc

  steampipe query "select c1, c2 from dyn_agg.t1 order by c2" --output json > output.json
  run jd "$TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch_3.json" output.json
  echo $output
  assert_success
  rm -f output.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_3.spc
}

# Aggregating two connections with same tables defined, but mismatching type of columns.
# Connection `con1` defines a table `t1` which has columns `c1(string)` and `c2(string)`, whereas connection `con2` also has a table `t1`
# but has columns `c1(string)` and `c2(ipaddr)`.
@test "dynamic aggregator - column type mismatch(string and ipaddr)" {
  skip "currently does not pass due to bug - https://github.com/turbot/steampipe/issues/3743"
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch_4.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_4.spc

  run steampipe query "select c1, c2 from dyn_agg.t1 order by c1,c2" --output json
  run jd "$TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch_4.json" output.json
  echo $output
  assert_success
  rm -f output.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_4.spc
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
