load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "select * from chaos.chaos_high_row_count order by column_0" {
  run steampipe query --output json  "select * from chaos.chaos_high_row_count order by column_0 limit 10"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_1.json)"
}

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'" {
  run steampipe query --output json  "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_2.json)"
}

@test "select * from chaos.chaos_high_column_count order by column_0" {
  run steampipe query --output json  "select * from chaos.chaos_high_column_count order by column_0 limit 10"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_3.json)"
}

@test "select * from chaos.chaos_hydrate_columns_dependency where id='0'" {
  run steampipe query --output json "select * from chaos.chaos_hydrate_columns_dependency where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_5.json)"
}

@test "select * from chaos.chaos_list_error" {
  run steampipe query "select fatal_error from chaos.chaos_list_errors"
  assert_output --partial 'fatalError'
}

@test "select panic from chaos.chaos_get_errors where id=0" {
  run steampipe query --output json "select panic from chaos.chaos_get_errors where id=0"
   assert_output --partial 'Panic'
}

@test "select error from chaos_transform_errors" {
  run steampipe query "select error from chaos_transform_errors"
  assert_output --partial 'TRANSFORM ERROR'
}

@test "select * from chaos.chaos_hydrate_delay" {
  run steampipe query --output json "select delay from chaos.chaos_hydrate_errors order by id"
  assert_success
}

@test "select * from chaos.chaos_parallel_hydrate_columns  where id='0'" {
  run steampipe query --output json "select * from chaos.chaos_parallel_hydrate_columns  where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_11.json)"
}

@test "select float32_data, id, int64_data, uint16_data from chaos.chaos_all_numeric_column  where id='31'" {
  run steampipe query --output json "select float32_data, id, int64_data, uint16_data from chaos.chaos_all_numeric_column  where id='31'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_12.json)"
}

@test "select transform_method_column from chaos_transforms order by id" {
  run steampipe query --output json "select transform_method_column from chaos_transforms order by id"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_14.json)"
}

@test "select parent_should_ignore_error from chaos.chaos_list_parent_child" {
  run steampipe query "select parent_should_ignore_error from chaos.chaos_list_parent_child"
  assert_success
}

@test "select from_qual_column from chaos_transforms where id=2" {
  run steampipe query --output json "select from_qual_column from chaos_transforms where id=2"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_13.json)"
}
