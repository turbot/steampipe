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
    run steampipe query "select * from chaos_group.chaos_transforms order by id" --output json
    # since the aggregator connection `chaos_group` contains two chaos connections,
    # we expect twice the number of rows
    assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_aggregated_result.json)"
}

@test "steampipe aggregator connections should fail when querying a different plugin" {
    run steampipe query "select * from chaos_group.chaos_transforms order by id"
    # the above query should pass since the aggregator contains only chaos connections
    0
    run steampipe query "select * from chaos_group.steampipe_registry_plugin order by id"
    # the above query should fail since the aggregator contains only chaos connections,
    # and we are querying from a steampipe table
    assert_failure
}
