load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe query fail with non-0 exit code" {
  # this query should fail with a non 0 exit code
  run steampipe query "select * from abc"
  echo $status
  [ $status -ne 0 ]
}

@test "steampipe check fail with non-0 exit code" {
  cd $WORKSPACE_DIR
  # this check should fail with a non 0 exit code, due to insufficient args
  run steampipe check
  echo $status
  [ $status -ne 0 ]
}

@test "steampipe plugin command fail with insufficient arguments" {
  # this should return a non 0 exit code, due to insufficient args
  run steampipe plugin install 
  echo $status
  [ $status -ne 0 ]
}

@test "steampipe query pass with 0 exit code" {
  # this query should pass and return a 0 exit code
  run steampipe query "select 1"
  echo $status
  [ $status -eq 0 ]
}
