load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# This test should always be the last acceptance test
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