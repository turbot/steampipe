load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# This test looks for a bug in the schema cloning code meaning when adding multiple connections 
# for the same plugin, only 1 of the connections will work when querying - the others will give an 
# FDW no schema loaded for connection error.
@test "schema cloning" {
  # remove existing connections
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos.spc

  # remove db, to trigger a clean installation with no connections
  rm -rf $STEAMPIPE_INSTALL_DIR/db

  # run steampipe(installs db)
  steampipe query "select 1"

  # add connections(more than 1) to trigger schema cloning
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc

  # query both connections(both should work)
  run steampipe query "select * from chaos.chaos_all_column_types"
  assert_success
  run steampipe query "select * from chaos2.chaos_all_column_types"
  assert_success
}

# This test looks for a bug in the schema cloning code where the schema clone function 
# used to fail if table had an LTREE column
@test "schema cloning - function fails if table has an LTREE column" {
  # remove existing connections
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos.spc

  # remove db, to trigger a clean installation with no connections
  rm -rf $STEAMPIPE_INSTALL_DIR/db

  # run steampipe(installs db)
  steampipe query "select 1"

  # add connections(more than 1) to trigger schema cloning
  cp $SRC_DATA_DIR/two_chaos.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc

  run steampipe query "select ltree_column from chaos2.chaos_all_column_types"
  assert_success
}

@test "schema cloning - quoting issue" {
  # remove existing connections
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos.spc

  # remove db, to trigger a clean installation with no connections
  rm -rf $STEAMPIPE_INSTALL_DIR/db

  # run steampipe(installs db)
  steampipe query "select 1"

  # add connections(more than 1 - with names containing both uppercase and lowercase chars) 
  # to trigger schema cloning
  cp $SRC_DATA_DIR/chaos_case_sensitivity.spc $STEAMPIPE_INSTALL_DIR/config/chaos.spc

  steampipe query "select 1"

  # query all connections(all connections should be ready and should work)
  run steampipe query 'select * from "M_t0".chaos_all_column_types'
  assert_success
  run steampipe query 'select * from "M_t1".chaos_all_column_types'
  assert_success
  run steampipe query 'select * from "M_t2".chaos_all_column_types'
  assert_success
  run steampipe query 'select * from "M_t3".chaos_all_column_types'
  assert_success
  run steampipe query 'select * from "M_t4".chaos_all_column_types'
  assert_success
  run steampipe query 'select * from "M_t5".chaos_all_column_types'
  assert_success
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}

function teardown() {
  # remove the files created as part of these tests 
  rm -f $STEAMPIPE_INSTALL_DIR/config/chaos.spc
}
