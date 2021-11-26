load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# tests for tablefunc module

@test "test crosstab function" {
  # create table and insert values
  steampipe query "CREATE TABLE ct(id SERIAL, rowid TEXT, attribute TEXT, value TEXT);"
  steampipe query "INSERT INTO ct(rowid, attribute, value) VALUES('test1','att1','val1');"
  steampipe query "INSERT INTO ct(rowid, attribute, value) VALUES('test1','att2','val2');"
  steampipe query "INSERT INTO ct(rowid, attribute, value) VALUES('test1','att3','val3');"

  # crosstab function
  run steampipe query "SELECT * FROM crosstab('select rowid, attribute, value from ct where attribute = ''att2'' or attribute = ''att3'' order by 1,2') AS ct(row_name text, category_1 text, category_2 text);"
  echo $output

  # drop table
  steampipe query "DROP TABLE ct"

  # match output with expected
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_crosstab_results.txt)"
}

@test "test normal_rand function" {
  # normal_rand function
  steampipe query "SELECT * FROM normal_rand(10, 5, 3);"

  # previous query should pass
  assert_success
}
