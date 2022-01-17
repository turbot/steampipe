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
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos2.json $STEAMPIPE_INSTALL_DIR/config/chaos2.json

    run steampipe query "select time_now from chaos4.chaos_cache_check"
    assert_success
}

@test "steampipe should return an error for duplicate connection name" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos.json $STEAMPIPE_INSTALL_DIR/config/chaos2.json

    # this should fail because of duplicate connection name
    run steampipe query "select time_now from chaos.chaos_cache_check"

    assert_output --partial 'Error: duplicate connection name'
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos2.json
}

@test "steampipe yaml connection config" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos2.yml $STEAMPIPE_INSTALL_DIR/config/chaos3.yml

    run steampipe query "select time_now from chaos5.chaos_cache_check"
    assert_success
}

@test "steampipe test connection config with options(hcl)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    run steampipe query "select time_now from chaos6.chaos_cache_check"
    assert_success
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
}

@test "steampipe test connection config with options(yml)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    run steampipe query "select time_now from chaos6.chaos_cache_check"
    assert_success
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml
}

@test "steampipe test connection config with options(json)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    run steampipe query "select time_now from chaos6.chaos_cache_check"
    assert_success
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json
}

@test "steampipe check options config is being parsed and used(cache=true; hcl)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    # cache functionality check since cache=true in options
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export json --max-parallel 1

    # store the date from 1st control in `content`
    content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
    # store the date from 2nd control in `new_content`
    new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content

    # verify that `content` and `new_content` are the same
    assert_equal "$new_content" "$content"
    
    rm -f benchmark.*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
}

@test "steampipe check options config is being parsed and used(cache=true; yml)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    # cache functionality check since cache=true in options
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export json --max-parallel 1

    # store the date from 1st control in `content`
    content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
    # store the date from 2nd control in `new_content`
    new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content

    # verify that `content` and `new_content` are the same
    assert_equal "$new_content" "$content"
    
    rm -f benchmark.*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml
}

@test "steampipe check options config is being parsed and used(cache=true; json)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    # cache functionality check since cache=true in options
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export json --max-parallel 1

    # store the date from 1st control in `content`
    content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
    # store the date from 2nd control in `new_content`
    new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content

    # verify that `content` and `new_content` are the same
    assert_equal "$new_content" "$content"
    
    rm -f benchmark.*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json
}

@test "steampipe check options config is being parsed and used(cache=false; hcl)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options_2.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.spc

    # cache functionality check since cache=false in options
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export json --max-parallel 1

    # store the date from 1st control in `content`
    content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
    # store the date from 2nd control in `new_content`
    new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content

    # verify that `content` and `new_content` are not the same
    if [[ "$content" == "$new_content" ]]; then
        flag=1
    else
        flag=0
    fi
    assert_equal "$flag" "0"

    rm -f benchmark.*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.spc
}

@test "steampipe check options config is being parsed and used(cache=false; yml)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options_2.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.yml

    # cache functionality check since cache=false in options
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export json --max-parallel 1

    # store the date from 1st control in `content`
    content=$(cat benchmark.*.json | jq '.groups[].controls[0].results[0].resource')
    # store the date from 2nd control in `new_content`
    new_content=$(cat benchmark.*.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content

    # verify that `content` and `new_content` are not the same
    if [[ "$content" == "$new_content" ]]; then
        flag=1
    else
        flag=0
    fi
    assert_equal "$flag" "0"
    
    rm -f benchmark.*.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options_2.yml
}

@test "steampipe check regions in connection config is being parsed and used(hcl)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    # check regions in connection config is being parsed and used
    run steampipe query "select * from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')

    # check output
    assert_equal "$result" '[{"id":0,"region_name":"us-east-1"},{"id":3,"region_name":"us-west-2"}]'

    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
}

@test "steampipe check regions in connection config is being parsed and used(yml)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    # check regions in connection config is being parsed and used
    run steampipe query "select * from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')

    # check output
    assert_equal "$result" '[{"id":0,"region_name":"us-east-1"},{"id":3,"region_name":"us-west-2"}]'

    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml
}

@test "steampipe check regions in connection config is being parsed and used(json)" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    # check regions in connection config is being parsed and used
    run steampipe query "select * from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')

    # check output
    assert_equal "$result" '[{"id":0,"region_name":"us-east-1"},{"id":3,"region_name":"us-west-2"}]'

    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json
}
