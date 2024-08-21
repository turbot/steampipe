load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

## connection config tests

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

    steampipe query "select 1"

    run steampipe query "select time_col from chaos5.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos3.yml

    assert_success
}

@test "steampipe test connection config with options(hcl)" {
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    steampipe query "select 1"

    run steampipe query "select time_col from chaos6.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    assert_success
}

@test "steampipe test connection config with options(yml)" {
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    steampipe query "select 1"

    run steampipe query "select time_col from chaos6.chaos_cache_check"
    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    assert_success
}

@test "steampipe test connection config with options(json)" {
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    steampipe query "select 1"

    run steampipe query "select time_col from chaos6.chaos_cache_check"
    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    assert_success
}

@test "steampipe check regions in connection config is being parsed and used(hcl)" {
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    steampipe query "select 1"

    # check regions in connection config is being parsed and used
    run steampipe query "select id,region_name from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')
    # set the trimmed result as output
    run echo $result
    echo $output

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
    # check output
    assert_output --partial '[{"id":0,"region_name":"us-east-1"},{"id":3,"region_name":"us-west-2"}]'

}

@test "steampipe check regions in connection config is being parsed and used(yml)" {
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    steampipe query "select 1"

    # check regions in connection config is being parsed and used
    run steampipe query "select id,region_name from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')
    # set the trimmed result as output
    run echo $result
    echo $output

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml
    # check output
    assert_output --partial '[{"id":0,"region_name":"us-east-1"},{"id":3,"region_name":"us-west-2"}]'

}

@test "steampipe check regions in connection config is being parsed and used(json)" {
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    steampipe query "select 1"

    # check regions in connection config is being parsed and used
    run steampipe query "select id,region_name from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')
    # set the trimmed result as output
    run echo $result
    echo $output

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json
    # check output
    assert_output --partial '[{"id":0,"region_name":"us-east-1"},{"id":3,"region_name":"us-west-2"}]'

}

@test "connection name escaping" {
    cp $SRC_DATA_DIR/chaos_conn_name_escaping.spc $STEAMPIPE_INSTALL_DIR/config/chaos_conn_name_escaping.spc

    steampipe query "select 1"

    # keywords should be escaped properly
    run steampipe query "select * from \"escape\".chaos_limit limit 1"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_conn_name_escaping.spc

    assert_success
}

# This test checks this pipes bug - https://github.com/turbot/steampipe/issues/4353
# With service running, if a new connection is added but is not in search_path, it should be available and ready
# previously, the connection was in error state
# NOTE: This test should always be the last test in this file
@test "dynamic schema - service running, new connection added(but not in search_path) - connection should be available and ready" {
  steampipe plugin install servicenow --install-dir $STEAMPIPE_INSTALL_DIR
  # start service
  steampipe service start --install-dir $STEAMPIPE_INSTALL_DIR

  # update search_path in db options, to exclude the new connection
  cp $SRC_DATA_DIR/default_search_path.spc $STEAMPIPE_INSTALL_DIR/config/default.spc

	cat $STEAMPIPE_INSTALL_DIR/config/default.spc

  # add a new connection
  cp $SRC_DATA_DIR/servicenow.spc $STEAMPIPE_INSTALL_DIR/config/servicenow2.spc

  sleep 10

  # check if the new connection is available and ready
  run steampipe query "select name, state from steampipe_connection" --output csv --install-dir $STEAMPIPE_INSTALL_DIR
  assert_output --partial 'servicenow,ready'
}

@test "cleanup" {
  steampipe service stop --install-dir $STEAMPIPE_INSTALL_DIR
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_agg.spc
  run steampipe plugin uninstall steampipe
  rm -f $STEAMPIPE_INSTALL_DIR/config/steampipe.spc
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
