load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "start service and verify that passwords stored in .passwd and steampipe.json are same" {
  # Start the service
  run steampipe service start

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .Password)
  echo $state_file_pass

  # Extract password stored in .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  pass_file_pass=\"${pass_file_pass}\"
  echo "$pass_file_pass"

  # Both should be equal
  assert_equal "$state_file_pass" "$pass_file_pass"

  run steampipe service stop
}

@test "start service with --database-password flag and verify that the password used in flag and stored in steampipe.json are same" {
  # Start the service with --database-password flag
  run steampipe service start --database-password "abcd-efgh-ijkl"

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .Password)
  echo $state_file_pass

  # Both should be equal
  assert_equal "$state_file_pass" "\"abcd-efgh-ijkl\""

  run steampipe service stop
}

@test "start service with password in env variable and verify that the password used in env and stored in steampipe.json are same" {
  # Set the STEAMPIPE_DATABASE_PASSWORD env variable
  export STEAMPIPE_DATABASE_PASSWORD="dcba-hgfe-lkji"

  # Start the service
  run steampipe service start

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .Password)
  echo $state_file_pass

  # Both should be equal
  assert_equal "$state_file_pass" "\"dcba-hgfe-lkji\""

  run steampipe service stop
}

@test "start service with --database-password flag and env variable set, verify that the password used in flag gets higher precedence and is stored in steampipe.json" {
  # Set the STEAMPIPE_DATABASE_PASSWORD env variable
  export STEAMPIPE_DATABASE_PASSWORD="dcba-hgfe-lkji"

  # Start the service with --database-password flag
  run steampipe service start --database-password "abcd-efgh-ijkl"

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .Password)
  echo $state_file_pass

  # Both should be equal
  assert_equal "$state_file_pass" "\"abcd-efgh-ijkl\""

  run steampipe service stop
}

@test "start service after removing .passwd file, verify new .passwd file gets created and also passwords stored in .passwd and steampipe.json are same" {
  # Remove the .passwd file
  rm -f $STEAMPIPE_INSTALL_DIR/internal/.passwd

  # Start the service
  run steampipe service start

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .Password)
  echo $state_file_pass

  # Extract password stored in new .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  pass_file_pass=\"${pass_file_pass}\"
  echo "$pass_file_pass"

  # Both should be equal
  assert_equal "$state_file_pass" "$pass_file_pass"

  run steampipe service stop
}

@test "start service with --database-password flag and verify that the password used in flag is not stored in .passwd file" {
  # Start the service with --database-password flag
  run steampipe service start --database-password "abcd-efgh-ijkl"

  # Extract password stored in .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  echo "$pass_file_pass"

  # Both should not be equal
  if [[ "$pass_file_pass" != "abcd-efgh-ijkl" ]]
  then
    temp=1
  fi

  assert_equal "$temp" "1"

  run steampipe service stop
}

@test "start service with password in env variable and verify that the password used in env is not stored in .passwd file" {
  # Set the STEAMPIPE_DATABASE_PASSWORD env variable
  export STEAMPIPE_DATABASE_PASSWORD="dcba-hgfe-lkji"

  # Start the service
  run steampipe service start

  # Extract password stored in .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  echo "$pass_file_pass"

  # Both should not be equal
  if [[ "$pass_file_pass" != "dcba-hgfe-lkji" ]]
  then
    temp=1
  fi

  assert_equal "$temp" "1"
  
  run steampipe service stop
}