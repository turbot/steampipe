load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# TODO rename tests properly

@test "Parsing case 1 - top level query providers do not require query/sql blocks (PASS)" {
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  export STEAMPIPE_LOG=info
  cd $FILE_PATH/test_data/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_top_level --output snapshot
  echo $output>&3
  assert_success
}

@test "Parsing case 2 - top level query providers do not require query/sql blocks except control/query (PASS)" {
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  export STEAMPIPE_LOG=info
  cd $FILE_PATH/test_data/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_top_level_require_sql --output snapshot
  echo $output>&3
  assert_success
}

@test "Parsing case 3 - top level control/query always require query/sql block (FAIL)" {
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  export STEAMPIPE_LOG=info
  cd $FILE_PATH/test_data/dashboard_parsing_top_level_query_providers_fail

  run steampipe dashboard dashboard.top_level_control_query_require_sql --output snapshot
  echo $output>&3
  assert_output --partial 'does not define a query or SQL'
}

@test "Parsing case 4 - nested query providers do require query/sql blocks (PASS)" {
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  export STEAMPIPE_LOG=info
  cd $FILE_PATH/test_data/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_nested --output snapshot
  echo $output>&3
  assert_success
}

@test "Parsing case 5 - nested query providers do require query/sql blocks (FAIL)" {
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  export STEAMPIPE_LOG=info
  cd $FILE_PATH/test_data/dashboard_parsing_nested_query_providers_fail

  run steampipe dashboard dashboard.query_providers_nested --output snapshot
  echo $output>&3
  assert_output --partial 'does not define a query or SQL'
}

@test "Parsing case 6 - nested query providers do not require require query/sql blocks except images/cards (PASS)" {
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  export STEAMPIPE_LOG=info
  cd $FILE_PATH/test_data/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_nested_dont_require_sql --output snapshot
  echo $output>&3
  assert_success
}

@test "Parsing case 7 - top level node and edge providers do not require a query/sql block or a node/edge block (PASS)" {
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  export STEAMPIPE_LOG=info
  cd $FILE_PATH/test_data/dashboard_parsing_validation

  run steampipe dashboard dashboard.node_edge_providers_top_level --output snapshot
  echo $output>&3
  assert_success
}

@test "Parsing case 8 - nested node and edge providers always require a query/sql block or a node/edge block (PASS)" {
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  export STEAMPIPE_LOG=info
  cd $FILE_PATH/test_data/dashboard_parsing_validation

  run steampipe dashboard dashboard.node_edge_providers_nested --output snapshot
  echo $output>&3
  assert_success
}

@test "Parsing case 9 - nested node and edge providers do require a query/sql block or a node/edge block (FAIL)" {
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  export STEAMPIPE_LOG=info
  cd $FILE_PATH/test_data/dashboard_parsing_nested_node_edge_providers_fail

  run steampipe dashboard dashboard.node_edge_providers_nested --output snapshot
  echo $output>&3
  assert_output --partial 'does not define a query or SQL, and has no edges/nodes'
}

# run teardown with 30s sleep after each test since it takes some time to kill all plugins in pluginMultiConnectionMap
function teardown() {
  echo "starting teardown">&3
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  sleep 1
  # list running processes
  psx=$(ps aux | grep steampipe)
  echo $psx>&3

  # check if any processes are running
  num=$($psx | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
  echo "end teardown">&3
  date +"%Y-%m-%dT%H:%M:%S%z">&3
  echo "">&3
}

# disable parallelisation only within the containing file.
function setup_file() {
  export BATS_NO_PARALLELIZE_WITHIN_FILE=true
}
