load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "simple dashboard test" {
  # run a dashboard and shapshot the output
  run steampipe dashboard dashboard.sibling_containers_report --export test.sps --output none --mod-location "$FILE_PATH/test_data/dashboard_sibling_containers"
  ls
  # rename the snapshot file into a json file, for ease of comparison
  mv test.sps actual_sps_sibling_containers_report.json

  # get the patch diff between the two snapshots
  run jd -f patch $SNAPSHOTS_DIR/expected_sps_sibling_containers_report.json actual_sps_sibling_containers_report.json

  # run the script to evaluate the patch
  # returns nothing if there is no diff(except start_time, end_time & search_path)
  diff=$($FILE_PATH/test_files/json_patch.sh $output)
  echo $diff
  rm -f actual_sps_sibling_containers_report.json

  # check if there is no diff returned by the script
  assert_equal "$diff" ""
}