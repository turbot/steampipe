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

#upodate

# @test "steampipe service start --database-port 8765" {
#     run steampipe service start --database-port 8765
#     assert_equal $(netstat -an tcp | grep LISTEN | grep tcp | grep 8765 | wc -l) 2
#     steampipe service stop
# }

# @test "steampipe service start --database-listen local --database-port 8765" {
#     run steampipe service start --database-listen local --database-port 8765
#     assert_equal $(netstat -an tcp | grep LISTEN | grep tcp | grep 8765 | wc -l) 2
#     assert_equal $(netstat -an tcp | grep LISTEN | grep tcp | grep 127.0.0.1 | grep 8765 | wc -l) 1
#     assert_equal $(netstat -an tcp | grep LISTEN | grep tcp | grep ::1 | grep 8765 | wc -l) 1
#     steampipe service stop
# }

@test "custom database name" {
  # Set the STEAMPIPE_INITDB_DATABASE_NAME env variable
  export STEAMPIPE_INITDB_DATABASE_NAME="custom_db_name"
  
  target_install_directory=$(mktemp -d)
  
  # Start the service
  run steampipe service start --install-dir $target_install_directory
  
  # Extract password from the state file
  db_name=$(cat $target_install_directory/internal/steampipe.json | jq .Database)
  echo $db_name
  echo $output
  
  # Both should be equal
  assert_equal "$db_name" "\"custom_db_name\""
  # Check if database name in the output is the same
  assert_output --partial 'Database: custom_db_name'
  
  run steampipe service stop --install-dir $target_install_directory
  
  rm -rf $target_install_directory
}

@test "custom database name - should not start with uppercase characters" {
  # Set the STEAMPIPE_INITDB_DATABASE_NAME env variable
  export STEAMPIPE_INITDB_DATABASE_NAME="Custom_db_name"
  
  target_install_directory=$(mktemp -d)
  
  # Start the service
  run steampipe service start --install-dir $target_install_directory
  
  assert_failure
  run steampipe service stop --force
  rm -rf $target_install_directory
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
