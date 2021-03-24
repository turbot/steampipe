load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe service start" {
    run steampipe service start
    assert_success
}

@test "steampipe service restart" {
    run steampipe service restart
    assert_success
}

@test "steampipe service stop" {
    run steampipe service stop
    assert_success
}

@test "steampipe service start --database-port 8765" {
    run steampipe service start --database-port 8765
    assert_equal "$(netstat -an tcp | grep LISTEN | grep tcp | grep 8765)" "$(cat $TEST_DATA_DIR/expected_service_start_port.txt)"
    steampipe service stop --force
}

@test "steampipe service start --database-listen local --database-port 8765" {
    run steampipe service start --database-listen local --database-port 8765
    assert_equal "$(netstat -an tcp | grep LISTEN | grep tcp | grep 8765)" "$(cat $TEST_DATA_DIR/expected_service_start_listen_local.txt)" # for tcp4 and tcp6
    steampipe service stop --force
}