load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "running steampipe query with mod plugin requirement not met" {
  skip
  cd $FILE_PATH/test_data/bad_mod_with_plugin_require_not_met

  run steampipe query
  assert_output --partial 'Error: 1 mod plugin requirement not satisfied.'
  cd -
}

@test "running steampipe check with mod plugin requirement not met" {
  cd $FILE_PATH/test_data/bad_mod_with_plugin_require_not_met

  run steampipe check all
  assert_output --partial 'Error: 1 mod plugin requirement not satisfied.'
  cd -
}

@test "running steampipe dashboard with mod plugin requirement not met" {
  cd $FILE_PATH/test_data/bad_mod_with_plugin_require_not_met

  run steampipe dashboard
  assert_output --partial 'Error: 1 mod plugin requirement not satisfied.'
  cd -
}

# @test "running steampipe query with steampipe CLI version requirement not met" {
#   cd $FILE_PATH/test_data/bad_mod_with_sp_version_require_not_met

#   run steampipe query
#   assert_output --partial 'Error: 1 mod plugin requirement not satisfied.'
#   cd -
# }

@test "running steampipe check with steampipe CLI version requirement not met" {
  cd $FILE_PATH/test_data/bad_mod_with_sp_version_require_not_met

  run steampipe check all
  assert_output --partial 'does not satisfy mod.bad_mod_with_sp_version_require_not_met which requires version 10.99.99'
  cd -
}

@test "running steampipe dashboard with steampipe CLI version requirement not met" {
  cd $FILE_PATH/test_data/bad_mod_with_sp_version_require_not_met

  run steampipe dashboard
  assert_output --partial 'does not satisfy mod.bad_mod_with_sp_version_require_not_met which requires version 10.99.99'
  cd -
}
