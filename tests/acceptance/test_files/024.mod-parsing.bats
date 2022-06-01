load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "mod parsing" {
  # create a directory to install the mods
  target_directory=$(mktemp -d)
  cd $target_directory

  # install steampipe-mod-aws-compliance
  steampipe mod install github.com/turbot/steampipe-mod-aws-compliance

  # go to the mod directory and run steampipe query to verify parsing
  cd .steampipe/mods/github.com/turbot/steampipe-mod-aws-compliance@*
  run steampipe query "select 1"
  assert_success
  cd -

  # install steampipe-mod-aws-thrifty
  steampipe mod install github.com/turbot/steampipe-mod-aws-thrifty

  # go to the mod directory and run steampipe query to verify parsing
  cd .steampipe/mods/github.com/turbot/steampipe-mod-aws-thrifty@*
  run steampipe query "select 1"
  assert_success
  cd -

  # install steampipe-mod-ibm-insights
  steampipe mod install github.com/turbot/steampipe-mod-ibm-insights

  # go to the mod directory and run steampipe query to verify parsing
  cd .steampipe/mods/github.com/turbot/steampipe-mod-ibm-insights@*
  run steampipe query "select 1"
  assert_success
  cd -

  # install steampipe-mod-oci-compliance
  steampipe mod install github.com/turbot/steampipe-mod-oci-compliance

  # go to the mod directory and run steampipe query to verify parsing
  cd .steampipe/mods/github.com/turbot/steampipe-mod-oci-compliance@*
  run steampipe query "select 1"
  assert_success
  cd -

  # install steampipe-mod-azure-compliance
  steampipe mod install github.com/turbot/steampipe-mod-azure-compliance

  # go to the mod directory and run steampipe query to verify parsing
  cd .steampipe/mods/github.com/turbot/steampipe-mod-azure-compliance@*
  run steampipe query "select 1"
  assert_success
  cd -
}

function setup() {
  # install necessary plugins
  steampipe plugin install aws
  steampipe plugin install ibm
  steampipe plugin install oci
  steampipe plugin install azure
}

function teardown() {
  # remove the directory
  cd ..
  rm -rf $target_directory
}
