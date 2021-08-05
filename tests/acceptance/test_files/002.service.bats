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

@test "steampipe service stop should not trigger daily checks and tasks" {
    run steampipe service start

    # set the `lastChecked` date in the update-check.json file to a past date
    echo $(cat $STEAMPIPE_INSTALL_DIR/internal/update-check.json | jq '.lastChecked="2021-04-10T17:53:40+05:30"') > $STEAMPIPE_INSTALL_DIR/internal/update-check.json

    # get the content of the current update-check.json file
    checkFileContent=$(cat $STEAMPIPE_INSTALL_DIR/internal/update-check.json)

    run steampipe service stop

    # get the content of the new update-check.json file
    newCheckFileContent=$(cat $STEAMPIPE_INSTALL_DIR/internal/update-check.json)

    assert_equal "$(echo $newCheckFileContent | jq '.lastChecked')" '"2021-04-10T17:53:40+05:30"'
}
