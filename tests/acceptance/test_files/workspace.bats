load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

## workspace tests

@test "nothing set - only defaults" {
  skip
  # set the ENV to get the diagnostics
  export STEAMPIPE_DIAGNOSTICS=config_json
  run steampipe query "select 1"

  # get the resolved values of the args
  cloudhost=$(echo $output | jq '."cloud-host"')
  cloudtoken=$(echo $output | jq '."cloud-token"')
  installdir=$(echo $output | jq '."install-dir"')
  modlocation=$(echo $output | jq '."mod-location"')
  snapshotlocation=$(echo $output | jq '."snapshot-location"')
  workspace=$(echo $output | jq '."workspace"')
  workspacedatabase=$(echo $output | jq '."workspace-database"')

  # get the current working and insall directory(for local)
  cwd=$(pwd)
  install_dir=$STEAMPIPE_INSTALL_DIR

  # print values for debugging
  echo "cloud-host: $cloudhost"
  echo "cloud-token: $cloudtoken"
  echo "install-dir: $installdir"
  echo "mod-location: $modlocation"
  echo "snapshot-location: $snapshotlocation"
  echo "workspace: $workspace"
  echo "workspace-database: $workspacedatabase"

  # check with expected
  assert_equal "$cloudhost" '"cloud.steampipe.io"'
  assert_equal "$cloudtoken" '""'
  assert_equal "$installdir" \"${install_dir}\"
  assert_equal "$modlocation" \"${cwd}\"
  assert_equal "$snapshotlocation" '""'
  assert_equal "$workspace" '"default"'
  assert_equal "$workspacedatabase" '"local"'

  unset STEAMPIPE_DIAGNOSTICS
}

@test "default workspace profile value set" {
  skip
  # set the ENV to get the diagnostics
  export STEAMPIPE_DIAGNOSTICS=config_json
  # set the ENV to the default workspace profile location
  export STEAMPIPE_WORKSPACE_PROFILES_LOCATION=$FILE_PATH/test_data/source_files/workspace_profiles
  unset STEAMPIPE_INSTALL_DIR
  run steampipe query "select 1"
  echo $output

  # get the resolved values of the args
  cloudhost=$(echo $output | jq '."cloud-host"')
  cloudtoken=$(echo $output | jq '."cloud-token"')
  installdir=$(echo $output | jq '."install-dir"')
  modlocation=$(echo $output | jq '."mod-location"')
  snapshotlocation=$(echo $output | jq '."snapshot-location"')
  workspace=$(echo $output | jq '."workspace"')
  workspacedatabase=$(echo $output | jq '."workspace-database"')

  # print values for debugging
  echo "cloud-host: $cloudhost"
  echo "cloud-token: $cloudtoken"
  echo "install-dir: $installdir"
  echo "mod-location: $modlocation"
  echo "snapshot-location: $snapshotlocation"
  echo "workspace: $workspace"
  echo "workspace-database: $workspacedatabase"
  env | grep "STEAMPIPE"

  # check with expected
  assert_equal "$cloudhost" '"latestpipe.turbot.io/"'
  assert_equal "$cloudtoken" '"spt_012faketoken34567890_012faketoken3456789099999"'
  assert_equal "$installdir" '"/Users/pskrbasu/.steampipe"'
  assert_equal "$modlocation" '"/Users/pskrbasu/turbot-delivery/Steampipe/steampipe"'
  assert_equal "$snapshotlocation" '"/Users/pskrbasu/turbot-delivery/Steampipe/steampipe"'
  assert_equal "$workspace" '"default"'
  assert_equal "$workspacedatabase" '"fk43e7"'

  unset STEAMPIPE_DIAGNOSTICS
  unset STEAMPIPE_WORKSPACE_PROFILES_LOCATION
}

@test "env variables set" {
  skip
  # create a temp dir for the test
  mkdir workflow_test
  # set the ENV to get the diagnostics
  export STEAMPIPE_DIAGNOSTICS=config_json
  # set the ENV
  export STEAMPIPE_INSTALL_DIR="workflow_test"
  export STEAMPIPE_MOD_LOCATION="workflow_test"
  export STEAMPIPE_CLOUD_HOST="latestpipe.turbot.io/"
  export STEAMPIPE_CLOUD_TOKEN="spt_012faketoken34567890_012faketoken3456789099999"
  export STEAMPIPE_SNAPSHOT_LOCATION="workflow_test"
  export STEAMPIPE_WORKSPACE_DATABASE="fk43e8"

  run steampipe query "select 1"
  echo $output

  # get the resolved values of the args
  cloudhost=$(echo $output | jq '."cloud-host"')
  cloudtoken=$(echo $output | jq '."cloud-token"')
  installdir=$(echo $output | jq '."install-dir"')
  modlocation=$(echo $output | jq '."mod-location"')
  snapshotlocation=$(echo $output | jq '."snapshot-location"')
  workspace=$(echo $output | jq '."workspace"')
  workspacedatabase=$(echo $output | jq '."workspace-database"')

  # print values for debugging
  echo "cloud-host: $cloudhost"
  echo "cloud-token: $cloudtoken"
  echo "install-dir: $installdir"
  echo "mod-location: $modlocation"
  echo "snapshot-location: $snapshotlocation"
  echo "workspace: $workspace"
  echo "workspace-database: $workspacedatabase"


  # check with expected
  assert_equal "$cloudhost" '"latestpipe.turbot.io/"'
  assert_equal "$cloudtoken" '"spt_012faketoken34567890_012faketoken3456789099999"'
  assert_equal "$installdir" '"workflow_test"'
  assert_equal "$modlocation" '"workflow_test"'
  assert_equal "$snapshotlocation" '"workflow_test"'
  assert_equal "$workspace" '"default"'
  assert_equal "$workspacedatabase" '"fk43e7"'

  unset STEAMPIPE_DIAGNOSTICS
  unset STEAMPIPE_INSTALL_DIR
  unset STEAMPIPE_MOD_LOCATION
  unset STEAMPIPE_CLOUD_HOST
  unset STEAMPIPE_CLOUD_TOKEN
  unset STEAMPIPE_SNAPSHOT_LOCATION
  unset STEAMPIPE_WORKSPACE_DATABASE
  rm -rf workflow_test
}

@test "generic test" {
  # setup test folder and read the test-cases file
  cd $FILE_PATH/test_data/source_files/config_tests
  tests=$(cat workspace_tests.json)

  echo $tests | jq -c -r '.[]' | while read i; do
    unset STEAMPIPE_INSTALL_DIR
    cwd=$(pwd)

    test_name=$(echo $i | jq '.test')
    echo ">>> Running: $test_name <<<"

    # exports needed for setup
    exports=$(echo $i | jq '.setup.exports')
    echo $exports

    for exp in $(echo "${exports}" | jq -r '.[]'); do
      export $exp
    done

    # args to run with steampipe query command
    args=$(echo $i | jq '.setup.args')
    echo $args

    # get the diagnostics by running steampipe
    diagnostics=$(steampipe query "select 1" "$args")
    echo $diagnostics
    # fetch the individual values from the diagnostics(echo these variables to debug)
    d_cloud_host=$(echo $diagnostics | jq '."cloud-host"')
    d_cloud_token=$(echo $diagnostics | jq '."cloud-token"')
    d_install_dir=$(echo $diagnostics | jq '."install-dir"')
    d_install_dir=$(echo $d_install_dir | tr -d '"') # trim quotes
    d_mod_location=$(echo $diagnostics | jq '."mod-location"')
    d_mod_location=$(echo $d_mod_location | tr -d '"') # trim quotes
    d_snapshot_location=$(echo $diagnostics | jq '."snapshot-location"')
    d_workspace=$(echo $diagnostics | jq '."workspace"')
    d_workspace_database=$(echo $diagnostics | jq '."workspace-database"')

    # get expected diagnostics
    expected=$(echo $i | jq '.expected')
    echo $expected
    # fetch the individual values from the expected(echo these variables to debug)
    e_cloud_host=$(echo $expected | jq '."cloud-host"')
    e_cloud_token=$(echo $expected | jq '."cloud-token"')
    e_install_dir=$(echo $expected | jq '."install-dir"')
    e_install_dir=$(echo $e_install_dir | tr -d '"') # trim quotes
    e_install_dir="${cwd}/${e_install_dir}" # create an absolute path
    e_mod_location=$(echo $expected | jq '."mod-location"')
    e_mod_location=$(echo $e_mod_location | tr -d '"') # trim quotes
    e_mod_location="${cwd}/${e_mod_location}" # create an absolute path
    e_snapshot_location=$(echo $expected | jq '."snapshot-location"')
    e_workspace=$(echo $expected | jq '."workspace"')
    e_workspace_database=$(echo $expected | jq '."workspace-database"')

    assert_equal $d_cloud_host $e_cloud_host
    assert_equal $d_cloud_token $e_cloud_token
    assert_equal $d_install_dir $e_install_dir
    assert_equal $d_mod_location $e_mod_location
    assert_equal $d_snapshot_location $e_snapshot_location
    assert_equal $d_workspace $e_workspace
    assert_equal $d_workspace_database $e_workspace_database

    env | grep "STEAMPIPE"
  done
  cd -
}

