load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "service start, no config, add connection, query" {
  steampipe service start
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_1.txt)"
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  steampipe service restart
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_2.txt)"
}

@test "service start, no config, delete connection, query with no restart" {
  steampipe service start
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_2.txt)"
  cp $SRC_DATA_DIR/single_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_1.txt)"
}

@test "no service start, no config, add connection, query with prefix" {
  steampipe service start
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_1.txt)"
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path" --search-path-prefix foo
  # NOTE had to add blank line to expected output for some reason
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_3.txt)"
}

@test "service start, no config, delete connection, query with prefix" {
  steampipe service start
  run steampipe query "show search_path"
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_2.txt)"
  cp $SRC_DATA_DIR/single_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_4.txt)"
}

@test "service start, no config, query with prefix, add connection, query with prefix" {
  steampipe service start
  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_5.txt)"
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path" --search-path-prefix foo2
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_6.txt)"
}

@test "service start, no config, query with prefix, delete connection, query with prefix" {
  steampipe service start
  run steampipe query "show search_path" --search-path-prefix foo2
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_6.txt)"
  cp $SRC_DATA_DIR/single_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
  run steampipe query "show search_path" --search-path-prefix foo
  assert_output "$(cat $TEST_DATA_DIR/expected_search_path_5.txt)"
}

# function setup() {
#   STEAMPIPE_PASSWD=$(cat $STEAMPIPE_INSTALL_DIR/db/12.1.0/postgres/.passwd | jq ".Steampipe")
#   STEAMPIPE_PASSWD="${STEAMPIPE_PASSWD%\"}"
#   STEAMPIPE_PASSWD="${STEAMPIPE_PASSWD#\"}"
#   steampipe service start
#   steampipe plugin install chaos
#   steampipe service stop --force
# }

function teardown() {
    steampipe service stop --force    
}
