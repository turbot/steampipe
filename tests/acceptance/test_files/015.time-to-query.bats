load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "time to query a chaos table" {

  # using bash's built-in time, set the timeformat to seconds
  TIMEFORMAT=%R

  # find the query time
  QUERY_TIME=$(time (run steampipe query "select time_now from chaos.chaos_cache_check where id=0" >/dev/null 2>&1) 2>&1)
  echo $QUERY_TIME
  echo $TIME_TO_QUERY

  # check whether time to query is less than 2 seconds(This value can be changed)
  assert_equal "$(echo $QUERY_TIME '<' $TIME_TO_QUERY | bc -l)" "1"
}

@test "time to query a chaos table that does not exist" {

  # using bash's built-in time, set the timeformat to seconds
  TIMEFORMAT=%R

  # find the time it takes to throw the error
  QUERY_TIME=$(time (run steampipe query "select time_now from chaos.chaos_cache_check_2 where id=0" >/dev/null 2>&1) 2>&1)
  echo $QUERY_TIME
  echo $TIME_TO_QUERY

  # check whether time to error out is less than 2 seconds(This value can be changed)
  assert_equal "$(echo $QUERY_TIME '<' $TIME_TO_QUERY | bc -l)" "1"
}
