load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe query fail with non-0 exit code" {
  # this query should fail with a non 0 exit code
  run steampipe query "select * from abc"
  echo $status
  [ $status -ne 0 ]
}

@test "steampipe query pass with 0 exit code" {
  # this query should pass and return a 0 exit code
  run steampipe query "select 1"
  echo $status
  [ $status -eq 0 ]
}

@test "steampipe nonexistant pass with 1 exit code" {
  # this command should exit one since nonexistent does not exist 
  run steampipe nonexistant
  echo $status
  [ $status -eq 1 ]
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
