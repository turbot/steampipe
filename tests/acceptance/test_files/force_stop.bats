load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# This set of tests should always be the last acceptance tests

@test "start errors nicely after state file deletion" {
  run steampipe service start

  # Delete the state file
  rm -f $STEAMPIPE_INSTALL_DIR/internal/steampipe.json

  # Trying to start the service should fail, check the error message
  run steampipe service start
  echo $output
  assert_output --partial 'service is running in an unknown state'

  # Trying to stop the service should fail, check the error message
  run steampipe service stop
  echo $output
  assert_output --partial 'service is running in an unknown state'
}

@test "force stop works after state file deletion" {
  run steampipe service start

  # Delete the state file
  rm -f $STEAMPIPE_INSTALL_DIR/internal/steampipe.json

  # Trying to start the service should fail
  run steampipe service start
  assert_failure

  # Trying to stop the service should fail
  run steampipe service stop
  assert_failure

  # Force stopping the service should work
  run steampipe service stop --force
  assert_success
}
