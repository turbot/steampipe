load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "simple dashboard test" {
  export STEAMPIPE_LOG=info
  env
  # run a dashboard and shapshot the output
  run steampipe dashboard dashboard.sibling_containers_report --export test.sps --output none --mod-location "$FILE_PATH/test_data/dashboard_sibling_containers"

  # get the patch diff between the two snapshots
  run jd -f patch $SNAPSHOTS_DIR/expected_sps_sibling_containers_report.json test.sps

  # run the script to evaluate the patch
  # returns nothing if there is no diff(except start_time, end_time & search_path)
  diff=$($FILE_PATH/test_files/json_patch.sh $output)
  echo $diff
  rm -f test.sps

  # check if there is no diff returned by the script
  assert_equal "$diff" ""
}

@test "dashboard with 'with' blocks" {
  export STEAMPIPE_LOG=info
  env
  # run a dashboard and shapshot the output
  run steampipe dashboard dashboard.testing_with_blocks --export test.sps --output none --mod-location "$FILE_PATH/test_data/dashboard_withs"

  # sort the panels data by 'name' using jq sort_by(for ease in comparison)
  cat test.sps | jq '.panels."dashbaord_withs.graph.with_testing".data.columns|=sort_by(.name)' > test2.json

  # get the patch diff between the two snapshots
  run jd -f patch $SNAPSHOTS_DIR/expected_sps_many_withs_dashboard.json test2.json

  # run the script to evaluate the patch
  # returns nothing if there is no diff(except start_time, end_time & search_path)
  diff=$($FILE_PATH/test_files/json_patch.sh $output)
  echo $diff
  rm -f test.sps
  rm -f test2.json

  # check if there is no diff returned by the script
  assert_equal "$diff" ""
}

@test "dashboard with 'text' blocks" {
  export STEAMPIPE_LOG=info
  env
  # run a dashboard and shapshot the output
  run steampipe dashboard dashboard.testing_text_blocks --export test.sps --output none --mod-location "$FILE_PATH/test_data/dashboard_texts"

  # get the patch diff between the two snapshots
  run jd -f patch $SNAPSHOTS_DIR/expected_sps_testing_text_blocks_dashboard.json test.sps

  # run the script to evaluate the patch
  # returns nothing if there is no diff(except start_time, end_time & search_path)
  diff=$($FILE_PATH/test_files/json_patch.sh $output)
  echo $diff
  rm -f test.sps

  # check if there is no diff returned by the script
  assert_equal "$diff" ""
}

@test "dashboard with 'card' blocks" {
  export STEAMPIPE_LOG=info
  env
  # run a dashboard and shapshot the output
  run steampipe dashboard dashboard.testing_card_blocks --export test.sps --output none --mod-location "$FILE_PATH/test_data/dashboard_cards"

  # get the patch diff between the two snapshots
  run jd -f patch $SNAPSHOTS_DIR/expected_sps_testing_card_blocks_dashboard.json test.sps


  # run the script to evaluate the patch
  # returns nothing if there is no diff(except start_time, end_time & search_path)
  diff=$($FILE_PATH/test_files/json_patch.sh $output)
  echo $diff
  rm -f test.sps

  # check if there is no diff returned by the script
  assert_equal "$diff" ""
}

@test "dashboard with node and edge blocks" {
  export STEAMPIPE_LOG=info
  env
  # run a dashboard and shapshot the output
  run steampipe dashboard dashboard.testing_nodes_and_edges --export test.sps --output none --mod-location "$FILE_PATH/test_data/dashboard_graphs"

  # sort the panels data by 'name' using jq sort_by(for ease in comparison)
  cat test.sps | jq '.panels."dashboard_graphs.graph.node_and_edge_testing".data.columns|=sort_by(.name)' > test2.json

  # get the patch diff between the two snapshots
  run jd -f patch $SNAPSHOTS_DIR/expected_sps_testing_nodes_and_edges_dashboard.json test2.json

  # run the script to evaluate the patch
  # returns nothing if there is no diff(except start_time, end_time & search_path)
  diff=$($FILE_PATH/test_files/json_patch.sh $output)
  echo $diff
  rm -f test.sps
  rm -f test2.json

  # check if there is no diff returned by the script
  assert_equal "$diff" ""
}

@test "dashboard with 'input' and test --dashboard-input arg" {
  export STEAMPIPE_LOG=info
  env
  # run a dashboard and shapshot the output
  run steampipe dashboard dashboard.testing_dashboard_inputs --export test.sps --output none --mod-location "$FILE_PATH/test_data/dashboard_inputs" --dashboard-input new_input=test

  # get the patch diff between the two snapshots
  run jd -f patch $SNAPSHOTS_DIR/expected_sps_testing_dashboard_inputs.json test.sps

  # run the script to evaluate the patch
  # returns nothing if there is no diff(except start_time, end_time & search_path)
  diff=$($FILE_PATH/test_files/json_patch.sh $output)
  echo $diff
  rm -f test.sps

  # check if there is no diff returned by the script
  assert_equal "$diff" ""
}

# run teardown with 30s sleep after each test since it takes some time to kill all plugins in pluginMultiConnectionMap
function setup() {
  # sleep 30
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
