load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "verify installed fdw version" {
  run steampipe query "select * from steampipe_internal.steampipe_server_settings" --output=json

  # extract the first mod_name from the list
  fdw_version=$(echo $output | jq '.[0].fdw_version')
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

  cd $FILE_PATH/test_data/service_mod

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
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq -c .listen)
  echo $listen

  assert_equal "$listen" '["127.0.0.1","::1","'$IPV4_ADDR'"]'

  run steampipe service stop

  assert_success
}

@test "steampipe test database config with local listen option(hcl)" {
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
  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc
  sed -i.bak 's/LISTEN_PLACEHOLDER/network/' $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq -c .listen)
  echo $listen

  assert_equal "$listen" '["127.0.0.1","::1","'$IPV4_ADDR'"]'

  run steampipe service stop

  # remove the config file
  rm -f $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc{,.bak}

  assert_success
}

@test "steampipe test database config with listen IPv4 loopback option(hcl)" {
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
  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  sed -i.bak "s/LISTEN_PLACEHOLDER/$IPV4_ADDR/" $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq -c .listen)
  echo $listen

  assert_equal "$listen" '["127.0.0.1","'$IPV4_ADDR'"]'

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
  $tmpdir/steampipe --install-dir $tmpdir plugin install chaos net
  $tmpdir/steampipe --install-dir $tmpdir query "select * from steampipe_internal.steampipe_connection_state" --output json

  run steampipe --install-dir $tmpdir query "select * from steampipe_internal.steampipe_connection_state" --output json

  rm -rf $tmpdir

  assert_success
}

function setup_file() {
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
