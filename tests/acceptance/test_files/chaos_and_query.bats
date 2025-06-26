load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "select from chaos.chaos_high_row_count order by column_0" {
  run steampipe query --output json  "select column_0,column_1,column_2,column_3,column_4,column_5,column_6,column_7,column_8,column_9,id from chaos.chaos_high_row_count order by column_0 limit 10"
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_1 files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_1.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""

  rm -f $TEST_DATA_DIR/actual_1.json
}

@test "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'" {
  run steampipe query --output json  "select id, string_column, json_column, boolean_column from chaos.chaos_all_column_types where id='0'"
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_2 files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_2.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""

  rm -f $TEST_DATA_DIR/actual_1.json
}

@test "select from chaos.chaos_high_column_count order by column_0" {
  skip
  run steampipe query --output json  "select * from chaos.chaos_high_column_count order by column_0 limit 10"
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_3 files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_3.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""

  rm -f $TEST_DATA_DIR/actual_1.json
}

@test "select from chaos.chaos_hydrate_columns_dependency where id='0'" {  
  run steampipe query --output json "select hydrate_column_1,hydrate_column_2,hydrate_column_3,hydrate_column_4,hydrate_column_5,id from chaos.chaos_hydrate_columns_dependency where id='0'"
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_5 files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_5.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""

  rm -f $TEST_DATA_DIR/actual_1.json
}

@test "select from chaos.chaos_list_error" {
  run steampipe query "select fatal_error from chaos.chaos_list_errors"
  assert_output --partial 'fatalError'
}

@test "select panic from chaos.chaos_get_errors where id=0" {
  run steampipe query --output json "select panic from chaos.chaos_get_errors where id=0"
   assert_output --partial 'Panic'
}

@test "select error from chaos_transform_errors" {
  skip "skipped till chaos_transform_errors table is modified"
  run steampipe query "select error from chaos_transform_errors"
  assert_output --partial 'TRANSFORM ERROR'
}

@test "select from chaos.chaos_hydrate_delay" {
  run steampipe query --output json "select delay from chaos.chaos_hydrate_errors order by id"
  assert_success
}

@test "select from chaos.chaos_parallel_hydrate_columns  where id='0'" {  
  run steampipe query --output json "select column_1,column_10,column_11,column_12,column_13,column_14,column_15,column_16,column_17,column_18,column_19,column_2,column_20,column_3,column_4,column_5,column_6,column_7,column_8,column_9,id from chaos.chaos_parallel_hydrate_columns  where id='0'"
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_11 files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_11.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""
  
  rm -f $TEST_DATA_DIR/actual_1.json
}

@test "select float32_data, id, int64_data, uint16_data from chaos.chaos_all_numeric_column  where id='31'" {
  run steampipe query --output json "select float32_data, id, int64_data, uint16_data from chaos.chaos_all_numeric_column  where id='31'"
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_12 files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_12.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""
  
  rm -f $TEST_DATA_DIR/actual_1.json
}

@test "select transform_method_column from chaos_transforms order by id" {
  run steampipe query --output json "select transform_method_column from chaos_transforms order by id"
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_14 files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_14.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""
  
  rm -f $TEST_DATA_DIR/actual_1.json
}

@test "select parent_should_ignore_error from chaos.chaos_list_parent_child" {
  run steampipe query "select parent_should_ignore_error from chaos.chaos_list_parent_child"
  assert_success
}

@test "select from_qual_column from chaos_transforms where id=2" {
  run steampipe query --output json "select from_qual_column from chaos_transforms where id=2"
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_13 files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_13.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""
  
  rm -f $TEST_DATA_DIR/actual_1.json
}

@test "public schema insert select all types" {
  skip
  steampipe query "drop table if exists all_columns"
  steampipe query "create table all_columns (nullcolumn CHAR(2), booleancolumn boolean, textcolumn1 CHAR(20), textcolumn2 VARCHAR(20),  textcolumn3 text, integercolumn1 smallint, integercolumn2 int, integercolumn3 SERIAL, integercolumn4 bigint,  integercolumn5 bigserial, numericColumn numeric(6,4), realColumn real, floatcolumn float,  date1 DATE,  time1 TIME,  timestamp1 TIMESTAMP, timestamp2 TIMESTAMPTZ, interval1 INTERVAL, array1 text[], jsondata jsonb, jsondata2 json, uuidcolumn UUID, ipAddress inet, macAddress macaddr, cidrRange cidr, xmlData xml, currency money)"
  steampipe query "INSERT INTO all_columns (nullcolumn, booleancolumn, textcolumn1, textcolumn2, textcolumn3, integercolumn1, integercolumn2, integercolumn3, integercolumn4, integercolumn5, numericColumn, realColumn, floatcolumn, date1, time1, timestamp1, timestamp2, interval1, array1, jsondata, jsondata2, uuidcolumn, ipAddress, macAddress, cidrRange, xmlData, currency) VALUES (NULL, TRUE, 'Yes', 'test for varchar', 'This is a very long text for the PostgreSQL text column', 3278, 21445454, 2147483645, 92233720368547758, 922337203685477580, 23.5141543, 4660.33777, 4.6816421254887534, '1978-02-05', '08:00:00', '2016-06-22 19:10:25-07', '2016-06-22 19:10:25-07', '1 year 2 months 3 days', '{\"(408)-589-5841\"}','{ \"customer\": \"John Doe\", \"items\": {\"product\": \"Beer\",\"qty\": 6}}', '{ \"customer\": \"John Doe\", \"items\": {\"product\": \"Beer\",\"qty\": 6}}', '6948DF80-14BD-4E04-8842-7668D9C001F5', '192.168.0.0', '08:00:2b:01:02:03', '10.1.2.3/32', '<book><title>Manual</title><chapter>...</chapter></book>', 922337203685477.57)"
  run steampipe query "select * from all_columns" --output json
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_6 files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_6.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""
  
  rm -f $TEST_DATA_DIR/actual_1.json
  run steampipe query "drop table all_columns"
}

@test "query json" {
  run steampipe query "select 1 as val, 2 as col" --output json
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_query_json files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_query_json.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""
  
  rm -f $TEST_DATA_DIR/actual_1.json
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

@test "table with null values" {
  run steampipe query "select 1 as id, 2 as val1, null as val2"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_table_with_null_values.txt)"
}

@test "csv with null values" {
  run steampipe query --output csv "select 1 as id, 2 as val1, null as val2"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_csv_with_null_values.csv)"
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

@test "verify system-ingestible format(json) values are unchanged" {
  skip "TODO: reenable this test after fixing the issue with FDW acceptance tests - https://github.com/turbot/steampipe-postgres-fdw/issues/571"
  run steampipe query --output json "select 100000 as id"
  id=$(echo $output | jq '.rows.[0].id')
  assert_equal "$id" "100000"
}

@test "verify system-ingestible formats(csv) values are unchanged" {
  run steampipe query --output csv "select 100000 as id"
  assert_equal "$output" "id
100000"
}

@test "json" {
  run steampipe query --output json "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'"
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_json files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_json.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""
  
  rm -f $TEST_DATA_DIR/actual_1.json
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

@test "sql file" {
  run steampipe query $FILE_PATH/test_data/mods/sample_workspace/query/named_query_7.sql
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_sql_file.txt)"
}

@test "sql file(not found)" {
  run steampipe query $FILE_PATH/test_files/workspace_folder/query_folder/named_query_70.sql
  assert_equal "$output" "Error: file '$FILE_PATH/test_files/workspace_folder/query_folder/named_query_70.sql' does not exist"
}

@test "verify fetch and hydrate data are populated with timing enabled" {
  run steampipe query --timing "select id, string_column, json_column from chaos.chaos_all_column_types where id='0'" 
  assert_output --partial "Time"
  assert_output --partial "Rows fetched"
  assert_output --partial "Hydrate calls"
}

@test "verify empty json result is empty list and not null" {
  run steampipe query "select * from steampipe_connection where plugin = 'random'" --output json
  echo $output > $TEST_DATA_DIR/actual_1.json

  # verify that the json contents of actual_1 and expected_query_empty_json files are the same
  run jd -f patch $TEST_DATA_DIR/actual_1.json $TEST_DATA_DIR/expected_query_empty_json.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""
  
  rm -f $TEST_DATA_DIR/actual_1.json
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}