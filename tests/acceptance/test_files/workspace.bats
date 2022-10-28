load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

## workspace tests

@test "generic workspace test" {
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
    export STEAMPIPE_DIAGNOSTICS=config_json
    cmd='steampipe query "select 1"'
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

    # construct the steampipe query command to be run with the args
    for arg in $(echo "${args}" | jq -r '.[]'); do
      cmd="${cmd} ${arg}"
    done
    # echo $cmd

    # get the actual config by running the constructed steampipe command
    run $cmd
    actual_config=$(echo $output | jq -c '.')
    # echo $actual_config

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
