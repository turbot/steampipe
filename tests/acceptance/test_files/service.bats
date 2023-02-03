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
