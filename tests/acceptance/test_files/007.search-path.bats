load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "add connection, check search path updated" {
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_1.txt)"
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_2.txt)"
}

@test "delete connection, check search path updated" {
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_2.txt)"
  cp $SRC_DATA_DIR/single_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_1.txt)"
}

@test "add connection, query with prefix" {
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_1.txt)"
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_3.txt)"
}

@test "delete connection, query with prefix" {
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_2.txt)"
  cp $SRC_DATA_DIR/single_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_4.txt)"
}

@test "query with prefix, add connection, query with prefix" {
  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_5.txt)"
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path" --search-path-prefix foo2
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_6.txt)"
}

@test "query with prefix, delete connection, query with prefix" {
  run steampipe query "show search_path" --search-path-prefix foo2
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_6.txt)"
  cp $SRC_DATA_DIR/single_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_5.txt)"
}
