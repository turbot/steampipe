load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

function setup_file() {
  cd $FILE_PATH/test_data/source_files/config_tests
  export STEAMPIPE_WORKSPACE_PROFILES_LOCATION=$FILE_PATH/test_data/source_files/config_tests/workspace_profiles_config
  export STEAMPIPE_DIAGNOSTICS=config_json
}

@test "timing" {

  #### test command line args ####

  # steampipe query with timing set
  run steampipe query "select 1" --timing
  timing=$(echo $output | jq .timing)
  echo "timing: $timing"
  # timing should be true
  assert_equal $timing true

  # steampipe check with timing set
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all --timing
  timing=$(echo $output | jq .timing)
  cd - # go back to the previous directory
  echo "timing: $timing"
  # timing should be true
  assert_equal $timing true

  #### test workspace profile options ####

  # steampipe query with no timing set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so timing should be set from the "default" workspace profile
  run steampipe query "select 1"
  timing=$(echo $output | jq .timing)
  echo "timing: $timing"
  # timing should be true(since options.query.timing=true in "default" workspace)
  assert_equal $timing true

  # steampipe query with no timing set, but --workspace is set to "sample",
  # so timing should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  timing=$(echo $output | jq .timing)
  echo "timing: $timing"
  # timing should be false(since options.query.timing=false in "sample" workspace)
  assert_equal $timing false

  # steampipe check with no timing set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so timing should be set from the "default" workspace profile
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all
  timing=$(echo $output | jq .timing)
  cd - # go back to the previous directory
  echo "timing: $timing"
  # timing should be true(since options.check.timing=true in "default" workspace)
  assert_equal $timing true

  # steampipe check with no timing set, but --workspace is set to "sample",
  # so timing should be set from the "sample" workspace profile
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all --workspace=sample
  timing=$(echo $output | jq .timing)
  cd - # go back to the previous directory
  echo "timing: $timing"
  # timing should be true(since options.check.timing=true in "default" workspace)
  assert_equal $timing false
}
