load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"
load "$LIB/connection_map_utils.bash"

@test "add connection, check search path updated" {
  cp $SRC_DATA_DIR/single_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_1.txt)"
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_2.txt)"
}

@test "delete connection, check search path updated" {

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_2.txt)"
  cp $SRC_DATA_DIR/single_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_1.txt)"
}

@test "add connection, query with prefix" {

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_1.txt)"
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_3.txt)"
}

@test "delete connection, query with prefix" {

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_2.txt)"
  cp $SRC_DATA_DIR/single_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_4.txt)"
}

@test "query with prefix, add connection, query with prefix" {

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_5.txt)"
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path" --search-path-prefix foo2
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_6.txt)"
}

@test "query with prefix, delete connection, query with prefix" {

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path" --search-path-prefix foo2
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_6.txt)"
  cp $SRC_DATA_DIR/single_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_5.txt)"
}

@test "verify that 'internal' schema is added" {

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path" --search-path foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_internal_schema_once_1.txt)"
}

@test "verify that 'internal' schema is always suffixed if passed in as custom" {

   # Wait for all connection states to be 'ready'
  run wait_connection_map_stable
  [ "$status" -eq 0 ]

  run steampipe query "show search_path" --search-path foo1,steampipe_internal,foo2
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_internal_schema_once_2.txt)"
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
