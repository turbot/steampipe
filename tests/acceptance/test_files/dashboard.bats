load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "simple dashboard test" {
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
  # run a dashboard and shapshot the output
  run steampipe dashboard dashboard.testing_with_blocks --export test.sps --mod-location "$FILE_PATH/test_data/dashboard_withs"

  # get the patch diff between the two snapshots
  run jd -f patch $SNAPSHOTS_DIR/expected_sps_many_withs_dashboard.json test.sps

  # run the script to evaluate the patch
  # returns nothing if there is no diff(except start_time, end_time & search_path)
  diff=$($FILE_PATH/test_files/json_patch.sh $output)
  echo $diff
  rm -f test.sps

  # check if there is no diff returned by the script
  assert_equal "$diff" ""
}

@test "dashboard with 'text' blocks" {
  # run a dashboard and shapshot the output
  run steampipe dashboard dashboard.testing_text_blocks --export test.sps --mod-location "$FILE_PATH/test_data/dashboard_texts"

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
  # run a dashboard and shapshot the output
  run steampipe dashboard dashboard.testing_card_blocks --export test.sps --mod-location "$FILE_PATH/test_data/dashboard_cards"

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