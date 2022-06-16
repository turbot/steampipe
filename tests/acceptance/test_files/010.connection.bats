load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe aggregator connection wildcard check" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
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

@test "steampipe check options config is being parsed and used(cache=true; hcl)" {
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    # cache functionality check since cache=true in options
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export json --max-parallel 1

    # store the unique number from 1st control in `content`
    content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
    # store the unique number from 2nd control in `new_content`
    new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content
    # remove the output and the config files
    rm -f benchmark.*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    # verify that `content` and `new_content` are the same
    assert_equal "$new_content" "$content"  
}

@test "steampipe check options config is being parsed and used(cache=true; yml)" {
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    # cache functionality check since cache=true in options
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export json --max-parallel 1

    # store the unique number from 1st control in `content`
    content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
    # store the unique number from 2nd control in `new_content`
    new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content
    # remove the output and the config files
    rm -f benchmark.*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    # verify that `content` and `new_content` are the same
    assert_equal "$new_content" "$content"
}

@test "steampipe check options config is being parsed and used(cache=true; json)" {
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    # cache functionality check since cache=true in options
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export json --max-parallel 1

    # store the unique number from 1st control in `content`
    content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
    # store the unique number from 2nd control in `new_content`
    new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content
    # remove the output and the config files
    rm -f benchmark.*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    # verify that `content` and `new_content` are the same
    assert_equal "$new_content" "$content"
    
}

@test "steampipe check options config is being parsed and used(cache=false; hcl)" {
    cp $SRC_DATA_DIR/chaos_options_2.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.spc

    # cache functionality check since cache=false in options
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export json --max-parallel 1

    # store the unique number from 1st control in `content`
    content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
    # store the unique number from 2nd control in `new_content`
    new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content

    # verify that `content` and `new_content` are not the same
    if [[ "$content" == "$new_content" ]]; then
        flag=1
    else
        flag=0
    fi
    # remove the output and the config files
    rm -f benchmark.*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.spc

    assert_equal "$flag" "0"
}

@test "steampipe check options config is being parsed and used(cache=false; yml)" {
    cp $SRC_DATA_DIR/chaos_options_2.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.yml

    # cache functionality check since cache=false in options
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export json --max-parallel 1

    # store the unique number from 1st control in `content`
    content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
    # store the unique number from 2nd control in `new_content`
    new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content

    # verify that `content` and `new_content` are not the same
    if [[ "$content" == "$new_content" ]]; then
        flag=1
    else
        flag=0
    fi
    # remove the output and the config files
    rm -f benchmark.*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.yml
    
    assert_equal "$flag" "0"
}

# This test checks whether the options in the hcl connection config(chaos_options.spc) is parsed and used correctly.
# To test this behaviour:
#   1. We create a config with options passed(cache=true; cache_ttl=300)
#   2. Start the service and run the query which selects an unique value(unique_col) from the chaos_cache_check table.
#   3. Run the query twice and store the values to compare, before stopping the service.
#   4. Compare the values, both the unique values should be equal since we had cache enabled in config.
@test "steampipe query options config is being parsed and used(cache=true; hcl)" {
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    # start the service
    steampipe service start

    # cache functionality check since cache=true in options
    steampipe query "select unique_col from chaos6.chaos_cache_check where id=2" --output json > out1.json
    steampipe query "select unique_col from chaos6.chaos_cache_check where id=2" --output json > out2.json

    # stop the service
    steampipe service stop

    # store the unique number from 1st query in `content`
    content=$(cat out1.json | jq '.[].unique_col')
    # store the unique number from 2nd query in `new_content`
    new_content=$(cat out2.json | jq '.[].unique_col')
    echo $content
    echo $new_content
    # remove the output and the config files
    rm -f out*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    # verify that `content` and `new_content` are the same
    assert_equal "$new_content" "$content"  
}

# This test checks whether the options in the yml connection config(chaos_options.yml) is parsed and used correctly.
# To test this behaviour:
#   1. We create a config with options passed(cache=true; cache_ttl=300)
#   2. Start the service and run the query which selects an unique value(unique_col) from the chaos_cache_check table.
#   3. Run the query twice and store the values to compare, before stopping the service.
#   4. Compare the values, both the unique values should be equal since we had cache enabled in config.
@test "steampipe query options config is being parsed and used(cache=true; yml)" {
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    # start the service
    steampipe service start

    # cache functionality check since cache=true in options
    steampipe query "select unique_col from chaos6.chaos_cache_check where id=2" --output json > out1.json
    steampipe query "select unique_col from chaos6.chaos_cache_check where id=2" --output json > out2.json

    # stop the service
    steampipe service stop

    # store the unique number from 1st query in `content`
    content=$(cat out1.json | jq '.[].unique_col')
    # store the unique number from 2nd query in `new_content`
    new_content=$(cat out2.json | jq '.[].unique_col')
    echo $content
    echo $new_content
    # remove the output and the config files
    rm -f out*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    # verify that `content` and `new_content` are the same
    assert_equal "$new_content" "$content"
}

# This test checks whether the options in the json connection config(chaos_options.spc) is parsed and used correctly.
# To test this behaviour:
#   1. We create a config with options passed(cache=true; cache_ttl=300)
#   2. Start the service and run the query which selects an unique value(unique_col) from the chaos_cache_check table.
#   3. Run the query twice and store the values to compare, before stopping the service.
#   4. Compare the values, both the unique values should be equal since we had cache enabled in config.
@test "steampipe query options config is being parsed and used(cache=true; json)" {
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    # start the service
    steampipe service start

    # cache functionality check since cache=true in options
    steampipe query "select unique_col from chaos6.chaos_cache_check where id=2" --output json > out1.json
    steampipe query "select unique_col from chaos6.chaos_cache_check where id=2" --output json > out2.json

    # stop the service
    steampipe service stop

    # store the unique number from 1st query in `content`
    content=$(cat out1.json | jq '.[].unique_col')
    # store the unique number from 2nd query in `new_content`
    new_content=$(cat out2.json | jq '.[].unique_col')
    echo $content
    echo $new_content
    # remove the output and the config files
    rm -f out*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    # verify that `content` and `new_content` are the same
    assert_equal "$new_content" "$content"
}

# This test checks whether the options in the hcl connection config(chaos_options.spc) is parsed and used correctly.
# To test this behaviour:
#   1. We create a config with options passed(cache=false; cache_ttl=300)
#   2. Start the service and run the query which selects an unique value(unique_col) from the chaos_cache_check table.
#   3. Run the query twice and store the values to compare, before stopping the service.
#   4. Compare the values, both the unique values should different since we had cache disabled in config.
@test "steampipe query options config is being parsed and used(cache=false; hcl)" {
    cp $SRC_DATA_DIR/chaos_options_2.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.spc

    # start the service
    steampipe service start

    # cache functionality check since cache=false in options
    steampipe query "select unique_col from chaos6.chaos_cache_check where id=2" --output json > out1.json
    steampipe query "select unique_col from chaos6.chaos_cache_check where id=2" --output json > out2.json

    # stop the service
    steampipe service stop

    # store the unique number from 1st query in `content`
    content=$(cat out1.json | jq '.[].unique_col')
    # store the unique number from 2nd query in `new_content`
    new_content=$(cat out2.json | jq '.[].unique_col')

    # verify that `content` and `new_content` are not the same
    if [[ "$content" == "$new_content" ]]; then
        flag=1
    else
        flag=0
    fi
    # remove the output and the config files
    rm -f out*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.spc

    assert_equal "$flag" "0"
}

# This test checks whether the options in the yml connection config(chaos_options.spc) is parsed and used correctly.
# To test this behaviour:
#   1. We create a config with options passed(cache=false; cache_ttl=300)
#   2. Start the service and run the query which selects an unique value(unique_col) from the chaos_cache_check table.
#   3. Run the query twice and store the values to compare, before stopping the service.
#   4. Compare the values, both the unique values should different since we had cache disabled in config.
@test "steampipe query options config is being parsed and used(cache=false; yml)" {
    cp $SRC_DATA_DIR/chaos_options_2.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.yml

    # start the service
    steampipe service start

    # cache functionality check since cache=false in options
    steampipe query "select unique_col from chaos6.chaos_cache_check where id=2" --output json > out1.json
    steampipe query "select unique_col from chaos6.chaos_cache_check where id=2" --output json > out2.json

    # stop the service
    steampipe service stop

    # store the unique number from 1st query in `content`
    content=$(cat out1.json | jq '.[].unique_col')
    # store the unique number from 2nd query in `new_content`
    new_content=$(cat out2.json | jq '.[].unique_col')

    # verify that `content` and `new_content` are not the same
    if [[ "$content" == "$new_content" ]]; then
        flag=1
    else
        flag=0
    fi
    # remove the output and the config files
    rm -f out*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.yml

    assert_equal "$flag" "0"
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
