load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# These set of tests are skipped locally(SKIP_CLOUD_TESTS env var set in run-local.sh)
# To run these tests locally unset the SKIP_CLOUD_TESTS in run-local.sh file and set the 
# SPIPETOOLS_PG_CONN_STRING and SPIPETOOLS_TOKEN env vars.

@test "connect to cloud workspace - passing the postgres connection string to workspace-database arg" {
  # run steampipe query and fetch an account from the cloud workspace
  run steampipe query "select account_aliases from all_aws.aws_account where account_id='632902152528'" --workspace-database $SPIPETOOLS_PG_CONN_STRING --output json

  # fetch the value of account_alias to compare
  op=$(echo $output | jq '.[0].account_aliases[0]')
  echo $op

  # check if values match
  assert_equal "$op" "\"nagraj-aaa\""
}

@test "connect to cloud workspace - passing the cloud-token arg and the workspace name to workspace-database arg" {
  # run steampipe query and fetch an account from the cloud workspace
  run steampipe query "select account_aliases from all_aws.aws_account where account_id='632902152528'" --cloud-token $SPIPETOOLS_TOKEN --workspace-database spipetools/toolstest --output json

  # fetch the value of account_alias to compare
  op=$(echo $output | jq '.[0].account_aliases[0]')
  echo $op

  # check if values match
  assert_equal "$op" "\"nagraj-aaa\""
}

@test "connect to cloud workspace - passing the cloud-host arg, the cloud-token arg and the workspace name to workspace-database arg" {
  # run steampipe query and fetch an account from the cloud workspace
  run steampipe query "select account_aliases from all_aws.aws_account where account_id='632902152528'" --cloud-host "cloud.steampipe.io" --cloud-token $SPIPETOOLS_TOKEN --workspace-database spipetools/toolstest --output json

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
  assert_output --partial 'Error: Not authenticated for Steampipe Cloud.'
}

function setup() {
  echo $SKIP_CLOUD_TESTS

  if [[ $SKIP_CLOUD_TESTS == true ]]; then
    skip
  else
    echo "Test not skipped"
  fi
}