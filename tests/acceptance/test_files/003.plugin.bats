load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe plugin install" {
    run steampipe plugin install chaos
    assert_success
}

@test "steampipe aggregator connection wildcard check" {
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc
    cat $STEAMPIPE_INSTALL_DIR/config/chaos.spc
    run steampipe query "select * from chaos_all_column_types"
    assert_success
}

@test "steampipe plugin list" {
    run steampipe plugin list
    assert_success
}

@test "steampipe plugin uninstall" {
    run steampipe plugin uninstall
    assert_success
}