load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

## workspace tests

@test "generic config precedence test" {
  cp $FILE_PATH/test_data/source_files/config_tests/default.spc $STEAMPIPE_INSTALL_DIR/config/default.spc
  
  # setup test folder and read the test-cases file
  cd $FILE_PATH/test_data/source_files/config_tests
  tests=$(cat workspace_tests.json)
  # echo $tests

  # to create the failure message
  err=""
  flag=0

  # fetch the keys(test names)
  test_keys=$(echo $tests | jq '. | keys[]')
  # echo $test_keys

  for i in $test_keys; do
    # each test case do the following
    unset STEAMPIPE_INSTALL_DIR
    cwd=$(pwd)
    export STEAMPIPE_CONFIG_DUMP=config_json

    # check the command(query/check/dashboard) and prepare the steampipe
    # command accordingly
    cmd=$(echo $tests | jq -c ".[${i}]" | jq ".cmd")
    if [[ $cmd == '"query"' ]]; then
      sp_cmd='steampipe query "select 1"'
    elif [[ $cmd == '"check"' ]]; then
      sp_cmd='steampipe check all'
    elif [[ $cmd == '"dashboard"' ]]; then
      sp_cmd='steampipe dashboard'
    fi
    # echo $sp_cmd

    # key=$(echo $i)
    echo -e "\n"
    test_name=$(echo $tests | jq -c ".[${i}]" | jq ".test")
    echo ">>> TEST NAME: $test_name"

    # env variables needed for setup
    env=$(echo $tests | jq -c ".[${i}]" | jq ".setup.env")
    # echo $env

    # set env variables
    for e in $(echo "${env}" | jq -r '.[]'); do
      export $e
    done

    # args to run with steampipe query command
    args=$(echo $tests | jq -c ".[${i}]" | jq ".setup.args")
    echo $args

    # construct the steampipe command to be run with the args
    for arg in $(echo "${args}" | jq -r '.[]'); do
      sp_cmd="${sp_cmd} ${arg}"
    done
    echo "steampipe command: $sp_cmd" # help debugging in case of failures

    # get the actual config by running the constructed steampipe command
    run $sp_cmd
    echo "output from steampipe command: $output" # help debugging in case of failures
    actual_config=$(echo $output | jq -c '.')
    echo "actual config: \n$actual_config" # help debugging in case of failures

    # get expected config from test case
    expected_config=$(echo $tests | jq -c ".[${i}]" | jq ".expected")
    # echo $expected_config

    # fetch only keys from expected config
    exp_keys=$(echo $expected_config | jq '. | keys[]' | jq -s 'flatten | @sh' | tr -d '\'\' | tr -d '"')

    for key in $exp_keys; do
      # get the expected and the actual value for the keys
      exp_val=$(echo $(echo $expected_config | jq --arg KEY $key '.[$KEY]' | tr -d '"'))
      act_val=$(echo $(echo $actual_config | jq --arg KEY $key '.[$KEY]' | tr -d '"'))

      # get the absolute paths for install-dir and mod-location
      if [[ $key == "install-dir" ]] || [[ $key == "mod-location" ]]; then
        exp_val="${cwd}/${exp_val}"
      fi
      echo "expected $key: $exp_val"
      echo "actual $key: $act_val"

      # check the values
      if [[ "$exp_val" != "$act_val" ]]; then
        flag=1
        err="FAILED: $test_name >> key: $key ; expected: $exp_val ; actual: $act_val \n${err}"
      fi
    done

    # check if all passed
    if [[ $flag -eq 0 ]]; then
      echo "PASSED ✅"
    else
      echo "FAILED ❌"
    fi
    # reset flag back to 0 for the next test case 
    flag=0
  done
  echo -e "\n"
  echo -e "$err"
  assert_equal "$err" ""
  rm -f err
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
