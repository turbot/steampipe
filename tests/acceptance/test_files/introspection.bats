load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe_introspection=none" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=none
  run steampipe query "select * from steampipe_query" --output json

  assert_output --partial 'relation "steampipe_query" does not exist'
}

@test "resource=query | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_query" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_query.json)"
}

@test "resource=control | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_control" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_control.json)"
}

@test "resource=variable | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_variable" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_variable.json)"
}

@test "resource=benchmark | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_benchmark" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_benchmark.json)"
}

@test "resource=dashboard | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_dashboard" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard.json)"
}

@test "resource=dashboard_card | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_dashboard_card" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_card.json)"
}

@test "resource=dashboard_image | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_dashboard_image" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_image.json)"
}

@test "resource=dashboard_text | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_dashboard_text" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_text.json)"
}

@test "resource=dashboard_chart | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_dashboard_chart" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_chart.json)"
}

@test "resource=dashboard_flow | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_dashboard_flow" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_flow.json)"
}

@test "resource=dashboard_graph | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_dashboard_graph" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_graph.json)"
}

@test "resource=dashboard_hierarchy | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_dashboard_hierarchy" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_hierarchy.json)"
}

@test "resource=dashboard_input | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_dashboard_input" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_input.json)"
}

@test "resource=dashboard_table | steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_dashboard_table" --output json

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_introspection_info_dashboard_table.json)"
}