load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "public schema insert select all types" {
  steampipe query "drop table if exists all_columns"
  run steampipe query "create table all_columns (nullcolumn CHAR(2), booleancolumn boolean, textcolumn1 CHAR(20), textcolumn2 VARCHAR(20),  textcolumn3 text, integercolumn1 smallint, integercolumn2 int, integercolumn3 SERIAL, integercolumn4 bigint,  integercolumn5 bigserial, numericColumn numeric(6,4), realColumn real, floatcolumn float,  date1 DATE,  time1 TIME,  timestamp1 TIMESTAMP, interval1 TIMESTAMPTZ, timestamp2 INTERVAL, array1 text[], jsondata jsonb, jsondata2 json, uuidcolumn UUID, ipAddress inet, macAddress macaddr, cidrRange cidr, xmlData xml, currency money)"
  run steampipe query "INSERT INTO all_columns (nullcolumn, booleancolumn, textcolumn1, textcolumn2, textcolumn3, integercolumn1, integercolumn2, integercolumn3, integercolumn4, integercolumn5, numericColumn, realColumn, floatcolumn, date1, time1, timestamp1, interval1, timestamp2, array1, jsondata, jsondata2, uuidcolumn, ipAddress, macAddress, cidrRange, xmlData, currency) VALUES (NULL, TRUE, 'Yes', 'test for varchar', 'This is a very long text for the PostgreSQL text column', 3278, 21445454, 2147483645, 92233720368547758, 922337203685477580, 23.5141543, 4660.33777, 4.6816421254887534, '1978-02-05', '08:00:00', '2016-06-22 19:10:25-07', '2016-06-22 19:10:25-07', '1 year 2 months 3 days', '{\"(408)-589-5841\"}','{ \"customer\": \"John Doe\", \"items\": {\"product\": \"Beer\",\"qty\": 6}}', '{ \"customer\": \"John Doe\", \"items\": {\"product\": \"Beer\",\"qty\": 6}}', '6948DF80-14BD-4E04-8842-7668D9C001F5', '192.168.0.0', '08:00:2b:01:02:03', '10.1.2.3/32', '<book><title>Manual</title><chapter>...</chapter></book>', 922337203685477.57)"
  run steampipe query "select * from all_columns" --output json
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_6.json)"
  run steampipe query "drop table all_columns"
}

@test "query json" {
  run steampipe query "select 1 as val, 2 as col" --output json
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_query_json.json)"
}

@test "query csv" {
  run steampipe query "select 1 as val, 2 as col" --output csv
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_query_csv.csv)"
}

@test "query line" {
  run steampipe query "select 1 as val, 2 as col" --output line
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_query_line.txt)"
}

@test "query line long" {
  run steampipe query "drop table if exists long_columns"
  run steampipe query "create table long_columns (shortstring char(20), longstring char(3900))"
  run steampipe query "INSERT INTO long_columns (shortstring,longstring) VALUES ('a short text','tincidunt dui ut ornare lectus sit amet est placerat in egestas erat imperdiet sed euismod nisi porta lorem mollis aliquam ut porttitor leo a diam sollicitudin tempor id eu nisl nunc mi ipsum faucibus vitae aliquet nec ullamcorper sit amet risus nullam eget felis eget nunc lobortis mattis aliquam faucibus purus in massa tempor nec feugiat nisl pretium fusce id velit ut tortor pretium viverra suspendisse potenti nullam ac tortor vitae purus faucibus ornare suspendisse sed nisi lacus sed viverra tellus in hac habitasse platea dictumst vestibulum rhoncus est pellentesque elit ullamcorper dignissim cras tincidunt lobortis feugiat vivamus at augue eget arcu dictum varius duis at consectetur lorem donec massa sapien faucibus et molestie ac feugiat sed lectus vestibulum mattis ullamcorper velit sed ullamcorper morbi tincidunt ornare massa eget egestas purus viverra accumsan in nisl nisi scelerisque eu ultrices vitae auctor eu augue ut lectus arcu bibendum at varius vel pharetra vel turpis nunc eget lorem dolor sed viverra ipsum nunc aliquet bibendum enim facilisis gravida neque convallis a cras semper auctor neque vitae tempus quam pellentesque nec nam aliquam sem et tortor consequat id porta nibh venenatis cras sed felis eget velit aliquet sagittis id consectetur purus ut faucibus pulvinar elementum integer enim neque volutpat ac tincidunt vitae semper quis lectus nulla at volutpat diam ut venenatis tellus in metus vulputate eu scelerisque felis imperdiet proin fermentum leo vel orci porta non pulvinar neque laoreet suspendisse interdum consectetur libero id faucibus nisl tincidunt eget nullam non nisi est sit amet facilisis magna etiam tempor orci eu lobortis elementum nibh tellus molestie nunc non blandit massa enim nec dui nunc mattis enim ut tellus elementum sagittis vitae et leo duis ut diam quam nulla porttitor massa id neque aliquam vestibulum morbi blandit cursus risus at ultrices mi tempus imperdiet nulla malesuada pellentesque elit eget gravida cum sociis natoque penatibus et magnis dis parturient montes nascetur ridiculus mus mauris vitae ultricies leo integer malesuada nunc vel risus commodo viverra maecenas accumsan lacus vel facilisis volutpat est velit egestas dui id ornare arcu odio ut sem nulla pharetra diam sit amet nisl suscipit adipiscing bibendum est ultricies integer quis auctor elit sed vulputate mi sit amet mauris commodo quis imperdiet massa tincidunt nunc pulvinar sapien et ligula ullamcorper malesuada proin libero nunc consequat interdum varius sit amet mattis vulputate enim nulla aliquet porttitor lacus luctus accumsan tortor posuere ac ut consequat semper viverra nam libero justo laoreet sit amet cursus sit amet dictum sit amet justo donec enim diam vulputate ut pharetra sit amet aliquam id diam maecenas ultricies mi eget mauris pharetra et ultrices neque ornare aenean euismod elementum nisi quis eleifend quam adipiscing vitae proin sagittis nisl rhoncus mattis rhoncus urna neque viverra justo nec ultrices dui sapien eget mi proin sed libero enim sed faucibus turpis in eu mi bibendum neque egestas congue quisque egestas diam in arcu cursus euismod quis viverra nibh cras pulvinar mattis nunc sed blandit libero volutpat sed cras ornare arcu dui vivamus arcu felis bibendum ut tristique et egestas quis ipsum suspendisse ultrices gravida dictum fusce ut placerat orci nulla pellentesque dignissim enim sit amet venenatis urna cursus eget nunc scelerisque viverra mauris in aliquam sem fringilla ut morbi tincidunt augue interdum velit euismod in pellentesque massa placerat duis ultricies lacus sed turpis tincidunt id aliquet risus feugiat in ante metus dictum at tempor commodo ullamcorper a lacus vestibulum sed arcu non odio euismod lacinia at quis risus sed vulputate odio ut enim blandit volutpat maecenas volutpat blandit aliquam etiam erat velit scelerisque in dictum non consectetur a erat nam at lectus urna duis')"
  run steampipe query "select * from long_columns" --output line
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_line_long.txt)"
  run steampipe query "drop table long_columns"
}

@test "query csv header off" {
  run steampipe query "select 1 as val, 2 as col" --output csv --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_query_csv_header_off.csv)"
}

@test "query table header off" {
  run steampipe query "select 1 as val, 2 as col" --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_query_table_header_off.txt)"
}

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
  run steampipe query query.named_query_4 --workspace-chdir "tests/acceptance/test_files/workspace_folder/"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_workspace_folder.txt)"
}

@test "named query workspace query folder(should fail - no mod.sp file)" {
  run steampipe query query.named_query_7 --workspace-chdir "tests/acceptance/test_files/workspace_folder/query_folder"
  
  # This query should fail since the folder does not contain a mod.sp file
  assert_output --partial 'not found in workspace'
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
