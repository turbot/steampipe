load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "service stability" {
  echo "# Setting up"
  steampipe query "select 1"
  echo "# Setup Done"
  echo "# Executing tests"
  while read -r name run; do
    echo "## Running $name"
    cd $FILE_PATH
    while read -r cmd; do
      echo "### Running '$cmd'"
      STEAMPIPE_LOG=trace
      run $cmd
      assert_success
      assert_equal $(ps aux | grep steampipe | grep -v bats |grep -v grep | wc -l | tr -d ' ') 0
    done< <(echo $run | jq --raw-output '.[] | @sh' | tr -d \')
  done< <(cat $FILE_PATH/test_data/source_files/service.json | jq --raw-output '.[] | "\(.name) \(.run)"')
  echo "# Execution Done"
  
  fail "inject failure"
}

@test "service stability 2" {
  echo "# Setting up"
  steampipe query "select 1"
  echo "# Setup Done"
  echo "# Executing tests"

  tests=$(cat $FILE_PATH/test_data/source_files/service.json)
  test_keys=$(echo $tests | jq '. | keys[]')

  for i in $test_keys; do
    test_name=$(echo $tests | jq -c ".[${i}]" | jq ".name")
    echo "### Running '$test_name'"
    runs=$(echo $tests | jq -c ".[${i}]" | jq ".run")
    echo $runs
    run_keys=$(echo $runs | jq '. | keys[]')

    for j in $run_keys; do
      command=''
      cmd=$(echo $runs | jq ".[${j}]" | tr -d '"')
      # echo $cmd
      command="${command}${cmd}"
      echo $command
      # $command
      # echo $output
    done
  done
  assert_equal 1 0
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