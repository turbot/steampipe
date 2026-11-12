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

# The query cache size limit is one total for all plugins. Each plugin gets a share of the total,
# in proportion to its number of connections. With the chaos2 connection added, chaos has 2
# connections (chaos, chaos2) and chaosdynamic has 1. So for a total of 300 MB, chaos gets
# 300 x 2/3 = 200 MB and chaosdynamic gets 300 x 1/3 = 100 MB.

# count the lines in the plugin logs
function plugin_log_line_count() {
  cat $STEAMPIPE_INSTALL_DIR/logs/plugin-*.log 2>/dev/null | wc -l | tr -d ' '
}

# print the query cache size(s) that plugin $1 logged when it started, after line $2 of the plugin logs
function cache_size_logged_by() {
  cat $STEAMPIPE_INSTALL_DIR/logs/plugin-*.log 2>/dev/null | tail -n +$(( $2 + 1 )) | grep -F "steampipe-plugin-$1.plugin:" | grep -o 'query cache created, max size [0-9]*Mb' | grep -o '[0-9][0-9]*' | sort -u
}

# start both plugins with one query
function query_both_plugins() {
  steampipe query "select (select count(*) from chaos2.chaos_all_column_types) as chaos_rows, (select count(*) from chaosdynamic.test1) as chaosdynamic_rows"
}

@test "cache_max_size_mb in database options is one total, shared by the plugins by connection count" {
  export STEAMPIPE_LOG=info
  cp $SRC_DATA_DIR/cache_size_chaos2.spc $STEAMPIPE_INSTALL_DIR/config/cache_size_chaos2.spc
  cp $SRC_DATA_DIR/cache_max_size_300.spc $STEAMPIPE_INSTALL_DIR/config/cache_max_size_300.spc
  ls $STEAMPIPE_INSTALL_DIR/config

  start=$(plugin_log_line_count)
  run query_both_plugins
  echo $output
  queried=$status

  chaos=$(cache_size_logged_by chaos $start)
  chaosdynamic=$(cache_size_logged_by chaosdynamic $start)
  echo "chaos: $chaos, chaosdynamic: $chaosdynamic"

  rm -f $STEAMPIPE_INSTALL_DIR/config/cache_size_chaos2.spc
  rm -f $STEAMPIPE_INSTALL_DIR/config/cache_max_size_300.spc

  assert_equal "$queried" "0"
  assert_equal "$chaos" "200"
  assert_equal "$chaosdynamic" "100"
}

@test "STEAMPIPE_CACHE_MAX_SIZE_MB is one total, shared by the plugins by connection count" {
  export STEAMPIPE_LOG=info
  cp $SRC_DATA_DIR/cache_size_chaos2.spc $STEAMPIPE_INSTALL_DIR/config/cache_size_chaos2.spc
  ls $STEAMPIPE_INSTALL_DIR/config

  start=$(plugin_log_line_count)
  export STEAMPIPE_CACHE_MAX_SIZE_MB=300
  run query_both_plugins
  echo $output
  queried=$status

  chaos=$(cache_size_logged_by chaos $start)
  chaosdynamic=$(cache_size_logged_by chaosdynamic $start)
  echo "chaos: $chaos, chaosdynamic: $chaosdynamic"

  rm -f $STEAMPIPE_INSTALL_DIR/config/cache_size_chaos2.spc

  assert_equal "$queried" "0"
  assert_equal "$chaos" "200"
  assert_equal "$chaosdynamic" "100"
}

@test "STEAMPIPE_CACHE_MAX_SIZE_MB takes precedence over cache_max_size_mb in database options" {
  export STEAMPIPE_LOG=info
  cp $SRC_DATA_DIR/cache_size_chaos2.spc $STEAMPIPE_INSTALL_DIR/config/cache_size_chaos2.spc
  cp $SRC_DATA_DIR/cache_max_size_300.spc $STEAMPIPE_INSTALL_DIR/config/cache_max_size_300.spc
  ls $STEAMPIPE_INSTALL_DIR/config

  start=$(plugin_log_line_count)
  # a total of 600 MB: chaos gets 600 x 2/3 = 400 MB, chaosdynamic gets 600 x 1/3 = 200 MB
  export STEAMPIPE_CACHE_MAX_SIZE_MB=600
  run query_both_plugins
  echo $output
  queried=$status

  chaos=$(cache_size_logged_by chaos $start)
  chaosdynamic=$(cache_size_logged_by chaosdynamic $start)
  echo "chaos: $chaos, chaosdynamic: $chaosdynamic"

  rm -f $STEAMPIPE_INSTALL_DIR/config/cache_size_chaos2.spc
  rm -f $STEAMPIPE_INSTALL_DIR/config/cache_max_size_300.spc

  assert_equal "$queried" "0"
  assert_equal "$chaos" "400"
  assert_equal "$chaosdynamic" "200"
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(count_steampipe_processes)
  assert_equal $num 0
}
