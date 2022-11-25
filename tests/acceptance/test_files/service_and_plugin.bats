load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe plugin help is displayed when no sub command given" {
  steampipe plugin > test.txt

  # checking for OS type, since sed command is different for linux and OSX
  # removing lines, since they contain absolute file paths
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".txt" "36d" test.txt
    run sed -i ".txt" "37d" test.txt
  else
    run sed -i "36d" test.txt
    run sed -i "37d" test.txt
  fi

  assert_equal "$(cat test.txt)" "$(cat $TEST_DATA_DIR/expected_plugin_help_output.txt)"
  rm -f test.txt*
}

@test "steampipe service help is displayed when no sub command given" {
  steampipe service > test.txt

  # checking for OS type, since sed command is different for linux and OSX
  # removing lines, since they contain absolute file paths
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".txt" "22d" test.txt
    run sed -i ".txt" "23d" test.txt
  else
    run sed -i "22d" test.txt
    run sed -i "23d" test.txt
  fi

  assert_equal "$(cat test.txt)" "$(cat $TEST_DATA_DIR/expected_service_help_output.txt)"
  rm -f test.txt*
}

@test "steampipe service start" {
    run steampipe service start
    assert_success
}

@test "steampipe service restart" {
    run steampipe service restart
    assert_success
}

@test "steampipe service stop" {
    run steampipe service stop
    assert_success
}

@test "custom database name" {
  # Set the STEAMPIPE_INITDB_DATABASE_NAME env variable
  export STEAMPIPE_INITDB_DATABASE_NAME="custom_db_name"
  
  target_install_directory=$(mktemp -d)
  
  # Start the service
  run steampipe service start --install-dir $target_install_directory
  echo $output
  # Check if database name in the output is the same
  assert_output --partial 'custom_db_name'
  
  # Extract password from the state file
  db_name=$(cat $target_install_directory/internal/steampipe.json | jq .database)
  echo $db_name
  
  # Both should be equal
  assert_equal "$db_name" "\"custom_db_name\""
  
  run steampipe service stop --install-dir $target_install_directory
  
  rm -rf $target_install_directory
}

@test "custom database name - should not start with uppercase characters" {
  # Set the STEAMPIPE_INITDB_DATABASE_NAME env variable
  export STEAMPIPE_INITDB_DATABASE_NAME="Custom_db_name"
  
  target_install_directory=$(mktemp -d)
  
  # Start the service
  run steampipe service start --install-dir $target_install_directory
  
  assert_failure
  run steampipe service stop --force
  rm -rf $target_install_directory
}

@test "start service, install plugin and query" {
  # start service
  steampipe service start

  # install plugin
  steampipe plugin install chaos

  # query the plugin
  run steampipe query "select time_col from chaos_cache_check limit 1"
  # check if the query passes
  assert_success

  # stop service
  steampipe service stop

  # check service status
  run steampipe service status

  assert_output "$output" "Service is not running"
}

@test "start service and verify that passwords stored in .passwd and steampipe.json are same" {
  # Start the service
  run steampipe service start

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .password)
  echo $state_file_pass

  # Extract password stored in .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  pass_file_pass=\"${pass_file_pass}\"
  echo "$pass_file_pass"

  # Both should be equal
  assert_equal "$state_file_pass" "$pass_file_pass"

  run steampipe service stop
}

@test "start service with --database-password flag and verify that the password used in flag and stored in steampipe.json are same" {
  # Start the service with --database-password flag
  run steampipe service start --database-password "abcd-efgh-ijkl"

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .password)
  echo $state_file_pass

  # Both should be equal
  assert_equal "$state_file_pass" "\"abcd-efgh-ijkl\""

  run steampipe service stop
}

@test "start service with password in env variable and verify that the password used in env and stored in steampipe.json are same" {
  # Set the STEAMPIPE_DATABASE_PASSWORD env variable
  export STEAMPIPE_DATABASE_PASSWORD="dcba-hgfe-lkji"

  # Start the service
  run steampipe service start

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .password)
  echo $state_file_pass

  # Both should be equal
  assert_equal "$state_file_pass" "\"dcba-hgfe-lkji\""

  run steampipe service stop
}

@test "start service with --database-password flag and env variable set, verify that the password used in flag gets higher precedence and is stored in steampipe.json" {
  # Set the STEAMPIPE_DATABASE_PASSWORD env variable
  export STEAMPIPE_DATABASE_PASSWORD="dcba-hgfe-lkji"

  # Start the service with --database-password flag
  run steampipe service start --database-password "abcd-efgh-ijkl"

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .password)
  echo $state_file_pass

  # Both should be equal
  assert_equal "$state_file_pass" "\"abcd-efgh-ijkl\""

  run steampipe service stop
}

@test "start service after removing .passwd file, verify new .passwd file gets created and also passwords stored in .passwd and steampipe.json are same" {
  # Remove the .passwd file
  rm -f $STEAMPIPE_INSTALL_DIR/internal/.passwd

  # Start the service
  run steampipe service start

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .password)
  echo $state_file_pass

  # Extract password stored in new .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  pass_file_pass=\"${pass_file_pass}\"
  echo "$pass_file_pass"

  # Both should be equal
  assert_equal "$state_file_pass" "$pass_file_pass"

  run steampipe service stop
}

@test "start service with --database-password flag and verify that the password used in flag is not stored in .passwd file" {
  # Start the service with --database-password flag
  run steampipe service start --database-password "abcd-efgh-ijkl"

  # Extract password stored in .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  echo "$pass_file_pass"

  # Both should not be equal
  if [[ "$pass_file_pass" != "abcd-efgh-ijkl" ]]
  then
    temp=1
  fi

  assert_equal "$temp" "1"

  run steampipe service stop
}

@test "start service with password in env variable and verify that the password used in env is not stored in .passwd file" {
  # Set the STEAMPIPE_DATABASE_PASSWORD env variable
  export STEAMPIPE_DATABASE_PASSWORD="dcba-hgfe-lkji"

  # Start the service
  run steampipe service start

  # Extract password stored in .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  echo "$pass_file_pass"

  # Both should not be equal
  if [[ "$pass_file_pass" != "dcba-hgfe-lkji" ]]
  then
    temp=1
  fi

  assert_equal "$temp" "1"
  
  run steampipe service stop
}

@test "steampipe plugin list" {
    run steampipe plugin list
    assert_success
}

## connection config

@test "steampipe aggregator connection wildcard check" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos_agg.spc
    run steampipe query "select * from chaos_group.chaos_all_column_types"
    assert_success
}

@test "steampipe aggregator connection check total results" {
    run steampipe query "select * from chaos.chaos_all_numeric_column" --output json

    # store the length of the result when queried using `chaos` connection
    length_chaos=$(echo $output | jq length)

    run steampipe query "select * from chaos2.chaos_all_numeric_column" --output json

    # store the length of the result when queried using `chaos2` connection
    length_chaos_2=$(echo $output | jq length)

    run steampipe query "select * from chaos_group.chaos_all_numeric_column" --output json

    # store the length of the result when queried using `chaos_group` aggregated connection
    length_chaos_agg=$(echo $output | jq length)

    # since the aggregator connection `chaos_group` contains two chaos connections, we expect
    # the number of results returned will be the summation of the two
    assert_equal "$length_chaos_agg" "$((length_chaos+length_chaos_2))"
}

@test "steampipe aggregator connections should fail when querying a different plugin" {
    run steampipe query "select * from chaos_group.chaos_all_numeric_column order by id"

    # this should pass since the aggregator contains only chaos connections
    assert_success
    
    run steampipe query "select * from chaos_group.steampipe_registry_plugin order by id"

    # this should fail since the aggregator contains only chaos connections, and we are
    # querying a steampipe table
    assert_failure
}

@test "steampipe json connection config" {
    cp $SRC_DATA_DIR/chaos2.json $STEAMPIPE_INSTALL_DIR/config/chaos2.json

    run steampipe query "select time_col from chaos4.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos2.json

    assert_success
}

@test "steampipe should return an error for duplicate connection name" {
    cp $SRC_DATA_DIR/chaos.json $STEAMPIPE_INSTALL_DIR/config/chaos2.json

    # this should fail because of duplicate connection name
    run steampipe query "select time_col from chaos.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos2.json

    assert_output --partial 'Error: duplicate connection name'
}

@test "steampipe yaml connection config" {
    cp $SRC_DATA_DIR/chaos2.yml $STEAMPIPE_INSTALL_DIR/config/chaos3.yml

    run steampipe query "select time_col from chaos5.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos3.yml

    assert_success
}

@test "steampipe test connection config with options(hcl)" {
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    run steampipe query "select time_col from chaos6.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    assert_success
}

@test "steampipe test connection config with options(yml)" {
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    run steampipe query "select time_col from chaos6.chaos_cache_check"
    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    assert_success
}

@test "steampipe test connection config with options(json)" {
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    run steampipe query "select time_col from chaos6.chaos_cache_check"
    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    assert_success
}

@test "steampipe check regions in connection config is being parsed and used(hcl)" {
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    # check regions in connection config is being parsed and used
    run steampipe query "select * from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
    # check output
    assert_equal "$result" '[{"_ctx":{"connection_name":"chaos6"},"id":0,"region_name":"us-east-1"},{"_ctx":{"connection_name":"chaos6"},"id":3,"region_name":"us-west-2"}]'

}

@test "steampipe check regions in connection config is being parsed and used(yml)" {
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    # check regions in connection config is being parsed and used
    run steampipe query "select * from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml
    # check output
    assert_equal "$result" '[{"_ctx":{"connection_name":"chaos6"},"id":0,"region_name":"us-east-1"},{"_ctx":{"connection_name":"chaos6"},"id":3,"region_name":"us-west-2"}]'

}

@test "steampipe check regions in connection config is being parsed and used(json)" {
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    # check regions in connection config is being parsed and used
    run steampipe query "select * from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json
    # check output
    assert_equal "$result" '[{"_ctx":{"connection_name":"chaos6"},"id":0,"region_name":"us-east-1"},{"_ctx":{"connection_name":"chaos6"},"id":3,"region_name":"us-west-2"}]'

}

@test "connection name escaping" {
    cp $SRC_DATA_DIR/chaos_conn_name_escaping.spc $STEAMPIPE_INSTALL_DIR/config/chaos_conn_name_escaping.spc

    # steampipe should accept default keyword in the connection configuration file, keywords should be escaped properly
    run steampipe query "select * from \"default\".chaos_limit limit 1"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_conn_name_escaping.spc

    assert_success
}

## service extensions

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
  run steampipe query "SELECT * FROM normal_rand(10, 5, 3);"

  # previous query should pass
  assert_success
}

@test "cleanup" {
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_agg.spc
  run steampipe plugin uninstall steampipe
  rm -f $STEAMPIPE_INSTALL_DIR/config/steampipe.spc
}
