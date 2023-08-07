load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe plugin help is displayed when no sub command given" {
  steampipe plugin > test.txt

  # checking for OS type, since sed command is different for linux and OSX
  # removing lines, since they contain absolute file paths
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".txt" "36d" test.txt
  else
    run sed -i "36d" test.txt
  fi

  assert_equal "$(cat test.txt)" "$(cat $TEST_DATA_DIR/expected_plugin_help_output.txt)"
  rm -f test.txt*
}

@test "plugin install" {
  run steampipe plugin install net
  assert_success
  steampipe plugin uninstall net
}

@test "plugin install from stream" {
  run steampipe plugin install net@0.2
  assert_success
  steampipe plugin uninstall net@0.2
}

@test "plugin install from stream (prefixed with v)" {
  run steampipe plugin install net@v0.2
  assert_success
  steampipe plugin uninstall net@0.2
}

@test "start service, install plugin and query" {
  # start service
  steampipe service start

  # install plugin
  steampipe plugin install chaos

  # query the plugin
  run steampipe query "select time_col from chaos_cache_check limit 1"
  # check if the query passes
  assert_success

  # stop service
  steampipe service stop

  # check service status
  run steampipe service status

  assert_output "$output" "Service is not running"
}

@test "steampipe plugin list" {
    run steampipe plugin list
    assert_success
}

@test "steampipe plugin list works with disabled connections" {
  rm -f $STEAMPIPE_INSTALL_DIR/config/*
  cp $SRC_DATA_DIR/chaos_conn_import_disabled.spc $STEAMPIPE_INSTALL_DIR/config/chaos_conn_import_disabled.spc
  run steampipe plugin list 2>&3 1>&3
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_conn_import_disabled.spc
  assert_success
}

## connection config

@test "steampipe aggregator connection wildcard check" {
    skip
    run steampipe plugin install chaos
    run steampipe plugin install steampipe
    cp $SRC_DATA_DIR/aggregator.spc $STEAMPIPE_INSTALL_DIR/config/chaos_agg.spc
    run steampipe query "select * from chaos_group.chaos_all_column_types"
    assert_success
}

@test "steampipe aggregator connection check total results" {
    skip
    run steampipe query "select * from chaos.chaos_all_numeric_column" --output json

    # store the length of the result when queried using `chaos` connection
    length_chaos=$(echo $output | jq length)

    run steampipe query "select * from chaos2.chaos_all_numeric_column" --output json

    # store the length of the result when queried using `chaos2` connection
    length_chaos_2=$(echo $output | jq length)

    run steampipe query "select * from chaos_group.chaos_all_numeric_column" --output json

    # store the length of the result when queried using `chaos_group` aggregated connection
    length_chaos_agg=$(echo $output | jq length)

    # since the aggregator connection `chaos_group` contains two chaos connections, we expect
    # the number of results returned will be the summation of the two
    assert_equal "$length_chaos_agg" "$((length_chaos+length_chaos_2))"
}

@test "steampipe aggregator connections should fail when querying a different plugin" {
    skip
    run steampipe query "select * from chaos_group.chaos_all_numeric_column order by id"

    # this should pass since the aggregator contains only chaos connections
    assert_success
    
    run steampipe query "select * from chaos_group.steampipe_registry_plugin order by id"

    # this should fail since the aggregator contains only chaos connections, and we are
    # querying a steampipe table
    assert_failure
}

@test "steampipe json connection config" {
    cp $SRC_DATA_DIR/chaos2.json $STEAMPIPE_INSTALL_DIR/config/chaos2.json

    run steampipe query "select time_col from chaos4.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos2.json

    assert_success
}

@test "steampipe should return an error for duplicate connection name" {
    cp $SRC_DATA_DIR/chaos.json $STEAMPIPE_INSTALL_DIR/config/chaos2.json
    cp $SRC_DATA_DIR/chaos.json $STEAMPIPE_INSTALL_DIR/config/chaos3.json
    
    # this should fail because of duplicate connection name
    run steampipe query "select time_col from chaos.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos2.json
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos3.json

    assert_output --partial 'duplicate connection name'
}

@test "steampipe yaml connection config" {
    cp $SRC_DATA_DIR/chaos2.yml $STEAMPIPE_INSTALL_DIR/config/chaos3.yml

    run steampipe query "select time_col from chaos5.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos3.yml

    assert_success
}

@test "steampipe test connection config with options(hcl)" {
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    run steampipe query "select time_col from chaos6.chaos_cache_check"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    assert_success
}

@test "steampipe test connection config with options(yml)" {
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    run steampipe query "select time_col from chaos6.chaos_cache_check"
    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    assert_success
}

@test "steampipe test connection config with options(json)" {
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    run steampipe query "select time_col from chaos6.chaos_cache_check"
    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    assert_success
}

@test "steampipe check regions in connection config is being parsed and used(hcl)" {
    cp $SRC_DATA_DIR/chaos_options.spc $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc

    # check regions in connection config is being parsed and used
    run steampipe query "select * from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.spc
    # check output
    assert_equal "$result" '[{"_ctx":{"connection_name":"chaos6"},"id":0,"region_name":"us-east-1"},{"_ctx":{"connection_name":"chaos6"},"id":3,"region_name":"us-west-2"}]'

}

@test "steampipe check regions in connection config is being parsed and used(yml)" {
    cp $SRC_DATA_DIR/chaos_options.yml $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml

    # check regions in connection config is being parsed and used
    run steampipe query "select * from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.yml
    # check output
    assert_equal "$result" '[{"_ctx":{"connection_name":"chaos6"},"id":0,"region_name":"us-east-1"},{"_ctx":{"connection_name":"chaos6"},"id":3,"region_name":"us-west-2"}]'

}

@test "steampipe check regions in connection config is being parsed and used(json)" {
    cp $SRC_DATA_DIR/chaos_options.json $STEAMPIPE_INSTALL_DIR/config/chaos_options.json

    # check regions in connection config is being parsed and used
    run steampipe query "select * from chaos6.chaos_regions order by id" --output json
    result=$(echo $output | tr -d '[:space:]')

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_options.json
    # check output
    assert_equal "$result" '[{"_ctx":{"connection_name":"chaos6"},"id":0,"region_name":"us-east-1"},{"_ctx":{"connection_name":"chaos6"},"id":3,"region_name":"us-west-2"}]'

}

@test "connection name escaping" {
    cp $SRC_DATA_DIR/chaos_conn_name_escaping.spc $STEAMPIPE_INSTALL_DIR/config/chaos_conn_name_escaping.spc

    # keywords should be escaped properly
    run steampipe query "select * from \"escape\".chaos_limit limit 1"

    # remove the config file
    rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_conn_name_escaping.spc

    assert_success
}

@test "plugin list - output table" {
  export STEAMPIPE_DISPLAY_WIDTH=100
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  run steampipe plugin list --install-dir $tmpdir
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_table.txt)"
}

@test "plugin list - output json" {
  export STEAMPIPE_DISPLAY_WIDTH=100
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  run steampipe plugin list --install-dir $tmpdir --output json
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_json.json)"
}

@test "plugin list - output table (with a missing plugin)" {
  export STEAMPIPE_DISPLAY_WIDTH=100
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  # uninstall a plugin but dont remove the config - to simulate the missing plugin scenario
  steampipe plugin uninstall hackernews@0.6.0 --install-dir $tmpdir
  run steampipe plugin list --install-dir $tmpdir
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_table_with_missing_plugins.txt)"
}

@test "plugin list - output json (with a missing plugin)" {
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  # uninstall a plugin but dont remove the config - to simulate the missing plugin scenario
  steampipe plugin uninstall hackernews@0.6.0 --install-dir $tmpdir
  run steampipe plugin list --install-dir $tmpdir --output json
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_json_with_missing_plugins.json)"
}

# TODO: finds other ways to simulate failed plugins

@test "plugin list - output table (with a failed plugin)" {
  export STEAMPIPE_DISPLAY_WIDTH=100
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  # remove the contents of a plugin execuatable to simulate the failed plugin scenario
  cat /dev/null > $tmpdir/plugins/hub.steampipe.io/plugins/turbot/hackernews@0.6.0/steampipe-plugin-hackernews.plugin
  run steampipe plugin list --install-dir $tmpdir
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_table_with_failed_plugins.txt)"
}

@test "plugin list - output json (with a failed plugin)" {
  tmpdir="$(mktemp -d)"
  steampipe plugin install hackernews@0.6.0 bitbucket@0.3.1 --progress=false --install-dir $tmpdir
  # remove the contents of a plugin binary execuatable to simulate the failed plugin scenario
  cat /dev/null > $tmpdir/plugins/hub.steampipe.io/plugins/turbot/hackernews@0.6.0/steampipe-plugin-hackernews.plugin
  run steampipe plugin list --install-dir $tmpdir --output json
  echo $output
  rm -rf $tmpdir
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_json_with_failed_plugins.json)"
}

@test "verify that installing plugins creates individual version.json files" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile1="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  vFile2="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/version.json"
  
  [ ! -f $vFile1 ] && fail "could not find $vFile1"
  [ ! -f $vFile2 ] && fail "could not find $vFile2"
  
  rm -rf $tmpdir
}

@test "verify that backfilling of individual plugin version.json works" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile1="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  vFile2="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/version.json"
  
  file1Content=$(cat $vFile1)
  file2Content=$(cat $vFile2)
  
  # remove the individual version files
  rm -f $vFile1
  rm -f $vFile2
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $tmpdir
  
  [ ! -f $vFile1 ] && fail "could not find $vFile1"
  [ ! -f $vFile2 ] && fail "could not find $vFile2"
  
  assert_equal "$(cat $vFile1)" "$file1Content"
  assert_equal "$(cat $vFile2)" "$file2Content"
  
  rm -rf $tmpdir
}

@test "verify that backfilling of individual plugin version.json works where it is only partially backfilled" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile1="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  vFile2="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/version.json"
  
  file1Content=$(cat $vFile1)
  file2Content=$(cat $vFile2)
  
  # remove one individual version file
  rm -f $vFile1
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $tmpdir
  
  [ ! -f $vFile1 ] && fail "could not find $vFile1"
  [ ! -f $vFile2 ] && fail "could not find $vFile2"
  
  assert_equal "$(cat $vFile1)" "$file1Content"
  assert_equal "$(cat $vFile2)" "$file2Content"
  
  rm -rf $tmpdir
}

@test "verify that global plugin/versions.json is composed from individual version.json files when it is absent" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile="$tmpdir/plugins/versions.json"
  
  fileContent=$(cat $vFile)
  
  # remove global version file
  rm -f $vFile
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $tmpdir
  
  ls -la $vFile
  
  [ ! -f $vFile ] && fail "could not find $vFile"
  
  assert_equal "$(cat $vFile)" "$fileContent"
  
  rm -rf $tmpdir
}

@test "verify that global plugin/versions.json is composed from individual version.json files when it is corrupt" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile="$tmpdir/plugins/versions.json"
  fileContent=$(cat $vFile)
  
  # remove global version file
  echo "badline to corrupt versions.json" >> $vFile
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $tmpdir
  
  [ ! -f $vFile ] && fail "could not find $vFile"
  
  assert_equal "$(cat $vFile)" "$fileContent"
  
  rm -rf $tmpdir
}

@test "verify that composition of global plugin/versions.json works when an individual version.json file is corrupt" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success
  
  vFile="$tmpdir/plugins/versions.json"  
  vFile1="$tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  
  # corrupt a version file
  echo "bad line to corrupt" >> $vFile1
  
  # remove global file
  rm -f $vFile
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $tmpdir

  # verify that global file got created
  [ ! -f $vFile ] && fail "could not find $vFile"
  
  rm -rf $tmpdir
}

@test "verify that plugin installed from registry are marked as 'local' when the modtime of the binary is after the install time" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net chaos --install-dir $tmpdir
  assert_success

  # wait for a couple of seconds
  sleep 2

  # touch one of the plugin binaries
  touch $tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/steampipe-plugin-net.plugin

  # run steampipe again so that the plugin version files get backfilled
  version=$(steampipe plugin list --install-dir $tmpdir --output json | jq '.installed' | jq '. | map(select(.name | contains("net@latest")))' | jq '.[0].version')

  # assert
  assert_equal "$version" '"local"'

  rm -rf $tmpdir
}

@test "verify that steampipe check should bypass plugin requirement detection if installed plugin is local" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install net --install-dir $tmpdir
  assert_success

  # wait for a couple of seconds
  sleep 2

  # touch one of the plugin binaries
  touch $tmpdir/plugins/hub.steampipe.io/plugins/turbot/net@latest/steampipe-plugin-net.plugin

  run steampipe plugin list --install-dir $tmpdir
  echo $output

  # clone a mod which has a net plugin requirement
  cd $tmpdir
  git clone https://github.com/turbot/steampipe-mod-net-insights.git
  cd steampipe-mod-net-insights

  # run steampipe check
  run steampipe check all --install-dir $tmpdir

  # check - the plugin requirement warning should not be present in the output
  substring="Warning: could not find plugin which satisfies requirement"
  if [[ ! $output == *"$substring"* ]]; then
    run echo "Warning is not present in the output"
  else
    run echo "Warning is present in the output"
  fi

  assert_equal "$output" "Warning is not present in the output"
  rm -rf $tmpdir
}

@test "verify that plugin installed with --skip-config as true, should not have create a default config .spc file in config folder" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install aws --skip-config --install-dir $tmpdir
  assert_success

  run test -f $tmpdir/config/aws.spc
  assert_failure

  rm -rf $tmpdir
}

@test "verify that plugin installed with --skip-config as false(default), should have default config .spc file in config folder" {
  tmpdir=$(mktemp -d)
  run steampipe plugin install aws --install-dir $tmpdir
  assert_success

  run test -f $tmpdir/config/aws.spc
  assert_success

  rm -rf $tmpdir
}

@test "verify reinstalling a plugin does not overwrite existing plugin config" {
  # check if the default/tweaked config file for a plugin is not deleted after
  # re-installation of a plugin

  tmpdir=$(mktemp -d)

  run steampipe plugin install aws --install-dir $tmpdir

  run test -f $tmpdir/config/aws.spc
  assert_success

  echo '
  connection "aws" {
    plugin = "aws"
    endpoint_url = "http://localhost:4566"
  }
  ' >> $tmpdir/config/aws.spc
  cp $tmpdir/config/aws.spc config.spc

  run steampipe plugin uninstall aws --install-dir $tmpdir

  run steampipe plugin install aws --skip-config --install-dir $tmpdir

  run test -f $tmpdir/config/aws.spc
  assert_success

  run diff $tmpdir/config/aws.spc config.spc
  assert_success

  rm config.spc
  rm -rf $tmpdir
}

@test "cleanup" {
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos_agg.spc
  run steampipe plugin uninstall steampipe
  rm -f $STEAMPIPE_INSTALL_DIR/config/steampipe.spc
}

function setup_file() {
  export BATS_TEST_TIMEOUT=180
  echo "# setup_file()">&3
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}