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
    # pick u p the commands that need to run for this test
    runs=$(echo $tests | jq -c ".[${i}]" | jq ".run")
    
    # get the indices of the commands to run
    run_keys=$(echo $runs | jq '. | keys[]')
    
    for k in 1..10; do
      # loop through the run indices
      for j in $run_keys; do
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

  # remove the samplke sql file
  rm -f sample.sql
}

# @test "implicit service from query" {
#   for i in {1..10}
#   do
#     # setup
#     run steampipe query "select 1"
    
#     # run a sleep in the background
#     # this should sstart up an implicit service
#     run steampipe query "select pg_sleep(10)" &
    
#     # execute a constants 10 times
#     for i in {1..10}
#     do
#       run steampipe query "select 1 as col"
#     done
#     sleep 10
    
#     assert_equal $(ps aux | grep steampipe | grep -v bats | wc -l | tr -d ' ') 0 # 1 because of the grep process itself
#   done
# }

# @test "implicit service from check" {
#   for i in {1..10}
#   do
#     # setup
#     run steampipe query "select 1"
    
#     # change to the service mod directory
#     cd $FILE_PATH/test_data/service_mod
    
#     # run check all - duration 20 seconds
#     run steampipe check all &
    
#     # execute another query 10 times - so that steampipe cycles repeatedly
#     for i in {1..10}
#     do
#       run steampipe query "select 1 as col"
#     done
#     # sleep as long as the initial steampipe instance will take
#     sleep 20
    
#     assert_equal $(ps aux | grep steampipe | grep -v bats | wc -l | tr -d ' ') 0 # 1 because of the grep process itself
#   done
# }