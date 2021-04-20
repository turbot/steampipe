load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "table with header" {
  run steampipe query "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_table_header.txt)"
}

@test "table no header" {
  run steampipe query "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'" --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_table_no_header.txt)"
}

@test "csv header" {
  run steampipe query --output csv "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_csv_header.csv)"
}

@test "csv no header" {
  run steampipe query --output csv "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'" --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_csv_no_header.csv)"
}

@test "csv | separator" {
  run steampipe query --output csv --separator "|" "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_csv_separator_header.csv)"
}

@test "csv | separator no header" {
  run steampipe query --output csv --separator "|" "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'" --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_csv_separator_no_header.csv)"
}

@test "json" {
  run steampipe query --output json "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_json.json)"
}

@test "line" {
  run steampipe query --output line "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_line.txt)"
}

@test "timer on" {
  run steampipe query "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'" --timing
  assert_output --partial 'Time:'
}

@test "select query install directory" {
  run steampipe query --output csv "select 1" --install-dir '~/.steampipe_test'
  assert_success
}

@test "named query current folder" {
  cd tests/acceptance/test_files
  run steampipe query query.named_query_1
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_named_query_current_folder.txt)"
}

@test "named query workspace folder" {
  run steampipe query query.named_query_4 --workspace "tests/acceptance/test_files/workspace_folder/"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_workspace_folder.txt)"
}

@test "sql file" {
  run steampipe query tests/acceptance/test_files/workspace_folder/query_folder/named_query_7.sql
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_sql_file.txt)"
}

@test "sql glob" {
  cd tests/acceptance/test_files
  run steampipe query *.sql
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_sql_glob.txt)"
}

@test "sql glob csv no header" {
  cd tests/acceptance/test_files
  run steampipe query *.sql --header=false --output csv
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_sql_glob_csv_no_header.txt)"
}

