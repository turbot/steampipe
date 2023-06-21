load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "verify steampipe_server_settings table" {
    run steampipe query "select * from steampipe_server_settings"
    assert_success
}
