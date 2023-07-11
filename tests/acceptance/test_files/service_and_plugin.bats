load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe plugin help is displayed when no sub command given" {
  steampipe plugin > test.txt

  # checking for OS type, since sed command is different for linux and OSX
  # removing lines, since they contain absolute file paths
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".txt" "36d" test.txt
  else
    run sed -i "36d" test.txt
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
  else
    run sed -i "22d" test.txt
  fi

  assert_equal "$(cat test.txt)" "$(cat $TEST_DATA_DIR/expected_service_help_output.txt)"
  rm -f test.txt*
}

@test "plugin install" {
  run steampipe plugin install net
  assert_success
  steampipe plugin uninstall net
}

@test "plugin install from stream" {
  run steampipe plugin install net@0.2
  assert_success
  steampipe plugin uninstall net@0.2
}

@test "plugin install from stream (prefixed with v)" {
  run steampipe plugin install net@v0.2
  assert_success
  steampipe plugin uninstall net@0.2
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

@test "steampipe plugin list works with disabled connections" {
  rm -f $STEAMPIPE_INSTALL_DIR/config/*
  cp $SRC_DATA_DIR/chaos_conn_import_disabled.spc $STEAMPIPE_INSTALL_DIR/config/chaos_conn_import_disabled.spc
  run steampipe plugin list 2>&3 1>&3
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_conn_import_disabled.spc
  assert_success
}

## connection config

@test "steampipe aggregator connection wildcard check" {
    skip
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos_agg.spc
    run steampipe query "select * from chaos_group.chaos_all_column_types"
    assert_success
}

@test "steampipe aggregator connection check total results" {
    skip
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
    skip
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
    cp $SRC_DATA_DIR/chaos.json $STEAMPIPE_INSTALL_DIR/config/chaos3.json
    
    # this should fail because of duplicate connection name
    run steampipe query "select time_col from chaos.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos2.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos3.json

    assert_output --partial 'duplicate connection name'
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

    # keywords should be escaped properly
    run steampipe query "select * from \"escape\".chaos_limit limit 1"

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

@test "plugin list - output table" {
  export STEAMPIPE_DISPLAY_WIDTH=100
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  run steampipe plugin list --install-dir $tmpdir
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_table.txt)"
}

@test "plugin list - output json" {
  export STEAMPIPE_DISPLAY_WIDTH=100
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  run steampipe plugin list --install-dir $tmpdir --output json
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_json.json)"
}

@test "plugin list - output table (with a missing plugin)" {
  export STEAMPIPE_DISPLAY_WIDTH=100
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  # uninstall a plugin but dont remove the config - to simulate the missing plugin scenario
  steampipe plugin uninstall hackernews@0.6.0 --install-dir $tmpdir
  run steampipe plugin list --install-dir $tmpdir
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_table_with_missing_plugins.txt)"
}

@test "plugin list - output json (with a missing plugin)" {
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  # uninstall a plugin but dont remove the config - to simulate the missing plugin scenario
  steampipe plugin uninstall hackernews@0.6.0 --install-dir $tmpdir
  run steampipe plugin list --install-dir $tmpdir --output json
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_json_with_missing_plugins.json)"
}

# TODO: finds other ways to simulate failed plugins

@test "plugin list - output table (with a failed plugin)" {
  export STEAMPIPE_DISPLAY_WIDTH=100
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  # remove the contents of a plugin execuatable to simulate the failed plugin scenario
  cat /dev/null > $tmpdir/plugins/hub.steampipe.io/plugins/turbot/hackernews@0.6.0/steampipe-plugin-hackernews.plugin
  run steampipe plugin list --install-dir $tmpdir
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_table_with_failed_plugins.txt)"
}

@test "plugin list - output json (with a failed plugin)" {
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  # remove the contents of a plugin binary execuatable to simulate the failed plugin scenario
  cat /dev/null > $tmpdir/plugins/hub.steampipe.io/plugins/turbot/hackernews@0.6.0/steampipe-plugin-hackernews.plugin
  run steampipe plugin list --install-dir $tmpdir --output json
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_json_with_failed_plugins.json)"
}

@test "verify that installing plugins creates individual version.json files" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile1="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  vFile2="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/version.json"
  
  [ ! -f $vFile1 ] && fail "could not find $vFile1"
  [ ! -f $vFile2 ] && fail "could not find $vFile2"
  
  rm -rf $tmpdir
}

@test "verify that backfilling of individual plugin version.json works" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile1="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  vFile2="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/version.json"
  
  file1Content=$(cat $vFile1)
  file2Content=$(cat $vFile2)
  
  # remove the individual version files
  rm -f $vFile1
  rm -f $vFile2
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $tmpdir
  
  [ ! -f $vFile1 ] && fail "could not find $vFile1"
  [ ! -f $vFile2 ] && fail "could not find $vFile2"
  
  assert_equal "$(cat $vFile1)" "$file1Content"
  assert_equal "$(cat $vFile2)" "$file2Content"
  
  rm -rf $tmpdir
}

@test "verify that backfilling of individual plugin version.json works where it is only partially backfilled" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile1="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  vFile2="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/version.json"
  
  file1Content=$(cat $vFile1)
  file2Content=$(cat $vFile2)
  
  # remove one individual version file
  rm -f $vFile1
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $tmpdir
  
  [ ! -f $vFile1 ] && fail "could not find $vFile1"
  [ ! -f $vFile2 ] && fail "could not find $vFile2"
  
  assert_equal "$(cat $vFile1)" "$file1Content"
  assert_equal "$(cat $vFile2)" "$file2Content"
  
  rm -rf $tmpdir
}

@test "verify that global plugin/versions.json is composed from individual version.json files when it is absent" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile="$tmpdir/plugins/versions.json"
  
  fileContent=$(cat $vFile)
  
  # remove global version file
  rm -f $vFile
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $tmpdir
  
  ls -la $vFile
  
  [ ! -f $vFile ] && fail "could not find $vFile"
  
  assert_equal "$(cat $vFile)" "$fileContent"
  
  rm -rf $tmpdir
}

@test "verify that global plugin/versions.json is composed from individual version.json files when it is corrupt" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile="$tmpdir/plugins/versions.json"
  fileContent=$(cat $vFile)
  
  # remove global version file
  echo "badline to corrupt versions.json" >> $vFile
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $tmpdir
  
  [ ! -f $vFile ] && fail "could not find $vFile"
  
  assert_equal "$(cat $vFile)" "$fileContent"
  
  rm -rf $tmpdir
}

@test "verify that composition of global plugin/versions.json works when an individual version.json file is corrupt" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile="$tmpdir/plugins/versions.json"  
  vFile1="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  
  # corrupt a version file
  echo "bad line to corrupt" >> $vFile1
  
  # remove global file
  rm -f $vFile
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $tmpdir

  # verify that global file got created
  [ ! -f $vFile ] && fail "could not find $vFile"
  
  rm -rf $tmpdir
}

@test "verify that plugin installed from registry are marked as 'local' when the modtime of the binary is after the install time" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success

  # wait for a couple of seconds
  sleep 2

  # touch one of the plugin binaries
  touch $tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/steampipe-plugin-net.plugin

  # run steampipe again so that the plugin version files get backfilled
  version=$(steampipe plugin list --install-dir $tmpdir --output json | jq '.installed' | jq '. | map(select(.name | contains("net@latest")))' | jq '.[0].version')

  # assert
  assert_equal "$version" '"local"'

  rm -rf $tmpdir
}

@test "verify that steampipe check should bypass plugin requirement detection if installed plugin is local" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success

  # wait for a couple of seconds
  sleep 2

  # touch one of the plugin binaries
  touch $tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/steampipe-plugin-net.plugin

  # install a mod which has a net plugin requirement
  mkdir $tmpdir/net-mod
  cd $tmpdir/net-mod
  run steampipe mod install https://github.com/turbot/steampipe-mod-net-insights
  echo $output
  cd .steampipe/mods/github.com/turbot/steampipe-mod-net-insights@v0.5.0

  # run steampipe check
  run steampipe check all

  echo $output

  # should succeed
  assert_equal 1 0

  rm -rf $tmpdir
}

@test "cleanup" {
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_agg.spc
  run steampipe plugin uninstall steampipe
  rm -f $STEAMPIPE_INSTALL_DIR/config/steampipe.spc
}

function setup_file() {
  export BATS_TEST_TIMEOUT=120
  echo "# setup_file()">&3
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}