load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

### require tests ###

@test "running steampipe query with mod plugin requirement not met" {
  cd $FILE_PATH/test_data/mods/bad_mod_with_plugin_require_not_met

  run steampipe query "select 1"
  assert_output --partial 'Warning: could not find plugin which satisfies requirement'
  cd -
}

@test "running steampipe check with mod plugin requirement not met" {
  cd $FILE_PATH/test_data/mods/bad_mod_with_plugin_require_not_met

  run steampipe check all
  assert_output --partial 'Warning: could not find plugin which satisfies requirement'
  cd -
}

@test "running steampipe dashboard with mod plugin requirement not met" {
  cd $FILE_PATH/test_data/mods/bad_mod_with_plugin_require_not_met

  run steampipe dashboard dashboard.sample_dashboard
  assert_output --partial "Warning: could not find plugin which satisfies requirement 'gcp@99.21.0' - required by 'bad_mod_with_require_not_met'"
  cd -
}

@test "running steampipe query with steampipe CLI version requirement not met" {
  cd $FILE_PATH/test_data/mods/bad_mod_with_sp_version_require_not_met

  run steampipe query "select 1"
  assert_output --partial 'does not satisfy mod.bad_mod_with_sp_version_require_not_met which requires version 10.99.99'
  cd -
}

@test "running steampipe check with steampipe CLI version requirement not met" {
  cd $FILE_PATH/test_data/mods/bad_mod_with_sp_version_require_not_met

  run steampipe check all
  assert_output --partial 'does not satisfy mod.bad_mod_with_sp_version_require_not_met which requires version 10.99.99'
  cd -
}

@test "running steampipe dashboard with steampipe CLI version requirement not met" {
  cd $FILE_PATH/test_data/mods/bad_mod_with_sp_version_require_not_met

  run steampipe dashboard dashboard.sample_dashboard
  assert_output --partial 'does not satisfy mod.bad_mod_with_sp_version_require_not_met which requires version 10.99.99'
  cd -
}

@test "running steampipe query with dependant mod version requirement not met(not installed)" {
  cd $FILE_PATH/test_data/mods/bad_mod_with_dep_mod_version_require_not_met

  run steampipe query "select 1"
  assert_output --partial  'Error: failed to load workspace: not all dependencies are installed'

  run steampipe mod install
  assert_output --partial 'Error: 1 dependency failed to install - no version of github.com/turbot/steampipe-mod-aws-compliance found satisfying version constraint: 99.21.0'
  cd -
}

@test "running steampipe check with dependant mod version requirement not met(not installed)" {
  cd $FILE_PATH/test_data/mods/bad_mod_with_dep_mod_version_require_not_met

  run steampipe check all
  assert_output --partial 'Error: failed to load workspace: not all dependencies are installed'

  run steampipe mod install
  assert_output --partial 'Error: 1 dependency failed to install - no version of github.com/turbot/steampipe-mod-aws-compliance found satisfying version constraint: 99.21.0'
  cd -
}

@test "running steampipe dashboard with dependant mod version requirement not met(not installed)" {
  cd $FILE_PATH/test_data/mods/bad_mod_with_dep_mod_version_require_not_met

  run steampipe dashboard dashboard.sample_dashboard
  assert_output --partial  'Error: failed to load workspace: not all dependencies are installed'

  run steampipe mod install
  assert_output --partial 'Error: 1 dependency failed to install - no version of github.com/turbot/steampipe-mod-aws-compliance found satisfying version constraint: 99.21.0'
  cd -
}

### deprecation tests ###

@test "old steampipe property" {
  # go to the mod directory and run steampipe to get the deprectaion warning
  # or error, and check the output
  cd $FILE_PATH/test_data/mods/mod_with_old_steampipe_in_require
  run steampipe query "select 1"

  assert_output --partial "Warning: Property 'steampipe' is deprecated for mod require block - use a steampipe block instead"
}

@test "new steampipe block with old steampipe property" {
  # go to the mod directory and run steampipe to get the deprectaion warning
  # or error, and check the output
  cd $FILE_PATH/test_data/mods/mod_with_old_steampipe_and_new_steampipe_block_in_require
  run steampipe query "select 1"

  assert_output --partial "Both 'steampipe' block and deprecated 'steampipe' property are set"
}

@test "new steampipe block with min_version" {
  # go to the mod directory and run steampipe to get the deprectaion warning
  # or error, and check the output
  cd $FILE_PATH/test_data/mods/mod_with_new_steampipe_block
  run steampipe query "select 1"

  assert_output --partial "1"
}

@test "legacy 'requires' block" {
  # go to the mod directory and run steampipe to get the deprectaion warning
  # or error, and check the output
  cd $FILE_PATH/test_data/mods/mod_with_legacy_requires_block
  run steampipe query "select 1"

  # TODO: update this test when the deprecation warning for legacy 'requries'
  # block is added
  assert_output --partial "1"
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
