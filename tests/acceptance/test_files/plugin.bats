load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "plugin install" {
  run steampipe plugin install chaos
  assert_success
  steampipe plugin uninstall chaos
}

@test "plugin install from stream" {
  run steampipe plugin install chaos@0.4
  assert_success
  steampipe plugin uninstall chaos@0.4
}

@test "plugin install from stream (prefixed with v)" {
  run steampipe plugin install chaos@v0.4
  assert_success
  steampipe plugin uninstall chaos@0.4
}

@test "plugin install from caret constraint" {
  run steampipe plugin install chaos@^0.4
  assert_success
  steampipe plugin uninstall chaos@^0.4
}

@test "plugin install from tilde constraint" {
  run steampipe plugin install chaos@~0.4.0
  assert_success
  steampipe plugin uninstall chaos@~0.4.0
}

@test "plugin install from wildcard constraint" {
  run steampipe plugin install chaos@0.4.*
  assert_success
  steampipe plugin uninstall chaos@0.4.*
}

@test "plugin install gte constraint" {
  run steampipe plugin install "chaos@>=0.4"
  assert_success
  steampipe plugin uninstall "chaos@>=0.4"
}

@test "create a local plugin, add connection and query" {
  run steampipe plugin install chaos

  # create a local plugin directory
  mkdir $STEAMPIPE_INSTALL_DIR/plugins/local
  mkdir $STEAMPIPE_INSTALL_DIR/plugins/local/myplugin
  # use the chaos plugin binary to get a plugin binary for the local plugin
  cp $STEAMPIPE_INSTALL_DIR/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/steampipe-plugin-chaos.plugin $STEAMPIPE_INSTALL_DIR/plugins/local/myplugin/myplugin.plugin
  # create a connection config file for the new local plugin
  echo "connection \"myplugin\" {
    plugin = \"local/myplugin\"
  }" > $STEAMPIPE_INSTALL_DIR/config/myplugin.spc

  run steampipe query "select * from myplugin.chaos_all_column_types"
  assert_success
  run steampipe plugin list
  assert_output --partial "local/myplugin"
}

@test "start service, install plugin and query" {
  skip
  # start service
  steampipe service start

  # install plugin
  steampipe plugin install chaos

  steampipe query "select 1"

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

@test "plugin list - output table and json" {
  export STEAMPIPE_DISPLAY_WIDTH=100

  # Create a copy of the install directory
  copy_install_directory

  steampipe plugin install hackernews@0.8.0 bitbucket@0.7.1 --progress=false --install-dir $MY_TEST_COPY

  # check table output
  run steampipe plugin list --install-dir $MY_TEST_COPY

  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_table.txt)"

  # check json output
  steampipe plugin list --install-dir $MY_TEST_COPY --output json > output.json
  run jd $TEST_DATA_DIR/expected_plugin_list_json.json output.json
  echo $output
  assert_success
  rm -rf $MY_TEST_COPY
}

@test "plugin list - output table and json (with a missing plugin)" {
  export STEAMPIPE_DISPLAY_WIDTH=100

  # Create a copy of the install directory
  copy_install_directory

  steampipe plugin install hackernews@0.8.0 bitbucket@0.7.1 --progress=false --install-dir $MY_TEST_COPY
  # uninstall a plugin but dont remove the config - to simulate the missing plugin scenario
  steampipe plugin uninstall hackernews@0.8.0 --install-dir $MY_TEST_COPY

  # check table output
  run steampipe plugin list --install-dir $MY_TEST_COPY
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_table_with_missing_plugins.txt)"

  # check json output
  steampipe plugin list --install-dir $MY_TEST_COPY --output json > output.json

  run jd $TEST_DATA_DIR/expected_plugin_list_json_with_missing_plugins.json output.json
  echo $output
  assert_success
  rm -rf $MY_TEST_COPY
}

# # TODO: finds other ways to simulate failed plugins

@test "plugin list - output table and json (with a failed plugin)" {
  skip "finds other ways to simulate failed plugins"
  export STEAMPIPE_DISPLAY_WIDTH=100
  
  # Create a copy of the install directory
  copy_install_directory

  steampipe plugin install hackernews@0.8.0 bitbucket@0.7.1 --progress=false --install-dir $MY_TEST_COPY
  # remove the contents of a plugin execuatable to simulate the failed plugin scenario
  cat /dev/null > $MY_TEST_COPY/plugins/hub.steampipe.io/plugins/turbot/hackernews@0.8.0/steampipe-plugin-hackernews.plugin

  # check table output
  run steampipe plugin list --install-dir $MY_TEST_COPY
  echo $output
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_plugin_list_table_with_failed_plugins.txt)"

  # check json output
  steampipe plugin list --install-dir $MY_TEST_COPY --output json > output.json
  run jd $TEST_DATA_DIR/expected_plugin_list_json_with_failed_plugins.json output.json
  echo $output
  assert_success
  rm -rf $MY_TEST_COPY
}

@test "verify that installing plugins creates individual version.json files" {
  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install net chaos --install-dir $MY_TEST_COPY
  assert_success
  
  vFile1="$MY_TEST_COPY/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  vFile2="$MY_TEST_COPY/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/version.json"
  
  [ ! -f $vFile1 ] && fail "could not find $vFile1"
  [ ! -f $vFile2 ] && fail "could not find $vFile2"
  
  rm -rf $MY_TEST_COPY
}

@test "verify that backfilling of individual plugin version.json works" {
  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install net chaos --install-dir $MY_TEST_COPY
  assert_success
  
  vFile1="$MY_TEST_COPY/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  vFile2="$MY_TEST_COPY/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/version.json"
  
  file1Content=$(cat $vFile1)
  file2Content=$(cat $vFile2)
  
  # remove the individual version files
  rm -f $vFile1
  rm -f $vFile2
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $MY_TEST_COPY
  
  [ ! -f $vFile1 ] && fail "could not find $vFile1"
  [ ! -f $vFile2 ] && fail "could not find $vFile2"
  echo "$file1Content" > $MY_TEST_COPY/f1.json
  echo "$file2Content" > $MY_TEST_COPY/f2.json
  cat "$vFile1" > $MY_TEST_COPY/v1.json
  cat "$vFile2" > $MY_TEST_COPY/v2.json

  # Compare the json file contents
  run jd "$MY_TEST_COPY/f1.json" "$MY_TEST_COPY/v1.json"
  echo $output
  assert_success

  run jd "$MY_TEST_COPY/f2.json" "$MY_TEST_COPY/v2.json"
  echo $output
  assert_success
  rm -rf $MY_TEST_COPY
}

@test "verify that backfilling of individual plugin version.json works where it is only partially backfilled" {
  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install net chaos --install-dir $MY_TEST_COPY
  assert_success
  
  vFile1="$MY_TEST_COPY/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  vFile2="$MY_TEST_COPY/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/version.json"
  
  file1Content=$(cat $vFile1)
  file2Content=$(cat $vFile2)
  
  # remove one individual version file
  rm -f $vFile1
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $MY_TEST_COPY
  
  [ ! -f $vFile1 ] && fail "could not find $vFile1"
  [ ! -f $vFile2 ] && fail "could not find $vFile2"
  
  echo "$file1Content" > $MY_TEST_COPY/f1.json
  echo "$file2Content" > $MY_TEST_COPY/f2.json
  cat "$vFile1" > $MY_TEST_COPY/v1.json
  cat "$vFile2" > $MY_TEST_COPY/v2.json

  # Compare the json file contents
  run jd "$MY_TEST_COPY/f1.json" "$MY_TEST_COPY/v1.json"
  echo $output
  assert_success

  run jd "$MY_TEST_COPY/f2.json" "$MY_TEST_COPY/v2.json"
  echo $output
  assert_success
  
  rm -rf $MY_TEST_COPY
}

@test "verify that global plugin/versions.json is composed from individual version.json files when it is absent" {
  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install net chaos --install-dir $MY_TEST_COPY
  assert_success
  
  vFile="$MY_TEST_COPY/plugins/versions.json"
  
  fileContent=$(cat $vFile)
  
  # remove global version file
  rm -f $vFile
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $MY_TEST_COPY
  
  ls -la $vFile
  
  [ ! -f $vFile ] && fail "could not find $vFile"
  
  echo "$fileContent" > $MY_TEST_COPY/f.json
  cat "$vFile" > $MY_TEST_COPY/v.json

  # Compare the json file contents
  run jd "$MY_TEST_COPY/f.json" "$MY_TEST_COPY/v.json"
  echo $output
  assert_success

  rm -rf $MY_TEST_COPY
}

@test "verify that global plugin/versions.json is composed from individual version.json files when it is corrupt" {
  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install net chaos --install-dir $MY_TEST_COPY
  assert_success
  
  vFile="$MY_TEST_COPY/plugins/versions.json"
  fileContent=$(cat $vFile)
  
  # remove global version file
  echo "badline to corrupt versions.json" >> $vFile
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $MY_TEST_COPY
  
  [ ! -f $vFile ] && fail "could not find $vFile"
  
  echo "$fileContent" > $MY_TEST_COPY/f.json
  cat "$vFile" > $MY_TEST_COPY/v.json

  # Compare the json file contents
  run jd "$MY_TEST_COPY/f.json" "$MY_TEST_COPY/v.json"
  echo $output
  assert_success
  
  rm -rf $MY_TEST_COPY
}

@test "verify that composition of global plugin/versions.json works when an individual version.json file is corrupt" {
  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install net chaos --install-dir $MY_TEST_COPY
  assert_success
  
  vFile="$MY_TEST_COPY/plugins/versions.json"  
  vFile1="$MY_TEST_COPY/plugins/hub.steampipe.io/plugins/turbot/net@latest/version.json"
  
  # corrupt a version file
  echo "bad line to corrupt" >> $vFile1
  
  # remove global file
  rm -f $vFile
  
  # run steampipe again so that the plugin version files get backfilled
  run steampipe plugin list --install-dir $MY_TEST_COPY

  # verify that global file got created
  [ ! -f $vFile ] && fail "could not find $vFile"
  
  rm -rf $MY_TEST_COPY
}

@test "verify that plugin installed from registry are marked as 'local' when the modtime of the binary is after the install time" {
  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install net chaos --install-dir $MY_TEST_COPY
  assert_success

  # wait for a couple of seconds
  sleep 2

  # touch one of the plugin binaries
  touch $MY_TEST_COPY/plugins/hub.steampipe.io/plugins/turbot/net@latest/steampipe-plugin-net.plugin

  # run steampipe again so that the plugin version files get backfilled
  version=$(steampipe plugin list --install-dir $MY_TEST_COPY --output json | jq '.installed' | jq '. | map(select(.name | contains("net@latest")))' | jq '.[0].version')

  # assert
  assert_equal "$version" '"local"'

  rm -rf $MY_TEST_COPY
}

@test "verify that steampipe check should bypass plugin requirement detection if installed plugin is local" {
  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install net --install-dir $MY_TEST_COPY
  assert_success

  # wait for a couple of seconds
  sleep 2

  # touch one of the plugin binaries
  touch $MY_TEST_COPY/plugins/hub.steampipe.io/plugins/turbot/net@latest/steampipe-plugin-net.plugin

  run steampipe plugin list --install-dir $MY_TEST_COPY
  echo $output

  # clone a mod which has a net plugin requirement
  cd $MY_TEST_COPY
  git clone https://github.com/turbot/steampipe-mod-net-insights.git
  cd steampipe-mod-net-insights

  # run steampipe check
  run steampipe check all --install-dir $MY_TEST_COPY

  # check - the plugin requirement warning should not be present in the output
  substring="Warning: could not find plugin which satisfies requirement"
  if [[ ! $output == *"$substring"* ]]; then
    run echo "Warning is not present in the output"
  else
    run echo "Warning is present in the output"
  fi

  assert_equal "$output" "Warning is not present in the output"
  rm -rf $MY_TEST_COPY
}

@test "verify that plugin installed with --skip-config as true, should not have create a default config .spc file in config folder" {
  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install aws --skip-config --install-dir $MY_TEST_COPY
  assert_success

  run test -f $MY_TEST_COPY/config/aws.spc
  assert_failure

  rm -rf $MY_TEST_COPY
}

@test "verify that plugin installed with --skip-config as false(default), should have default config .spc file in config folder" {
  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install aws --install-dir $MY_TEST_COPY
  assert_success

  run test -f $MY_TEST_COPY/config/aws.spc
  assert_success

  rm -rf $MY_TEST_COPY
}

@test "verify reinstalling a plugin does not overwrite existing plugin config" {
  # check if the default/tweaked config file for a plugin is not deleted after
  # re-installation of a plugin

  # Create a copy of the install directory
  copy_install_directory

  run steampipe plugin install aws --install-dir $MY_TEST_COPY

  run test -f $MY_TEST_COPY/config/aws.spc
  assert_success

  echo '
  connection "aws" {
    plugin = "aws"
    endpoint_url = "http://localhost:4566"
  }
  ' >> $MY_TEST_COPY/config/aws.spc
  cp $MY_TEST_COPY/config/aws.spc config.spc

  run steampipe plugin uninstall aws --install-dir $MY_TEST_COPY

  run steampipe plugin install aws --skip-config --install-dir $MY_TEST_COPY

  run test -f $MY_TEST_COPY/config/aws.spc
  assert_success

  run diff $MY_TEST_COPY/config/aws.spc config.spc
  assert_success

  rm config.spc
  rm -rf $MY_TEST_COPY
}

# Custom function to create a copy of the install directory
copy_install_directory() {
  cp -r "$MY_TEST_DIRECTORY" "/tmp/test_copy"
  export MY_TEST_COPY="/tmp/test_copy"
}

function setup_file() {
  export BATS_TEST_TIMEOUT=180
  echo "# setup_file()">&3

  tmpdir="$(mktemp -d)"
  steampipe query "select 1" --install-dir $tmpdir
  # Export the directory path as an environment variable
  export MY_TEST_DIRECTORY=$tmpdir
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
