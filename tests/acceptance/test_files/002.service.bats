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
    assert_equal $(netstat -an tcp | grep LISTEN | grep tcp | grep 8765 | wc -l) 2
    steampipe service stop --force
}

@test "steampipe service start --database-listen local --database-port 8765" {
    run steampipe service start --database-listen local --database-port 8765
    assert_equal $(netstat -an tcp | grep LISTEN | grep tcp | grep 8765 | wc -l) 2
    assert_equal $(netstat -an tcp | grep LISTEN | grep tcp | grep 127.0.0.1 | grep 8765 | wc -l) 1
    assert_equal $(netstat -an tcp | grep LISTEN | grep tcp | grep ::1 | grep 8765 | wc -l) 1
    steampipe service stop --force
}