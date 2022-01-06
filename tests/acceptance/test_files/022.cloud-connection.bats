load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "test with connection string" {
    run steampipe query "select * from aws_account" --workspace-database postgresql://pskrbasu:ee7d-47fc-9672@pskrbasu-chaos.usea1.db.steampipe.io:9193/bp43e7
    assert_output --partial '"redhood-aaa"'
}
