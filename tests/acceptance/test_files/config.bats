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
  # timing should be false(since options.check.timing=false in "sample" workspace)
  assert_equal $timing false
}

@test "query-timeout" {

  #### test command line args ####

  # steampipe query with query-timeout set to 250
  run steampipe query "select 1" --query-timeout=250
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 250
  assert_equal $querytimeout 250

  # steampipe check with query-timeout set to 240
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all --query-timeout=240
  querytimeout=$(echo $output | jq '."query-timeout"')
  cd - # go back to the previous directory
  echo "querytimeout: $querytimeout"
  # query-timeout should be 240
  assert_equal $querytimeout 240

  #### test ENV vars ####

  # steampipe query with STEAMPIPE_QUERY_TIMEOUT set to 250
  export STEAMPIPE_QUERY_TIMEOUT=250
  run steampipe query "select 1"
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 250
  assert_equal $querytimeout 250

  # steampipe check with STEAMPIPE_QUERY_TIMEOUT set to 240
  export STEAMPIPE_QUERY_TIMEOUT=240
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all
  querytimeout=$(echo $output | jq '."query-timeout"')
  cd - # go back to the previous directory
  echo "querytimeout: $querytimeout"
  # query-timeout should be 240
  assert_equal $querytimeout 240
  unset STEAMPIPE_QUERY_TIMEOUT # unset the env var

  #### test workspace profile ####

  # steampipe query with no query-timeout set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so query-timeout should be set from the "default" workspace profile
  run steampipe query "select 1"
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 180(since query-timeout=180 in "default" workspace)
  assert_equal $querytimeout 180

  # steampipe query with no query-timeout set, but --workspace is set to "sample",
  # so query-timeout should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 200(since query-timeout=200 in "sample" workspace)
  assert_equal $querytimeout 200

  # steampipe check with no query-timeout set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so query-timeout should be set from the "default" workspace profile
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all
  querytimeout=$(echo $output | jq '."query-timeout"')
  cd - # go back to the previous directory
  echo "querytimeout: $querytimeout"
  # query-timeout should be 180(since query-timeout=180 in "default" workspace)
  assert_equal $querytimeout 180

  # steampipe check with no query-timeout set, but --workspace is set to "sample",
  # so query-timeout should be set from the "sample" workspace profile
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all --workspace=sample
  querytimeout=$(echo $output | jq '."query-timeout"')
  cd - # go back to the previous directory
  echo "querytimeout: $querytimeout"
  # query-timeout should be 200(since query-timeout=200 in "sample" workspace)
  assert_equal $querytimeout 200
}

@test "output" {

  #### test command line args ####

  # steampipe query with output set to json
  run steampipe query "select 1" --output=json
  op=$(echo $output | jq .output)
  echo "output: $op"
  # output should be json
  assert_equal $op '"json"'

  # steampipe check with output set to line
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all --output=line
  op=$(echo $output | jq .output)
  cd - # go back to the previous directory
  echo "output: $op"
  # output should be line
  assert_equal $op '"line"'

  #### test workspace profile options ####

  # steampipe query with no output set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so output should be set from the "default" workspace profile
  run steampipe query "select 1"
  op=$(echo $output | jq .output)
  echo "output: $op"
  # output should be json(since options.query.output=json in "default" workspace)
  assert_equal $op '"json"'

  # steampipe query with no output set, but --workspace is set to "sample",
  # so output should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  op=$(echo $output | jq .output)
  echo "output: $op"
  # output should be csv(since options.query.output=csv in "sample" workspace)
  assert_equal $op '"csv"'

  # steampipe check with no output set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so output should be set from the "default" workspace profile
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all
  op=$(echo $output | jq .output)
  cd - # go back to the previous directory
  echo "output: $op"
  # output should be json(since options.check.output=json in "default" workspace)
  assert_equal $op '"json"'

  # steampipe check with no output set, but --workspace is set to "sample",
  # so output should be set from the "sample" workspace profile
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all --workspace=sample
  op=$(echo $output | jq .output)
  cd - # go back to the previous directory
  echo "output: $op"
  # output should be csv(since options.check.output=csv in "sample" workspace)
  assert_equal $op '"csv"'
}

@test "header" {

  #### test command line args ####

  # steampipe query with header set
  run steampipe query "select 1" --header
  header=$(echo $output | jq .header)
  echo "header: $header"
  # header should be true
  assert_equal $header true

  # steampipe check with header set
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all --header
  header=$(echo $output | jq .header)
  cd - # go back to the previous directory
  echo "header: $header"
  # header should be true
  assert_equal $header true

  #### test workspace profile options ####

  # steampipe query with no header set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so header should be set from the "default" workspace profile
  run steampipe query "select 1"
  header=$(echo $output | jq .header)
  echo "header: $header"
  # header should be false(since options.query.header=false in "default" workspace)
  assert_equal $header false

  # steampipe query with no header set, but --workspace is set to "sample",
  # so header should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  header=$(echo $output | jq .header)
  echo "header: $header"
  # header should be true(since options.query.header=true in "sample" workspace)
  assert_equal $header true

  # steampipe check with no header set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so header should be set from the "default" workspace profile
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all
  header=$(echo $output | jq .header)
  cd - # go back to the previous directory
  echo "header: $header"
  # header should be false(since options.check.header=false in "default" workspace)
  assert_equal $header false

  # steampipe check with no header set, but --workspace is set to "sample",
  # so header should be set from the "sample" workspace profile
  cd $FILE_PATH/test_data/mods/functionality_test_mod # cd to a mod directory to run check
  run steampipe check all --workspace=sample
  header=$(echo $output | jq .header)
  cd - # go back to the previous directory
  echo "header: $header"
  # header should be true(since options.check.header=true in "sample" workspace)
  assert_equal $header true
}

@test "multi" {

  #### test workspace profile options ####

  # steampipe query with no multi set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so multi should be set from the "default" workspace profile
  run steampipe query "select 1"
  multi=$(echo $output | jq .multi)
  echo "multi: $multi"
  # multi should be true(since options.query.multi=true in "default" workspace)
  assert_equal $multi true

  # steampipe query with no multi set, but --workspace is set to "sample",
  # so multi should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  multi=$(echo $output | jq .multi)
  echo "multi: $multi"
  # multi should be false(since options.query.multi=false in "sample" workspace)
  assert_equal $multi false
}

@test "autocomplete" {

  #### test workspace profile options ####

  # steampipe query with no autocomplete set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so autocomplete should be set from the "default" workspace profile
  run steampipe query "select 1"
  autocomplete=$(echo $output | jq .autocomplete)
  echo "autocomplete: $autocomplete"
  # autocomplete should be false(since options.query.autocomplete=false in "default" workspace)
  assert_equal $autocomplete false

  # steampipe query with no autocomplete set, but --workspace is set to "sample",
  # so autocomplete should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  autocomplete=$(echo $output | jq .autocomplete)
  echo "autocomplete: $autocomplete"
  # autocomplete should be true(since options.query.autocomplete=true in "sample" workspace)
  assert_equal $autocomplete true
}
