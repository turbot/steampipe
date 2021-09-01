load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe install" {
    run steampipe query "select 1 as val"
    assert_success
}

@test "steampipe plugin help is displayed when no sub command given" {
    run steampipe plugin
    assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_help_output.txt)"
}

@test "steampipe service help is displayed when no sub command given" {
    run steampipe service
    assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_service_help_output.txt)"
}

# Check that when disabled in config, we do not perform HTTP requests for update checks, 
# but we perform other scheduled operations
@test "scheduled task run - no update check when disabled in config - TEST DISABLED" {
#  mkdir -p $STEAMPIPE_INSTALL_DIR/internal
#  mkdir -p $STEAMPIPE_INSTALL_DIR/config
#  mkdir -p $STEAMPIPE_INSTALL_DIR/logs
#  
#  echo "" > $STEAMPIPE_INSTALL_DIR/internal/update-check.json
#  
#  # set the `lastChecked` date in the update-check.json file to a past date
#  echo $(cat $STEAMPIPE_INSTALL_DIR/internal/update-check.json | jq '.lastChecked="2021-04-10T17:53:40+05:30"') > $STEAMPIPE_INSTALL_DIR/internal/update-check.json
#  
#  # extract the content of the current state file
#  checkFileContent=$(cat $STEAMPIPE_INSTALL_DIR/internal/update-check.json)
#    
#  # put in the config file with update disabled
#  cp ${SRC_DATA_DIR}/update_check_disabled.spc $STEAMPIPE_INSTALL_DIR/config/default.spc
#  
#  # put a dummy file for log - which should get deleted
#  touch $STEAMPIPE_INSTALL_DIR/logs/database-2021-03-16.log
#
#  # setup trace logging
#  STEAMPIPE_LOG=TRACE
#
#  # run steampipe
#  run steampipe plugin list
#
#  # verify update request HTTP call was not made - the following TRACE output SHOULD NOT appear: "Sending HTTP Request"
#  [ $(echo $output | grep "Sending HTTP Request" | wc -l | tr -d ' ') -eq 0 ]
#
#  # get the content of the new update-check.json file
#  newCheckFileContent=$(cat $STEAMPIPE_INSTALL_DIR/internal/update-check.json)
#  
#  # verify that the last check time was not updated.
#  assert_equal "$(echo $checkFileContent | jq '.lastChecked')" "$(echo $newCheckFileContent | jq '.lastChecked')"
#
}

# Check that when disabled in environment, we do not perform HTTP requests for update checks, 
# but we perform other scheduled operations
#@test "scheduled task run - no update check when disabled in ENV - TEST DISABLED" {
#  # set the `lastChecked` date in the update-check.json file to a past date
#  echo $(cat $STEAMPIPE_INSTALL_DIR/internal/update-check.json| jq '.lastChecked="2021-04-10T17:53:40+05:30"') > $STEAMPIPE_INSTALL_DIR/internal/update-check.json
#  
#  # extract the content of the current state file
#  checkFileContent=$(cat $STEAMPIPE_INSTALL_DIR/internal/update-check.json)
#  
#  # update ENV to disable update check
#  echo "" > $STEAMPIPE_INSTALL_DIR/config/default.spc
#  STEAMPIPE_UPDATE_CHECK=false
#  
#  # put a dummy file for log - which should get deleted
#  touch $STEAMPIPE_INSTALL_DIR/logs/database-2021-03-16.log
#
#  # setup trace logging
#  STEAMPIPE_LOG=TRACE
#
#  # run steampipe
#  run steampipe plugin list
#
#  # verify update request HTTP call was not made - the following TRACE output SHOULD NOT appear: "Sending HTTP Request"
#  [ $(echo $output | grep "Sending HTTP Request" | wc -l | tr -d ' ') -eq 0 ]
#
#  # get the content of the new update-check.json file
#  newCheckFileContent=$(cat $STEAMPIPE_INSTALL_DIR/internal/update-check.json)
#  
#  # verify that the last check time was not updated.
#  assert_equal "$(echo $checkFileContent | jq '.lastChecked')" "$(echo $newCheckFileContent | jq '.lastChecked')"
#
#}
