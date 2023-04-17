load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "generic mod require test" {

  # setup test folder and read the test-cases file
  cd $FILE_PATH/test_data/mod_install
  tests=$(cat mod_require_tests.json)
  # echo $tests

  # to create the failure message
  err=""
  flag=0

  # fetch the keys(test names)
  test_keys=$(echo $tests | jq '. | keys[]')
  # echo $test_keys

  for i in $test_keys; do
    # for each test case do the following

    # Extract the test_name, modsp, cmd and expected properties from the test key
    test_name=$(echo $tests | jq -c ".[${i}]" | jq ".test_name")
    modsp=$(echo $tests | jq -c ".[${i}]" | jq ".modsp")
    cmd=$(echo $tests | jq -c ".[${i}]" | jq ".cmd")
    expected=$(echo $tests | jq -c ".[${i}]" | jq ".expected")

    # Remove the first and last quotes from the modsp property
    modsp="${modsp%\"}"
    modsp="${modsp#\"}"

    # Create the mod.sp file
    file_name="mod.sp"
    touch "$file_name"

    # Write the contents of the modsp property into the mod.sp file
    echo "$modsp" | sed -e 's/\\n/\n/g' -e 's/\\"/"/g' > "$file_name"

    # cat $file_name

    # Remove the first and last quotes from the steampipe command and expected
    cmd="${cmd%\"}"
    cmd="${cmd#\"}"
    expected="${expected%\"}"
    expected="${expected#\"}"

    # run the steampipe command
    run $cmd

    echo "Checking >> $test_name"
    assert_output --partial "$expected"
  done
}

function teardown() {
  cd $FILE_PATH/test_data/mod_install
  rm -rf .steampipe/
  rm -f .mod.cache.json
  rm -f mod.sp
}
