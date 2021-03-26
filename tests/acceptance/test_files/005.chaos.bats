load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "select * from chaos.chaos_high_row_count order by column_0" {
  run steampipe query --output json  "select * from chaos.chaos_high_row_count order by column_0"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_1.json)"
}

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'" {
  run steampipe query --output json  "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_2.json)"
}

@test "select * from chaos.chaos_high_column_count order by column_0" {
  run steampipe query --output json  "select * from chaos.chaos_high_column_count order by column_0"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_3.json)"
}

@test "select * from chaos.chaos_parent_child_dependency order by column_1 limit 2" {
  run steampipe query --output json "select * from chaos.chaos_parent_child_dependency order by column_1 limit 2"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_4.json)"
}

@test "select * from chaos.chaos_hydrate_columns_dependency where id='0'" {
  run steampipe query --output json "select * from chaos.chaos_hydrate_columns_dependency where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_5.json)"
}

@test "select * from chaos.chaos_list_error" {
  run steampipe query "select * from chaos.chaos_list_error"
  assert_output --partial 'LIST ERROR'
}

@test "select * from chaos.chaos_get_panic where column_0='column_0-3'" {
  run steampipe query "select * from chaos.chaos_get_panic where column_0='column_0-3'"
  assert_output --partial 'GET PANIC'
}

@test "select * from chaos.chaos_transform_error" {
  run steampipe query "select * from chaos.chaos_transform_error"
  assert_output --partial 'TRANSFORM ERROR'
}

@test "select * from chaos.chaos_hydrate_delay" {
  run steampipe query "select * from chaos.chaos_hydrate_delay"
  assert_success
}

@test "elect * from chaos.chaos_parallel_hydrate_columns  where id='0'" {
  run steampipe query --output json "select * from chaos.chaos_parallel_hydrate_columns  where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_11.json)"
}

@test "select float32_data, id, int64_data, uint16_data from chaos.chaos_all_numeric_column  where id='31'" {
  run steampipe query --output json "select float32_data, id, int64_data, uint16_data from chaos.chaos_all_numeric_column  where id='31'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_12.json)"
}

@test "select * from chaos.chaos_default_transform  where id='0'" {
  run steampipe query --output json "select * from chaos.chaos_default_transform  where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_13.json)"
}

@test "select * from chaos.chaos_transform_method_test  where id='0'" {
  run steampipe query --output json "select * from chaos.chaos_transform_method_test  where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_14.json)"
}
