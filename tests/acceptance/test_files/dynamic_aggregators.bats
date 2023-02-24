load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "dynamic aggregator - same table and columns" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_same_table_cols.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_same_table_cols.spc

  run steampipe query "select c1,c2 from dyn_agg.t1 order by c1" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_same_table_cols.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_same_tables_cols_result.json)"
}

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

@test "dynamic aggregator - column mismatch" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_mismatch.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_mismatch.spc

  run steampipe query "select c1,c2,c3 from dyn_agg.t1 order by c1,c2,c3" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_mismatch.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_col_mismatch.json)"
}

@test "dynamic aggregator - column type mismatch(string and int)" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch.spc

  run steampipe query "select c1, c2 from dyn_agg.t1 order by c2" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch.json)"
}

@test "dynamic aggregator - column type mismatch(string and double)" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch_2.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_2.spc

  run steampipe query "select c1, c2 from dyn_agg.t1 order by c2" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_2.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch_2.json)"
}

@test "dynamic aggregator - column type mismatch(string and bool)" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch_3.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_3.spc

  run steampipe query "select c1, c2 from dyn_agg.t1 order by c2" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_3.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch_3.json)"
}

@test "dynamic aggregator - column type mismatch(string and ipaddr)" {
  cp $SRC_DATA_DIR/dynamic_aggregator_tests/dynamic_aggregator_col_type_mismatch_4.spc $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_4.spc

  run steampipe query "select c1, c2 from dyn_agg.t1 order by c1,c2" --output json
  echo $output

  rm -f $STEAMPIPE_INSTALL_DIR/config/dynamic_aggregator_col_type_mismatch_4.spc
  assert_equal "$output" "$(cat $TEST_DATA_DIR/dynamic_aggregators_col_type_mismatch_4.json)"
}
