load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# TODO rename tests properly

@test "Parsing case 1 - top level query providers do not require query/sql blocks (PASS)" {
  cd $FILE_PATH/test_data/mods/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_top_level --output snapshot
  assert_success
}

@test "Parsing case 2 - top level query providers do not require query/sql blocks except control/query (PASS)" {
  cd $FILE_PATH/test_data/mods/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_top_level_require_sql --output snapshot
  assert_success
}

@test "Parsing case 3 - top level control/query always require query/sql block (FAIL)" {
  cd $FILE_PATH/test_data/mods/dashboard_parsing_top_level_query_providers_fail

  run steampipe dashboard dashboard.top_level_control_query_require_sql --output snapshot
  assert_output --partial 'does not define a query or SQL'
}

@test "Parsing case 4 - nested query providers do require query/sql blocks (PASS)" {
  cd $FILE_PATH/test_data/mods/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_nested --output snapshot
  assert_success
}

@test "Parsing case 5 - nested query providers do require query/sql blocks (FAIL)" {
  cd $FILE_PATH/test_data/mods/dashboard_parsing_nested_query_providers_fail

  run steampipe dashboard dashboard.query_providers_nested --output snapshot
  assert_output --partial 'does not define a query or SQL'
}

@test "Parsing case 6 - nested query providers do not require require query/sql blocks except images/cards (PASS)" {
  cd $FILE_PATH/test_data/mods/dashboard_parsing_validation

  run steampipe dashboard dashboard.query_providers_nested_dont_require_sql --output snapshot
  assert_success
}

@test "Parsing case 7 - top level node and edge providers do not require a query/sql block or a node/edge block (PASS)" {
  cd $FILE_PATH/test_data/mods/dashboard_parsing_validation

  run steampipe dashboard dashboard.node_edge_providers_top_level --output snapshot
  assert_success
}

@test "Parsing case 8 - nested node and edge providers always require a query/sql block or a node/edge block (PASS)" {
  cd $FILE_PATH/test_data/mods/dashboard_parsing_validation

  run steampipe dashboard dashboard.node_edge_providers_nested --output snapshot
  assert_success
}

@test "Parsing case 9 - nested node and edge providers do require a query/sql block or a node/edge block (FAIL)" {
  cd $FILE_PATH/test_data/mods/dashboard_parsing_nested_node_edge_providers_fail

  run steampipe dashboard dashboard.node_edge_providers_nested --output snapshot
  assert_output --partial 'does not define a query or SQL, and has no edges/nodes'
}

@test "Parsing case 10 - nested dashboards (PASS)" {
  cd $FILE_PATH/test_data/mods/dashboard_parsing_validation

  run steampipe dashboard dashboard.nested_dashboards --output snapshot
  assert_success
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
