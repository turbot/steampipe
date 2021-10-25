load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe plugin install" {
    run steampipe plugin install chaos
    assert_success
    run steampipe plugin uninstall chaos
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos.spc
}

@test "steampipe plugin list" {
    run steampipe plugin list
    assert_success
}

@test "steampipe plugin install chaos(spepcific version)" {
    run steampipe plugin install chaos@0.0.6
    assert_output --partial 'chaos@0.0.6'
    run steampipe plugin uninstall chaos@0.0.6
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos.spc
}

@test "steampipe plugin install chaos(latest)" {
    run steampipe plugin install chaos
    run steampipe plugin list
    assert_output --partial 'chaos@latest'
    run steampipe plugin uninstall chaos
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos.spc
}

@test "steampipe plugin install chaos(spepcific version) should not create conn config file" {
    # run steampipe plugin install chaos@0.0.6
    # cd $STEAMPIPE_INSTALL_DIR/config
    # run ls -al
    # assert_output 
    # run steampipe plugin uninstall chaos@0.0.6
}

@test "steampipe plugin install chaos(latest) should create conn config file" {
    run steampipe plugin install chaos
    cd $STEAMPIPE_INSTALL_DIR/config
    run ls -al
    assert_output --partial 'chaos.spc'
    cd -
}
