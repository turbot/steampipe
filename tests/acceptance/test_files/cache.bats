load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe cache functionality check ON" {
  run steampipe plugin install chaos
  cd $FUNCTIONALITY_TEST_MOD

  run steampipe check benchmark.check_cache_benchmark --export test.json  --max-parallel 1

  # store the unique number from 1st control in `content`
  content=$(cat test.json | jq '.groups[].controls[0].results[0].resource')
  # store the unique number from 2nd control in `new_content`
  new_content=$(cat test.json | jq '.groups[].controls[1].results[0].resource')
  echo $content
  echo $new_content

  # verify that `content` and `new_content` are the same
  assert_equal "$new_content" "$content"
  rm -f test.json
}

@test "steampipe cache functionality check OFF" {
  cd $FUNCTIONALITY_TEST_MOD

  # set the env variable to false
  export STEAMPIPE_CACHE=false
  run steampipe check benchmark.check_cache_benchmark --export test.json  --max-parallel 1

  # store the unique number from 1st control in `content`
  content=$(cat test.json | jq '.groups[].controls[0].results[0].resource')
  # store the unique number from 2nd control in `new_content`
  new_content=$(cat test.json | jq '.groups[].controls[1].results[0].resource')
  echo $content
  echo $new_content

  # verify that `content` and `new_content` are not the same
  if [[ "$content" == "$new_content" ]]; then
    flag=1
  else
    flag=0
  fi
  assert_equal "$flag" "0"
  rm -f test.json
}

@test "steampipe cache functionality check ON(check content of results, not just the unique column)" {
  # start service to turn on caching
  steampipe service start

  steampipe query "select unique_col, a, b from chaos_cache_check" --output json &> output1.json

  steampipe query "select unique_col, a, b from chaos_cache_check" --output json &> output2.json

  # stop service
  steampipe service stop

  # verify that the json contents of output1 and output2 files are the same
  run jd output1.json output2.json
  echo $output
  assert_success

  rm -f output1.json
  rm -f output2.json
}

@test "verify cache ttl works when set in Environment" {
  cp $SRC_DATA_DIR/chaos_no_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_no_options.spc

  # start the service
  steampipe service start
  
  export STEAMPIPE_CACHE_TTL=10

  # cache functionality check since cache=true in options
  steampipe query "select unique_col from chaos_no_options.chaos_cache_check where id=2" --output json > out1.json
  steampipe query "select unique_col from chaos_no_options.chaos_cache_check where id=2" --output json > out2.json
  
  # wait for 15 seconds - the value of the TTL in environment
  sleep 15
  
  # run the query again
  steampipe query "select unique_col from chaos_no_options.chaos_cache_check where id=2" --output json > out3.json

  # stop the service
  steampipe service stop

  unique1=$(cat out1.json | jq '.[].unique_col')
  unique2=$(cat out2.json | jq '.[].unique_col')
  unique3=$(cat out3.json | jq '.[].unique_col')
  # remove the output and the config files
  rm -f out*.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_no_options.spc

  # the first and the seconds query should have the same value
  assert_equal "$unique1" "$unique2"
  # the third query should have a different value
  assert_not_equal "$unique1" "$unique3"
}

@test "verify cache ttl works when set in connections options" {
  skip "skipping - this is a deprecated functionality"
  
  cp $SRC_DATA_DIR/chaos_ttl_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_ttl_options.spc

  # start the service
  steampipe service start

  # cache functionality check since cache=true in options
  steampipe query "select unique_col from chaos_ttl_options.chaos_cache_check where id=2" --output json > out1.json
  steampipe query "select unique_col from chaos_ttl_options.chaos_cache_check where id=2" --output json > out2.json

  # wait for 15 seconds - the value of the TTL in connection options
  sleep 15
  
  # run the query again
  steampipe query "select unique_col from chaos_ttl_options.chaos_cache_check where id=2" --output json > out3.json

  # stop the service
  steampipe service stop

  unique1=$(cat out1.json | jq '.[].unique_col')
  unique2=$(cat out2.json | jq '.[].unique_col')
  unique3=$(cat out3.json | jq '.[].unique_col')

  # remove the output and the config files
  rm -f out*.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_ttl_options.spc

  # the first and the seconds query should have the same value
  assert_equal "$unique1" "$unique2"
  # the third query should have a different value
  assert_not_equal "$unique1" "$unique3"
}

@test "verify cache ttl works when set in database options" {
  export STEAMPIPE_LOG=info

  cp $SRC_DATA_DIR/chaos_no_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_no_options.spc
  cp $SRC_DATA_DIR/default_cache_ttl_10.spc $STEAMPIPE_INSTALL_DIR/config/default.spc
  
  cat $STEAMPIPE_INSTALL_DIR/config/default.spc

  # start the service
  steampipe service start
  cat $STEAMPIPE_INSTALL_DIR/config/default.spc

  # cache functionality check since cache=true in options
  steampipe query "select unique_col from chaos_no_options.chaos_cache_check where id=2" --output json > out1.json
  cat $STEAMPIPE_INSTALL_DIR/config/default.spc
  steampipe query "select unique_col from chaos_no_options.chaos_cache_check where id=2" --output json > out2.json
  cat $STEAMPIPE_INSTALL_DIR/config/default.spc

  # wait for 15 seconds - the value of the TTL in connection options
  sleep 15

  # run the query again
  steampipe query "select unique_col from chaos_no_options.chaos_cache_check where id=2" --output json > out3.json
  cat $STEAMPIPE_INSTALL_DIR/config/default.spc

  # stop the service
  steampipe service stop

  unique1=$(cat out1.json | jq '.[].unique_col')
  unique2=$(cat out2.json | jq '.[].unique_col')
  unique3=$(cat out3.json | jq '.[].unique_col')

  cat $STEAMPIPE_INSTALL_DIR/config/default.spc
  cat $STEAMPIPE_INSTALL_DIR/config/chaos_no_options.spc

  # remove the output and the config files
  rm -f out*.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_no_options.spc
  rm -f $STEAMPIPE_INSTALL_DIR/config/default.spc

  # the first and the seconds query should have the same value
  assert_equal "$unique1" "$unique2"
  # the third query should have a different value
  assert_not_equal "$unique1" "$unique3"
}

@test "test caching with cache=true in workspace profile" {
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
    cp $SRC_DATA_DIR/workspace_cache_enabled.spc $STEAMPIPE_INSTALL_DIR/config/workspace_cache_enabled.spc

    # cache functionality check since cache=true in workspace profile
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export test.json --max-parallel 1

    # store the unique number from 1st control in `content`
    content=$(cat test.json | jq '.groups[].controls[0].results[0].resource')
    # store the unique number from 2nd control in `new_content`
    new_content=$(cat test.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content
    # remove the output and the config files
    rm -f test.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
    rm -f $STEAMPIPE_INSTALL_DIR/config/workspace_cache_enabled.spc

    # verify that `content` and `new_content` are the same
    assert_equal "$new_content" "$content"
}

@test "test caching with cache=false in workspace profile" {
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
    cp $SRC_DATA_DIR/workspace_cache_disabled.spc $STEAMPIPE_INSTALL_DIR/config/workspace_cache_disabled.spc

    # cache functionality check since cache=false in workspace profile
    cd $CONFIG_PARSING_TEST_MOD
    run steampipe check benchmark.config_parsing_benchmark --export test.json --max-parallel 1

    # store the unique number from 1st control in `content`
    content=$(cat test.json | jq '.groups[].controls[0].results[0].resource')
    # store the unique number from 2nd control in `new_content`
    new_content=$(cat test.json | jq '.groups[].controls[1].results[0].resource')
    echo $content
    echo $new_content

    # verify that `content` and `new_content` are not the same
    if [[ "$content" == "$new_content" ]]; then
        flag=1
    else
        flag=0
    fi
    # remove the output and the config files
    rm -f test.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
    rm -f $STEAMPIPE_INSTALL_DIR/config/workspace_cache_disabled.spc

    assert_equal "$flag" "0"
}

@test "verify cache ttl works when set in workspace profile" {
  cp $FILE_PATH/test_data/source_files/workspace_cache_ttl.spc $STEAMPIPE_INSTALL_DIR/config/workspace.spc
  cp $SRC_DATA_DIR/chaos_no_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_no_options.spc

  # start the service
  steampipe service start

  # cache functionality check since cache=true in options
  steampipe query "select unique_col from chaos_no_options.chaos_cache_check where id=2" --output json > out1.json
  steampipe query "select unique_col from chaos_no_options.chaos_cache_check where id=2" --output json > out2.json

  # wait for 15 seconds - the value of the TTL in connection options
  sleep 15

  # run the query again
  steampipe query "select unique_col from chaos_no_options.chaos_cache_check where id=2" --output json > out3.json

  # stop the service
  steampipe service stop

  unique1=$(cat out1.json | jq '.[].unique_col')
  unique2=$(cat out2.json | jq '.[].unique_col')
  unique3=$(cat out3.json | jq '.[].unique_col')

  # remove the output and the config files
  rm -f out*.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_no_options.spc
  rm -f $STEAMPIPE_INSTALL_DIR/config/workspace.spc

  # the first and the seconds query should have the same value
  assert_equal "$unique1" "$unique2"
  # the third query should have a different value
  assert_not_equal "$unique1" "$unique3"
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
