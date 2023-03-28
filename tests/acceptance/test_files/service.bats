load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

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
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .listen)
  echo $listen

  assert_equal "$listen" '"localhost"'

  run steampipe service stop

  assert_success
}

@test "steampipe test database config with local listen option(hcl)" {
  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc
  sed -i.bak 's/LISTEN_PLACEHOLDER/local/' $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .listen)
  echo $listen

  assert_equal "$listen" '"localhost"'

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
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .listen)
  echo $listen

  assert_equal "$listen" '"*"'

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
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .listen)
  echo $listen

  assert_equal "$listen" '"127.0.0.1"'

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
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .listen)
  echo $listen

  assert_equal "$listen" '"127.0.0.1,::1"'

  run steampipe service stop

  # remove the config file
  rm -f $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc{,.bak}

  assert_success
}

@test "steampipe test database config with listen IPv4 address option(hcl)" {
  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  IPV4_ADDR=$(ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1)
  sed -i.bak "s/LISTEN_PLACEHOLDER/$IPV4_ADDR/" $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .listen)
  echo $listen

  assert_equal "$listen" '"127.0.0.1,'$IPV4_ADDR'"'

  run steampipe service stop

  # remove the config file
  rm -f $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc{,.bak}

  assert_success
}

@test "steampipe test database config with listen IPv6 address option(hcl)" {
  IPV6_ADDR=$(ifconfig | grep -Eo 'inet6 (addr:)?([0-9a-f]*:){7}[0-9a-f]*' | grep -Eo '([0-9a-f]*:){7}[0-9a-f]*' | head -n 1)

  if [ -z "$IPV6_ADDR" ]; then
    skip "No IPv6 address is available, skipping test."
  fi

  cp $SRC_DATA_DIR/database_options_listen_placeholder.spc $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc
  sed -i.bak "s/LISTEN_PLACEHOLDER/$IPV6_ADDR/" $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc

  run steampipe service start

  assert_success

  # Extract listen from the state file
  listen=$(cat $STEAMPIPE_INSTALL_DIR/internal/steampipe.json | jq .listen)
  echo $listen

  assert_equal "$listen" '"127.0.0.1,'$IPV6_ADDR'"'

  run steampipe service stop

  # remove the config file
  rm -f $STEAMPIPE_INSTALL_DIR/config/database_options_listen.spc{,.bak}

  assert_success
}
