load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe install" {
    run steampipe query "select 1 as val"
    assert_success
}