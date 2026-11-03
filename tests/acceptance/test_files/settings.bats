load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"
load ../helpers/steampipe_processes

@test "verify steampipe_server_settings table" {
    run steampipe query "select * from steampipe_server_settings"
    assert_success
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(count_steampipe_processes)
  assert_equal $num 0
}
