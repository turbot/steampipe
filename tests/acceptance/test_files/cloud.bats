load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# These set of tests are skipped locally
# To run these tests locally set the SPIPETOOLS_PG_CONN_STRING and SPIPETOOLS_TOKEN env vars.
# These tests will be skipped locally unless both of the above env vars are set.

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

@test "install a large mod, query and check if time taken is less than 20s" {
  # using bash's built-in time, set the timeformat to seconds
  TIMEFORMAT=%R

  # create a directory to install the mods
  target_directory=$(mktemp -d)
  cd $target_directory

  # install steampipe-mod-aws-compliance
  steampipe mod install github.com/turbot/steampipe-mod-aws-compliance
  # go to the mod directory and run steampipe query
  cd .steampipe/mods/github.com/turbot/steampipe-mod-aws-compliance@*

  # max time to query(we expect it to be less than 20s)
  TIME_TO_QUERY=20
  # find the query time
  QUERY_TIME=$(time (run steampipe query "query.ec2_instance_detailed_monitoring_enabled" --workspace-database $SPIPETOOLS_PG_CONN_STRING --output json >/dev/null 2>&1) 2>&1)
  echo $QUERY_TIME

  assert_equal "$(echo $QUERY_TIME '<' $TIME_TO_QUERY | bc -l)" "1"
}

function setup() {
  
  if [[ -z "${SPIPETOOLS_PG_CONN_STRING}" ||  -z "${SPIPETOOLS_TOKEN}" ]]; then
    skip
  else
    echo "Both SPIPETOOLS_PG_CONN_STRING and SPIPETOOLS_TOKEN are set..."
  fi
}
