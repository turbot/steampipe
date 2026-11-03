load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"
load ../helpers/steampipe_processes

@test "steampipe cache functionality check ON" {
  run steampipe plugin install chaos

  # start service to turn on caching
  steampipe service start

  # run two queries to check if the results are the same
  run steampipe query "select unique_col from chaos_cache_check limit 1" --output json > output1.json
  run steampipe query "select unique_col from chaos_cache_check limit 1" --output json > output2.json

  # stop service
  steampipe service stop

  unique1=$(cat output1.json | jq '.rows[0].unique_col')
  unique2=$(cat output2.json | jq '.rows[0].unique_col')

  echo $unique1
  echo $unique2

  assert_equal "$unique1" "$unique2"

  rm -f output1.json
  rm -f output2.json
}

@test "steampipe cache functionality check ON(check content of results, not just the unique column)" {
  # start service to turn on caching
  steampipe service start

  steampipe query "select unique_col, a, b from chaos_cache_check" --output json &> output1.json

  steampipe query "select unique_col, a, b from chaos_cache_check" --output json &> output2.json

  # stop service
  steampipe service stop

  # verify that the json contents of output1 and output2 files are the same
  run jd -f patch output1.json output2.json
  echo $output

  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  # check if there is no diff returned by the script
  assert_equal "$diff" ""

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

  unique1=$(cat out1.json | jq '.rows[0].unique_col')
  unique2=$(cat out2.json | jq '.rows[0].unique_col')
  unique3=$(cat out3.json | jq '.rows[0].unique_col')
  # remove the output and the config files
  rm -f out*.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_no_options.spc

  # the first and the seconds query should have the same value
  assert_equal "$unique1" "$unique2"
  # the third query should have a different value
  assert_not_equal "$unique1" "$unique3"
}

@test "verify cache ttl works when set in database options" {
  skip "TODO - fix and test using steampipe query command"
  export STEAMPIPE_LOG=info
  cp $SRC_DATA_DIR/chaos_no_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_no_options.spc

  # start the service
  steampipe service start

  cp $SRC_DATA_DIR/default_cache_ttl_10.spc $STEAMPIPE_INSTALL_DIR/config/default.spc
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

  unique1=$(cat out1.json | jq '.rows[0].unique_col')
  unique2=$(cat out2.json | jq '.rows[0].unique_col')
  unique3=$(cat out3.json | jq '.rows[0].unique_col')

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

@test "test caching with cache=false in workspace profile" {
  cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
  cp $SRC_DATA_DIR/workspace_cache_disabled.spc $STEAMPIPE_INSTALL_DIR/config/workspace_cache_disabled.spc

  run steampipe service start
  started=$status

  # each of the table's 10 rows gets a new random unique_col on every uncached fetch
  steampipe query "select unique_col from chaos6.chaos_cache_check order by id" --output json > out1.json || true
  steampipe query "select unique_col from chaos6.chaos_cache_check order by id" --output json > out2.json || true

  run steampipe service stop

  unique1=$(jq -c '[.rows[].unique_col]' out1.json || true)
  unique2=$(jq -c '[.rows[].unique_col]' out2.json || true)

  # remove the output and the config files before asserting, so a failure leaves nothing behind
  rm -f out*.json
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
  rm -f $STEAMPIPE_INSTALL_DIR/config/workspace_cache_disabled.spc

  assert_equal "$started" "0"
  assert_equal "$(echo $unique1 | jq 'length')" "10"
  assert_equal "$(echo $unique2 | jq 'length')" "10"
  assert_not_equal "$unique1" "$unique2"
}

@test "verify cache ttl works when set in workspace profile" {
  skip "TODO - test using steampipe query command"
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

  unique1=$(cat out1.json | jq '.rows[0].unique_col')
  unique2=$(cat out2.json | jq '.rows[0].unique_col')
  unique3=$(cat out3.json | jq '.rows[0].unique_col')

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
  num=$(count_steampipe_processes)
  assert_equal $num 0
}
