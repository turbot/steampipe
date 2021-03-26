load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe query chaos" {
    run steampipe query "select * from chaos.chaos_high_row_count"
    assert_success
}

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

@test "query csv header off" {
  run steampipe query "select 1 as val, 2 as col" --output csv --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_query_csv_header_off.csv)"
}

@test "query table header off" {
  run steampipe query "select 1 as val, 2 as col" --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_query_table_header_off.txt)"
}