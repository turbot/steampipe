load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe query chaos" {
    run steampipe query "select * from chaos.chaos_high_row_count"
    assert_success
}