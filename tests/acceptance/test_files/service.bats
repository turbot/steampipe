load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "implicit service from query" {
  for i in {1..10}
  do
    # setup
    run steampipe query "select 1"
    
    # run a sleep in the background
    # this should sstart up an implicit service
    run steampipe query "select pg_sleep(10)" &
    
    # execute a constants 10 times
    for i in {1..10}
    do
      run steampipe query "select 1 as col"
    done
    sleep 10
    
    assert_equal $(ps aux | grep steampipe | grep -v bats | wc -l | tr -d ' ') 0 # 1 because of the grep process itself
  done
}

@test "implicit service from check" {
  for i in {1..10}
  do
    # setup
    run steampipe query "select 1"
    
    # change to the service mod directory
    cd $FILE_PATH/test_data/service_mod
    
    # run check all - duration 20 seconds
    run steampipe check all &
    
    # execute another query 10 times - so that steampipe cycles repeatedly
    for i in {1..10}
    do
      run steampipe query "select 1 as col"
    done
    # sleep as long as the initial steampipe instance will take
    sleep 20
    
    assert_equal $(ps aux | grep steampipe | grep -v bats | wc -l | tr -d ' ') 0 # 1 because of the grep process itself
  done
}