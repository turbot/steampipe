load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe aggregator connection wildcard check" {
    run steampipe plugin install chaos
    run steampipe plugin install aws
    cp $SRC_DATA_DIR/aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
    run steampipe query "select * from chaos_group.chaos_all_column_types"
    assert_success
}
