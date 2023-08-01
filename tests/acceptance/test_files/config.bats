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
  echo $output
  timing=$(echo $output | jq .timing)
  echo "timing: $timing"
  # timing should be true
  assert_equal $timing true

  # steampipe check with timing set
  run steampipe check all --timing
  echo $output
  timing=$(echo $output | jq .timing)
  echo "timing: $timing"
  # timing should be true
  assert_equal $timing true

  #### test workspace profile options ####

  # steampipe query with no timing set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so timing should be set from the "default" workspace profile
  run steampipe query "select 1"
  echo $output
  timing=$(echo $output | jq .timing)
  echo "timing: $timing"
  # timing should be true(since options.query.timing=true in "default" workspace)
  assert_equal $timing true

  # steampipe query with no timing set, but --workspace is set to "sample",
  # so timing should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  echo $output
  timing=$(echo $output | jq .timing)
  echo "timing: $timing"
  # timing should be false(since options.query.timing=false in "sample" workspace)
  assert_equal $timing false

  # steampipe check with no timing set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so timing should be set from the "default" workspace profile
  run steampipe check all
  echo $output
  timing=$(echo $output | jq .timing)
  echo "timing: $timing"
  # timing should be true(since options.check.timing=true in "default" workspace)
  assert_equal $timing true

  # steampipe check with no timing set, but --workspace is set to "sample",
  # so timing should be set from the "sample" workspace profile
  run steampipe check all --workspace=sample
  echo $output
  timing=$(echo $output | jq .timing)
  echo "timing: $timing"
  # timing should be false(since options.check.timing=false in "sample" workspace)
  assert_equal $timing false
}

@test "query-timeout" {

  #### test command line args ####

  # steampipe query with query-timeout set to 250
  run steampipe query "select 1" --query-timeout=250
  echo $output
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 250
  assert_equal $querytimeout 250

  # steampipe check with query-timeout set to 240
  run steampipe check all --query-timeout=240
  echo $output
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 240
  assert_equal $querytimeout 240

  #### test ENV vars ####

  # steampipe query with STEAMPIPE_QUERY_TIMEOUT set to 250
  export STEAMPIPE_QUERY_TIMEOUT=250
  run steampipe query "select 1"
  echo $output
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 250
  assert_equal $querytimeout 250

  # steampipe check with STEAMPIPE_QUERY_TIMEOUT set to 240
  export STEAMPIPE_QUERY_TIMEOUT=240
  run steampipe check all
  echo $output
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 240
  assert_equal $querytimeout 240
  unset STEAMPIPE_QUERY_TIMEOUT # unset the env var

  #### test workspace profile ####

  # steampipe query with no query-timeout set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so query-timeout should be set from the "default" workspace profile
  run steampipe query "select 1"
  echo $output
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 180(since query-timeout=180 in "default" workspace)
  assert_equal $querytimeout 180

  # steampipe query with no query-timeout set, but --workspace is set to "sample",
  # so query-timeout should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  echo $output
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 200(since query-timeout=200 in "sample" workspace)
  assert_equal $querytimeout 200

  # steampipe check with no query-timeout set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so query-timeout should be set from the "default" workspace profile
  run steampipe check all
  echo $output
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 180(since query-timeout=180 in "default" workspace)
  assert_equal $querytimeout 180

  # steampipe check with no query-timeout set, but --workspace is set to "sample",
  # so query-timeout should be set from the "sample" workspace profile
  run steampipe check all --workspace=sample
  echo $output
  querytimeout=$(echo $output | jq '."query-timeout"')
  echo "querytimeout: $querytimeout"
  # query-timeout should be 200(since query-timeout=200 in "sample" workspace)
  assert_equal $querytimeout 200
}

@test "output" {

  #### test command line args ####

  # steampipe query with output set to json
  run steampipe query "select 1" --output=json
  echo $output
  op=$(echo $output | jq .output)
  echo "output: $op"
  # output should be json
  assert_equal $op '"json"'

  # steampipe check with output set to line
  run steampipe check all --output=line
  echo $output
  op=$(echo $output | jq .output)
  echo "output: $op"
  # output should be line
  assert_equal $op '"line"'

  #### test workspace profile options ####

  # steampipe query with no output set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so output should be set from the "default" workspace profile
  run steampipe query "select 1"
  echo $output
  op=$(echo $output | jq .output)
  echo "output: $op"
  # output should be json(since options.query.output=json in "default" workspace)
  assert_equal $op '"json"'

  # steampipe query with no output set, but --workspace is set to "sample",
  # so output should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  echo $output
  op=$(echo $output | jq .output)
  echo "output: $op"
  # output should be csv(since options.query.output=csv in "sample" workspace)
  assert_equal $op '"csv"'

  # steampipe check with no output set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so output should be set from the "default" workspace profile
  run steampipe check all
  echo $output
  op=$(echo $output | jq .output)
  echo "output: $op"
  # output should be json(since options.check.output=json in "default" workspace)
  assert_equal $op '"json"'

  # steampipe check with no output set, but --workspace is set to "sample",
  # so output should be set from the "sample" workspace profile
  run steampipe check all --workspace=sample
  echo $output
  op=$(echo $output | jq .output)
  echo "output: $op"
  # output should be csv(since options.check.output=csv in "sample" workspace)
  assert_equal $op '"csv"'
}

@test "header" {

  #### test command line args ####

  # steampipe query with header set
  run steampipe query "select 1" --header
  echo $output
  header=$(echo $output | jq .header)
  echo "header: $header"
  # header should be true
  assert_equal $header true

  # steampipe check with header set
  run steampipe check all --header
  echo $output
  header=$(echo $output | jq .header)
  echo "header: $header"
  # header should be true
  assert_equal $header true

  #### test workspace profile options ####

  # steampipe query with no header set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so header should be set from the "default" workspace profile
  run steampipe query "select 1"
  echo $output
  header=$(echo $output | jq .header)
  echo "header: $header"
  # header should be false(since options.query.header=false in "default" workspace)
  assert_equal $header false

  # steampipe query with no header set, but --workspace is set to "sample",
  # so header should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  echo $output
  header=$(echo $output | jq .header)
  echo "header: $header"
  # header should be true(since options.query.header=true in "sample" workspace)
  assert_equal $header true

  # steampipe check with no header set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so header should be set from the "default" workspace profile
  run steampipe check all
  echo $output
  header=$(echo $output | jq .header)
  echo "header: $header"
  # header should be false(since options.check.header=false in "default" workspace)
  assert_equal $header false

  # steampipe check with no header set, but --workspace is set to "sample",
  # so header should be set from the "sample" workspace profile
  run steampipe check all --workspace=sample
  echo $output
  header=$(echo $output | jq .header)
  echo "header: $header"
  # header should be true(since options.check.header=true in "sample" workspace)
  assert_equal $header true
}

@test "multi" {

  #### test workspace profile options ####

  # steampipe query with no multi set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so multi should be set from the "default" workspace profile
  run steampipe query "select 1"
  echo $output
  multi=$(echo $output | jq .multi)
  echo "multi: $multi"
  # multi should be true(since options.query.multi=true in "default" workspace)
  assert_equal $multi true

  # steampipe query with no multi set, but --workspace is set to "sample",
  # so multi should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  echo $output
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
  echo $output
  autocomplete=$(echo $output | jq .autocomplete)
  echo "autocomplete: $autocomplete"
  # autocomplete should be false(since options.query.autocomplete=false in "default" workspace)
  assert_equal $autocomplete false

  # steampipe query with no autocomplete set, but --workspace is set to "sample",
  # so autocomplete should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  echo $output
  autocomplete=$(echo $output | jq .autocomplete)
  echo "autocomplete: $autocomplete"
  # autocomplete should be true(since options.query.autocomplete=true in "sample" workspace)
  assert_equal $autocomplete true
}

@test "separator" {

  #### test command line args ####

  # steampipe query with separator set
  run steampipe query "select 1" --separator="|"
  echo $output
  separator=$(echo $output | jq .separator)
  echo "separator: $separator"
  # separator should be |
  assert_equal $separator '"|"'

  # steampipe check with separator set
  run steampipe check all --separator=","
  echo $output
  separator=$(echo $output | jq .separator)
  echo "separator: $separator"
  # separator should be ,
  assert_equal $separator '","'

  #### test workspace profile options ####

  # steampipe query with no separator set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so separator should be set from the "default" workspace profile
  run steampipe query "select 1"
  echo $output
  separator=$(echo $output | jq .separator)
  echo "separator: $separator"
  # separator should be |(since options.query.separator="|" in "default" workspace)
  assert_equal $separator '"|"'

  # steampipe query with no separator set, but --workspace is set to "sample",
  # so separator should be set from the "sample" workspace profile
  run steampipe query "select 1" --workspace=sample
  echo $output
  separator=$(echo $output | jq .separator)
  echo "separator: $separator"
  # separator should be ,(since options.query.separator="," in "sample" workspace)
  assert_equal $separator '","'

  # steampipe check with no separator set, but STEAMPIPE_WORKSPACE_PROFILES_LOCATION is set,
  # so separator should be set from the "default" workspace profile
  run steampipe check all
  echo $output
  separator=$(echo $output | jq .separator)
  echo "separator: $separator"
  # separator should be |(since options.check.separator="|" in "default" workspace)
  assert_equal $separator '"|"'

  # steampipe check with no separator set, but --workspace is set to "sample",
  # so separator should be set from the "sample" workspace profile
  run steampipe check all --workspace=sample
  echo $output
  separator=$(echo $output | jq .separator)
  echo "separator: $separator"
  # separator should be ,(since options.check.separator="," in "sample" workspace)
  assert_equal $separator '","'
}

@test "database-password" {

  #### test command line args ####

  # steampipe service start with database-password set
  run steampipe service start --database-password=redhood
  echo $output
  databasepassword=$(echo $output | jq '."database-password"')
  echo "databasepassword: $databasepassword"
  # database-password should be redhood
  assert_equal $databasepassword '"redhood"'

  #### test ENV vars ####

  # steampipe query with STEAMPIPE_DATABASE_PASSWORD set
  export STEAMPIPE_DATABASE_PASSWORD=deathstroke
  run steampipe service start
  echo $output
  databasepassword=$(echo $output | jq '."database-password"')
  echo "databasepassword: $databasepassword"
  # database-password should be deathstroke
  assert_equal $databasepassword '"deathstroke"'

  unset STEAMPIPE_DATABASE_PASSWORD # unset the env var
}

@test "show-password" {

  #### test command line args ####

  # steampipe service start with show-password set
  run steampipe service start --show-password
  echo $output
  showpassword=$(echo $output | jq '."show-password"')
  echo "showpassword: $showpassword"
  # show-password should be true
  assert_equal $showpassword true
}

@test "database-port" {

  #### test command line args ####

  # steampipe service start with database-port set
  run steampipe service start --database-port=123
  echo $output
  databaseport=$(echo $output | jq '."database-port"')
  echo "databaseport: $databaseport"
  # database-port should be 123
  assert_equal $databaseport 123

  #### global options(default.spc) ####

  cp $FILE_PATH/test_data/source_files/config_tests/default_config.spc $STEAMPIPE_INSTALL_DIR/config/default.spc

  # steampipe service start with no database-port set, but database.port is set in default.spc(global config),
  # so database-port should be set from there
  run steampipe service start
  echo $output
  databaseport=$(echo $output | jq '."database-port"')
  echo "databaseport: $databaseport"
  # database-port should be 9193
  assert_equal $databaseport 9193

  rm -f $STEAMPIPE_INSTALL_DIR/config/default.spc
}

@test "database-listen" {

  #### test command line args ####

  # steampipe service start with database-listen set
  run steampipe service start --database-listen=network
  echo $output
  databaselisten=$(echo $output | jq '."database-listen"')
  echo "databaselisten: $databaselisten"
  # database-listen should be network
  assert_equal $databaselisten '"network"'

  #### global options(default.spc) ####

  cp $FILE_PATH/test_data/source_files/config_tests/default_config.spc $STEAMPIPE_INSTALL_DIR/config/default.spc

  # steampipe service start with no database-listen set, but database.listen is set in default.spc(global config),
  # so database-listen should be set from there
  run steampipe service start
  echo $output
  databaselisten=$(echo $output | jq '."database-listen"')
  echo "databaselisten: $databaselisten"
  # database-listen should be local
  assert_equal $databaselisten '"local"'

  rm -f $STEAMPIPE_INSTALL_DIR/config/default.spc
}

@test "cache-max-ttl" {

  #### test ENV vars ####

  # steampipe query with STEAMPIPE_CACHE_MAX_TTL set
  export STEAMPIPE_CACHE_MAX_TTL=1000
  run steampipe service start
  echo $output
  cachemaxttl=$(echo $output | jq '."cache-max-ttl"')
  echo "cachemaxttl: $cachemaxttl"
  # cache-max-ttl should be 1000
  assert_equal $cachemaxttl 1000

  unset STEAMPIPE_CACHE_MAX_TTL # unset the env var
}
