load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe_introspection=none" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=none
  run steampipe query "select * from steampipe_query" --output json

  assert_output --partial 'relation "steampipe_query" does not exist'
}

@test "steampipe_introspection=info" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_query" --output json

  assert_output --partial '"resource_name": "sample_query_1"'
}

@test "steampipe_introspection=control" {
  cd $SIMPLE_MOD_DIR
  export STEAMPIPE_INTROSPECTION=info
  run steampipe query "select * from steampipe_control" --output json

  assert_output --partial '"resource_name": "sample_control_1"'
}

