load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "old steampipe property" {
  # go to the mod directory and run steampipe to get the deprectaion warning
  # or error, and check the output
  cd $FILE_PATH/test_data/mod_require_tests/mod_with_old_steampipe_in_require
  run steampipe query "select 1"

  assert_output --partial "Warning: Property 'steampipe' is deprecated for mod require block - use a steampipe block instead"
}

@test "new steampipe block with old steampipe property" {
  # go to the mod directory and run steampipe to get the deprectaion warning
  # or error, and check the output
  cd $FILE_PATH/test_data/mod_require_tests/mod_with_old_steampipe_and_new_steampipe_block_in_require
  run steampipe query "select 1"

  assert_output --partial "Both 'steampipe' block and deprecated 'steampipe' property are set"
}

@test "new steampipe block with min_version" {
  # go to the mod directory and run steampipe to get the deprectaion warning
  # or error, and check the output
  cd $FILE_PATH/test_data/mod_require_tests/mod_with_new_steampipe_block
  run steampipe query "select 1"

  assert_output --partial "1"
}

function teardown() {
  cd -
}
