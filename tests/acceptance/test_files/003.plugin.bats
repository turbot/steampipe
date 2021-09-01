load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe plugin install" {
    run steampipe plugin install chaos
    assert_success
}

@test "steampipe plugin list" {
    run steampipe plugin list
    assert_success
}

#@test "steampipe plugin uninstall" {
#    run steampipe plugin uninstall
#    assert_success
#}