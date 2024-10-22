load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe service start" {
    run steampipe service start
    assert_success
}

@test "steampipe service restart" {
    run steampipe service restart
    assert_success
}

@test "steampipe service stop" {
    run steampipe service stop
    assert_success
}

@test "custom database name" {
  # Set the STEAMPIPE_INITDB_DATABASE_NAME env variable 
  export STEAMPIPE_INITDB_DATABASE_NAME="custom_db_name"
  
  target_install_directory=$(mktemp -d)
  
  # Start the service
  run steampipe service start --install-dir $target_install_directory
  echo $output
  # Check if database name in the output is the same
  assert_output --partial 'custom_db_name'
  
  # Extract password from the state file
  db_name=$(cat $target_install_directory/internal/steampipe.json | jq .database)
  echo $db_name
  
  # Both should be equal
  assert_equal "$db_name" "\"custom_db_name\""
  
  run steampipe service stop --install-dir $target_install_directory
  
  rm -rf $target_install_directory
}

@test "custom database name - should not start with uppercase characters" {
  # Set the STEAMPIPE_INITDB_DATABASE_NAME env variable
  export STEAMPIPE_INITDB_DATABASE_NAME="Custom_db_name"
  
  target_install_directory=$(mktemp -d)
  
  # Start the service
  run steampipe service start --install-dir $target_install_directory
  
  assert_failure
  run steampipe service stop --force
  rm -rf $target_install_directory
}

@test "start service and verify that passwords stored in .passwd and steampipe.json are same" {
  # Start the service
  run steampipe service start

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .password)
  echo $state_file_pass

  # Extract password stored in .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  pass_file_pass=\"${pass_file_pass}\"
  echo "$pass_file_pass"

  # Both should be equal
  assert_equal "$state_file_pass" "$pass_file_pass"

  run steampipe service stop
}

@test "start service with --database-password flag and verify that the password used in flag and stored in steampipe.json are same" {
  # Start the service with --database-password flag
  run steampipe service start --database-password "abcd-efgh-ijkl"

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .password)
  echo $state_file_pass

  # Both should be equal
  assert_equal "$state_file_pass" "\"abcd-efgh-ijkl\""

  run steampipe service stop
}

@test "start service with password in env variable and verify that the password used in env and stored in steampipe.json are same" {
  # Set the STEAMPIPE_DATABASE_PASSWORD env variable
  export STEAMPIPE_DATABASE_PASSWORD="dcba-hgfe-lkji"

  # Start the service
  run steampipe service start

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .password)
  echo $state_file_pass

  # Both should be equal
  assert_equal "$state_file_pass" "\"dcba-hgfe-lkji\""

  run steampipe service stop
}

@test "start service with --database-password flag and env variable set, verify that the password used in flag gets higher precedence and is stored in steampipe.json" {
  # Set the STEAMPIPE_DATABASE_PASSWORD env variable
  export STEAMPIPE_DATABASE_PASSWORD="dcba-hgfe-lkji"

  # Start the service with --database-password flag
  run steampipe service start --database-password "abcd-efgh-ijkl"

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .password)
  echo $state_file_pass

  # Both should be equal
  assert_equal "$state_file_pass" "\"abcd-efgh-ijkl\""

  run steampipe service stop
}

@test "start service after removing .passwd file, verify new .passwd file gets created and also passwords stored in .passwd and steampipe.json are same" {
  # Remove the .passwd file
  rm -f $STEAMPIPE_INSTALL_DIR/internal/.passwd

  # Start the service
  run steampipe service start

  # Extract password from the state file
  state_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .password)
  echo $state_file_pass

  # Extract password stored in new .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  pass_file_pass=\"${pass_file_pass}\"
  echo "$pass_file_pass"

  # Both should be equal
  assert_equal "$state_file_pass" "$pass_file_pass"

  run steampipe service stop
}

@test "start service with --database-password flag and verify that the password used in flag is not stored in .passwd file" {
  # Start the service with --database-password flag
  run steampipe service start --database-password "abcd-efgh-ijkl"

  # Extract password stored in .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  echo "$pass_file_pass"

  # Both should not be equal
  if [[ "$pass_file_pass" != "abcd-efgh-ijkl" ]]
  then
    temp=1
  fi

  assert_equal "$temp" "1"

  run steampipe service stop
}

@test "start service with password in env variable and verify that the password used in env is not stored in .passwd file" {
  # Set the STEAMPIPE_DATABASE_PASSWORD env variable
  export STEAMPIPE_DATABASE_PASSWORD="dcba-hgfe-lkji"

  # Start the service
  run steampipe service start

  # Extract password stored in .passwd file
  pass_file_pass=$(cat $STEAMPIPE_INSTALL_DIR/internal/.passwd)
  echo "$pass_file_pass"

  # Both should not be equal
  if [[ "$pass_file_pass" != "dcba-hgfe-lkji" ]]
  then
    temp=1
  fi

  assert_equal "$temp" "1"
  
  run steampipe service stop
}

## service extensions

# tests for tablefunc module

@test "test crosstab function" {
  # create table and insert values
  steampipe query "CREATE TABLE ct(id SERIAL, rowid TEXT, attribute TEXT, value TEXT);"
  steampipe query "INSERT INTO ct(rowid, attribute, value) VALUES('test1','att1','val1');"
  steampipe query "INSERT INTO ct(rowid, attribute, value) VALUES('test1','att2','val2');"
  steampipe query "INSERT INTO ct(rowid, attribute, value) VALUES('test1','att3','val3');"

  # crosstab function
  run steampipe query "SELECT * FROM crosstab('select rowid, attribute, value from ct where attribute = ''att2'' or attribute = ''att3'' order by 1,2') AS ct(row_name text, category_1 text, category_2 text);"
  echo $output

  # drop table
  steampipe query "DROP TABLE ct"

  # match output with expected
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_crosstab_results.txt)"
}

@test "test normal_rand function" {
  # normal_rand function
  run steampipe query "SELECT * FROM normal_rand(10, 5, 3);"

  # previous query should pass
  assert_success
}

@test "verify installed fdw version" {
  run steampipe query "select * from steampipe_internal.steampipe_server_settings" --output=json

  # extract the first mod_name from the list
  fdw_version=$(echo $output | jq '.rows[0].fdw_version')
  desired_fdw_version=$(cat $STEAMPIPE_INSTALL_DIR/db/versions.json | jq '.fdw_extension.version')

  assert_equal "$fdw_version" "$desired_fdw_version"
}

@test "service stability" {
  echo "# Setting up"
  steampipe query "select 1"
  echo "# Setup Done"
  echo "# Executing tests"

  # pick up the test definitions
  tests=$(cat $FILE_PATH/test_data/source_files/service.json)

  test_indices=$(echo $tests | jq '. | keys[]')

  cd $FILE_PATH/test_data/mods/service_mod

  # prepare a sample sql file
  echo 'select 1' > sample.sql

  # loop through the tests
  for i in $test_indices; do
    test_name=$(echo $tests | jq -c ".[${i}]" | jq ".name")
    echo ">>> TEST NAME: '$test_name'"
    # pick up the commands that need to run for this test
    runs=$(echo $tests | jq -c ".[${i}]" | jq ".run")

    # get the indices of the commands to run
    run_indices=$(echo $runs | jq '. | keys[]')

    for k in 1..10; do
      # loop through the run indices
      for j in $run_indices; do
        cmd=$(echo $runs | jq ".[${j}]" | tr -d '"')
        echo ">>>>>>Command: $cmd"
        # run the command
        run $command

        # make sure that the command executed successfully
        assert_success
      done

      # make sure that there are no steampipe service processes running
      assert_equal $(ps aux | grep steampipe | grep -v bats |grep -v grep | wc -l | tr -d ' ') 0
    done
  done

  # remove the sample sql file
  rm -f sample.sql
}

@test "steampipe test database config with default listen option(hcl)" {
  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq -c '.listen | index("'$IPV4_ADDR'")')
  echo $listen

  assert_not_equal "$listen" "null"

  run steampipe service stop

  assert_success
}

@test "steampipe test database config with local listen option(hcl)" {
  skip "TODO - fix test"
  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc
  sed -i.bak 's/LISTEN_PLACEHOLDER/local/' $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq -c .listen)
  echo $listen

  assert_equal "$listen" '["127.0.0.1","::1","localhost"]'

  run steampipe service stop

  # remove the config file
  rm -f $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc{,.bak}

  assert_success
}

@test "steampipe test database config with network listen option(hcl)" {
  skip "TODO - fix test"
  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc
  sed -i.bak 's/LISTEN_PLACEHOLDER/network/' $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq -c '.listen | index("'$IPV4_ADDR'")')
  echo $listen

  assert_not_equal "$listen" "null"

  run steampipe service stop

  # remove the config file
  rm -f $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc{,.bak}

  assert_success
}

@test "steampipe test database config with listen IPv4 loopback option(hcl)" {
  skip "TODO - fix test"
  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc
  sed -i.bak 's/LISTEN_PLACEHOLDER/127.0.0.1/' $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq -c .listen)
  echo $listen

  assert_equal "$listen" '["127.0.0.1"]'

  run steampipe service stop

  # remove the config file
  rm -f $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc{,.bak}

  assert_success
}

@test "steampipe test database config with listen IPv6 loopback option(hcl)" {
  skip "TODO - fix test"
  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc
  sed -i.bak 's/LISTEN_PLACEHOLDER/::1/' $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq -c .listen)
  echo $listen

  assert_equal "$listen" '["127.0.0.1","::1"]'

  run steampipe service stop

  # remove the config file
  rm -f $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc{,.bak}

  assert_success
}

@test "steampipe test database config with listen IPv4 address option(hcl)" {
  skip "TODO - fix test"
  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  sed -i.bak "s/LISTEN_PLACEHOLDER/$IPV4_ADDR/" $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq -c '.listen | index("'$IPV4_ADDR'")')
  echo $listen

  assert_not_equal "$listen" "null"

  run steampipe service stop

  # remove the config file
  rm -f $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc{,.bak}

  assert_success
}

@test "steampipe test database config with listen IPv6 address option(hcl)" {
  if [ -z "$IPV6_ADDR" ]; then
    skip "No IPv6 address is available, skipping test."
  fi

  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc
  sed -i.bak "s/LISTEN_PLACEHOLDER/$IPV6_ADDR/" $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq -c .listen)
  echo $listen

  assert_equal "$listen" '["127.0.0.1","'$IPV6_ADDR'"]'

  run steampipe service stop

  # remove the config file
  rm -f $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc{,.bak}

  assert_success
}

@test "verify steampipe_connection_state table is getting properly migrated" {
  skip "needs updating when new migration is complete"

  # create a temp directory to install steampipe(0.13.6)
  tmpdir="$(mktemp -d)"
  mkdir -p "${tmpdir}"
  tmpdir="${tmpdir%/}"

  # find the name of the zip file as per OS and arch
  case $(uname -sm) in
	"Darwin x86_64") target="darwin_amd64.zip" ;;
	"Darwin arm64") target="darwin_arm64.zip" ;;
	"Linux x86_64") target="linux_amd64.tar.gz" ;;
	"Linux aarch64") target="linux_arm64.tar.gz" ;;
	*) echo "Error: '$(uname -sm)' is not supported yet." 1>&2;exit 1 ;;
	esac

  # download the zip and extract
  steampipe_uri="https://github.com/turbot/steampipe/releases/download/v0.20.6/steampipe_${target}"
  case $(uname -s) in
    "Darwin") zip_location="${tmpdir}/steampipe.zip" ;;
    "Linux") zip_location="${tmpdir}/steampipe.tar.gz" ;;
    *) echo "Error: steampipe is not supported on '$(uname -s)' yet." 1>&2;exit 1 ;;
  esac
  curl --fail --location --progress-bar --output "$zip_location" "$steampipe_uri"
  tar -xf "$zip_location" -C "$tmpdir"

  # install a couple of plugins which can work with default config
  $tmpdir/steampipe --install-dir $tmpdir plugin install chaos net --progress=false
  $tmpdir/steampipe --install-dir $tmpdir query "select * from steampipe_internal.steampipe_connection_state" --output json

  run steampipe --install-dir $tmpdir query "select * from steampipe_internal.steampipe_connection_state" --output json

  rm -rf $tmpdir

  assert_success
}

function setup_file() {
  export BATS_TEST_TIMEOUT=180
  echo "# setup_file()">&3
  export IPV4_ADDR=$(ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1)
  export IPV6_ADDR=$(ifconfig | grep -Eo 'inet6 (addr:)?([0-9a-f]*:){7}[0-9a-f]*' | grep -Eo '([0-9a-f]*:){7}[0-9a-f]*' | head -n 1)
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
