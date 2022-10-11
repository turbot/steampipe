load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"


@test "dynamic schema - add csv and query" {
 # copy the csv file from csv source folder
 cp $SRC_DATA_DIR/csv/a.csv $FILE_PATH/test_data/csv_plugin_test/a.csv

 # run the query and verify - should pass
 run steampipe query "select * from csv1.a"
 assert_success
}

@test "dynamic schema - add another column to csv and query the new column" {
 # run the query and verify - should pass
 run steampipe query "select * from csv1.a"
 assert_success

 # remove the a.csv file
 rm -f $FILE_PATH/test_data/csv_plugin_test/a.csv

 # copy the csv file with extra column from csv source folder and give the same name(a.csv)
 cp $SRC_DATA_DIR/csv/a_extra_col.csv $FILE_PATH/test_data/csv_plugin_test/a.csv

 # query the extra column and verify - should pass
 run steampipe query 'select "column_D" from csv1.a'
 assert_success
}

@test "dynamic schema - remove the csv with extra column and query (should fail)" {
 # query the extra column and verify - should pass
 run steampipe query 'select "column_D" from csv1.a'
 assert_success

 # remove the a.csv file with extra column and copy the old one again
 rm -f $FILE_PATH/test_data/csv_plugin_test/a.csv
 cp $SRC_DATA_DIR/csv/a.csv $FILE_PATH/test_data/csv_plugin_test/a.csv

 # query the extra column and verify - should fail
 run steampipe query 'select "column_D" from csv1.a'
 assert_output --partial 'does not exist'

 rm -f $FILE_PATH/test_data/csv_plugin_test/a.csv
}

@test "dynamic schema - remove csv and query (should fail)" {
 # copy the csv file from csv source folder
 cp $SRC_DATA_DIR/csv/b.csv $FILE_PATH/test_data/csv_plugin_test/b.csv
  
 # run the query and verify - should pass
 run steampipe query "select * from csv1.b"
 assert_success

 # remove the b.csv file
 rm -f $FILE_PATH/test_data/csv_plugin_test/b.csv

# run the query and verify - should fail
 run steampipe query "select * from csv1.b"
 assert_output --partial 'does not exist'

 rm -f $FILE_PATH/test_data/csv_plugin_test/b.csv
}


function setup() {
 # install csv plugin
 run steampipe plugin install csv

 cd $SRC_DATA_DIR
 # appending the csv_plugin_test path
 full_path="${FILE_PATH}/test_data/csv_plugin_test/*.csv"
 echo "${full_path}"

 # escaping the slashes(/)
 b=$(echo -e "${full_path}" | sed -e 's/\//\\\//g')
 echo -e $b

 # reading each line from the config template and storing in a file
 while IFS= read -r line
 do
   echo "$line" >> output.spc
 done < "csv_template.spc"

 # replace the config file template with required path
 sed -i -e "s/abc/${b}/g" 'output.spc'

 # copy the new connection config
 cp output.spc $STEAMPIPE_INSTALL_DIR/config/csv1.spc
}

function teardown() {
  # remove the files created as part of these tests 
  rm -f $STEAMPIPE_INSTALL_DIR/config/csv*.spc
  rm -f output.*
}