load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'" {
  run steampipe query "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_table_header.txt)"
}

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0' --header=false" {
  run steampipe query "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'" --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_table_no_header.txt)"
}

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'" {
  run steampipe query --output csv "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_csv_header.csv)"
}

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0' --header=false" {
  run steampipe query --output csv "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'" --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_csv_no_header.csv)"
}

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'" {
  run steampipe query --output csv --separator "|" "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_csv_separator_header.csv)"
}

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0' --header=false" {
  run steampipe query --output csv --separator "|" "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'" --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_csv_separator_no_header.csv)"
}

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'" {
  run steampipe query --output json "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_json.json)"
}

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'" {
  run steampipe query --output line "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_line.txt)"
}

@test "select query install directory" {
  run steampipe query "select 1 as val, 2 as col" --install-dir '~/.steampipe_test'
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_install_directory.txt)"
}

@test "named query current folder" {
  run steampipe query named_query_1.sql
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_named_query_current_folder.txt)"
}

#@test "named query workspace folder" {
#  run steampipe query named_query_4.sql --workspace "/workspace_folder/"
#  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_install_directory.txt)"
#}

@test "sql file" {
  run steampipe query workspace_folder/query_folder/named_query_7.sql
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_sql_file.txt)"
}

@test "sql glob" {
  run steampipe query *.sql
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_sql_glob.txt)"
}

@test "sql glob csv no header" {
  run steampipe query *.sql --header=false --output csv
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_sql_glob_csv_no_header.txt)"
}

