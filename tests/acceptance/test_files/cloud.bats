load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# These set of tests are skipped locally
# To run these tests locally set the SPIPETOOLS_PG_CONN_STRING and SPIPETOOLS_TOKEN env vars.
# These tests will be skipped locally unless both of the above env vars are set.

@test "connect to cloud workspace - passing the postgres connection string to workspace-database arg" {
  # run steampipe query and fetch an account from the cloud workspace
  run steampipe query "select account_aliases from all_aws.aws_account where account_id='632902152528'" --workspace-database $SPIPETOOLS_PG_CONN_STRING --output json
  echo $output

  # fetch the value of account_alias to compare
  op=$(echo $output | jq '.[0].account_aliases[0]')
  echo $op

  # check if values match
  assert_equal "$op" "\"nagraj-aaa\""
}

@test "connect to cloud workspace - passing the cloud-token arg and the workspace name to workspace-database arg" {
  # run steampipe query and fetch an account from the cloud workspace
  run steampipe query "select account_aliases from all_aws.aws_account where account_id='632902152528'" --cloud-token $SPIPETOOLS_TOKEN --workspace-database spipetools/toolstest --output json
  echo $output

  # fetch the value of account_alias to compare
  op=$(echo $output | jq '.[0].account_aliases[0]')
  echo $op

  # check if values match
  assert_equal "$op" "\"nagraj-aaa\""
}

@test "connect to cloud workspace - passing the cloud-host arg, the cloud-token arg and the workspace name to workspace-database arg" {
  # run steampipe query and fetch an account from the cloud workspace
  run steampipe query "select account_aliases from all_aws.aws_account where account_id='632902152528'" --cloud-host "pipes.turbot.com" --cloud-token $SPIPETOOLS_TOKEN --workspace-database spipetools/toolstest --output json
  echo $output

  # fetch the value of account_alias to compare
  op=$(echo $output | jq '.[0].account_aliases[0]')
  echo $op

  # check if values match
  assert_equal "$op" "\"nagraj-aaa\""
}

@test "connect to cloud workspace(FAILED TO CONNECT) - passing wrong postgres connection string to workspace-database arg" {
  # run steampipe query using wrong connection string
  run steampipe query "select account_aliases from all_aws.aws_account where account_id='632902152528'" --workspace-database abcd --output json
  echo $output

  # check the error message
  assert_output --partial 'Error: Not authenticated for Turbot Pipes.'
}

@test "connect to cloud workspace - passing the workspace name to workspace-database arg (unsetting ENV - the token should get picked from tptt file)" {
  # write the pipes.turbot.com.tptt file in internal
  # write the token to the file
  file_name=$STEAMPIPE_INSTALL_DIR/internal/pipes.turbot.com.tptt
  echo $SPIPETOOLS_TOKEN > $file_name

  cat $file_name
  
  STEAMPIPE_CONFIG_DUMP=config_json steampipe query "select 1"

  # this step will create snapshots in the workspace - but that's ok
  # workspaces expire snapshots anyway
  run steampipe query "select 1" --share
  echo $output

  assert_success
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}

function setup() {
  if [[ -z "${SPIPETOOLS_PG_CONN_STRING}" ||  -z "${SPIPETOOLS_TOKEN}" ]]; then
    skip
  else
    echo "Both SPIPETOOLS_PG_CONN_STRING and SPIPETOOLS_TOKEN are set..."
  fi
}
