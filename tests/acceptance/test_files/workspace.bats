load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

## workspace tests

@test "generic test 2" {
  # setup test folder and read the test-cases file
  cd $FILE_PATH/test_data/source_files/config_tests
  tests=$(cat workspace_tests.json)

  # to create the failure message
  err=""

  while read i; do
    # each test case
    echo $i
    unset STEAMPIPE_INSTALL_DIR
    cwd=$(pwd)

    # test name
    test_name=$(echo $i | jq '.test')
    echo ">>> Running: $test_name <<<"

    # env variables needed for setup
    env=$(echo $i | jq '.setup.env')
    echo $env

    # set env variables
    for e in $(echo "${env}" | jq -r '.[]'); do
      export $e
    done

    # args to run with steampipe query command
    args=$(echo $i | jq '.setup.args')
    echo $args

    # get the diagnostics by running steampipe
    diagnostics=$(STEAMPIPE_DIAGNOSTICS=config_json steampipe query "select 1" "$args")
    echo $diagnostics

    # get expected diagnostics
    expected=$(echo $i | jq '.expected')
    # echo $expected

    # get only keys
    exp_keys=$(echo $expected | jq '. | keys[]' | jq -s 'flatten | @sh' | tr -d '\'\' | tr -d '"')
    # echo $exp_keys

    for key in $exp_keys; do
      # get the expected and the actual value for the keys
      ex_val=$(echo $(echo $expected | jq --arg KEY $key '.[$KEY]' | tr -d '"'))
      diag_val=$(echo $(echo $diagnostics | jq --arg KEY $key '.[$KEY]' | tr -d '"'))

      # get the absolute paths for install-dir and mod-location
      if [[ $key == "install-dir" ]] || [[ $key == "mod-location" ]]; then
        ex_val="${cwd}/${ex_val}"
      fi
      echo $ex_val
      echo $diag_val

      # check
      if [[ "$ex_val" != "$diag_val" ]]; then
        err="FAILED: $test_name >> key: $key ; expected: $ex_val ; actual: $diag_val"
      fi
    done

  done < <(echo $tests | jq -c -r '.[]')
  echo $err
  assert_equal "$err" ""
  rm -f err
}