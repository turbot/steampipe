load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

## workspace tests

@test "nothing set - only defaults" {
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
  # set the ENV to get the diagnostics
  export STEAMPIPE_DIAGNOSTICS=config_json
  # set the ENV to the default workspace profile location
  export STEAMPIPE_WORKSPACE_PROFILES_LOCATION=$DEFAULT_WORKSPACE_PROFILE_LOCATION
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
  assert_equal "$installdir" '"/Users/pskrbasu/.steampipe"'
  assert_equal "$modlocation" '"/Users/pskrbasu/turbot-delivery/Steampipe/steampipe"'
  assert_equal "$snapshotlocation" '"/Users/pskrbasu/turbot-delivery/Steampipe/steampipe"'
  assert_equal "$workspace" '"default"'
  assert_equal "$workspacedatabase" '"fk43e7"'

  unset STEAMPIPE_WORKSPACE_PROFILES_LOCATION
}