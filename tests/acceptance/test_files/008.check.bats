load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe check check_rendering_benchmark" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check benchmark.control_check_rendering_benchmark
  assert_equal $status 27
  cd -
}

@test "steampipe check long control title" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.control_long_title --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_long_title.txt)"
  cd -
}

@test "steampipe check short control title" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.control_short_title --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_short_title.txt)"
  cd -
}

@test "steampipe check unicode control title" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.control_unicode_title --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_unicode_title.txt)"
  cd -
}

@test "steampipe check reasons(very long, very short, unicode)" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.control_long_short_unicode_reasons --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_reasons.txt)"
  cd -
}

@test "steampipe check control with all possible statuses(10 OK, 5 ALARM, 2 ERROR, 1 SKIP and 3 INFO)" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_mixed_results.txt)"
  cd -
}

@test "steampipe check control with all resources in ALARM" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_all_alarms --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_all_alarm.txt)"
  cd -
}

@test "steampipe check - output csv - no header" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --output=csv --progress=false --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_csv_noheader.csv)"
  cd -
}

@test "steampipe check - output json" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --output=json --progress=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_json.json)"
  cd -
}

@test "steampipe check - export csv" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export=csv:./test.csv --progress=false
  assert_equal "$(cat ./test.csv)" "$(cat $TEST_DATA_DIR/expected_check_csv.csv)"
  rm -f ./test.csv
  cd -
}

@test "steampipe check - export json" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export=json:./test.json --progress=false
  assert_equal "$(cat ./test.json)" "$(cat $TEST_DATA_DIR/expected_check_json.json)"
  rm -f ./test.json
  cd -
}

@test "steampipe check - export html" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export=html:./test.html --progress=false
  
  # checking for OS type, since sed command is different for linux and OSX
  # removing the 641st line, since it contains file locations and timestamps
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".html" "641d" ./test.html
    run sed -i ".html" "641d" ./test.html
    run sed -i ".html" "641d" ./test.html
  else
    run sed -i "641d" ./test.html
    run sed -i "641d" ./test.html
    run sed -i "641d" ./test.html
  fi

  assert_equal "$(cat ./test.html)" "$(cat $TEST_DATA_DIR/expected_check_html.html)"
  rm -rf ./test.html*
  cd -
}

@test "steampipe check - export md" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export=md:./test.md --progress=false
  
  # checking for OS type, since sed command is different for linux and OSX
  # removing the 41st line, since it contains file locations and timestamps
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".md" "41d" ./test.md
  else
    run sed -i "41d" ./test.md
  fi

  assert_equal "$(cat ./test.md)" "$(cat $TEST_DATA_DIR/expected_check_markdown.md)"
  rm -rf ./test.md*
  cd -
}
