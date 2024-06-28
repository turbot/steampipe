load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe check exitCode - no control alarms or errors" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check benchmark.all_controls_ok
  assert_equal $status 0
  cd -
}

@test "steampipe check exitCode - with controls in error" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check benchmark.control_check_rendering_benchmark
  assert_equal $status 2
  cd -
}

#@test "steampipe check exitCode - with controls in alarm" {
#  cd $FUNCTIONALITY_TEST_MOD
#  run steampipe check benchmark.check_search_path_benchmark
#  assert_equal $status 1
#  cd -
#}

@test "steampipe check exitCode - with controls in error(running multiple benchmarks together)" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check benchmark.control_summary_benchmark benchmark.check_cache_benchmark
  assert_equal $status 2
  cd -
}

@test "steampipe check exitCode - runtime error(insufficient args)" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check
  assert_equal $status 254
  cd -
}

@test "steampipe check long control title" {
  cd $CONTROL_RENDERING_TEST_MOD
  export STEAMPIPE_DISPLAY_WIDTH=100
  run steampipe check control.control_long_title --progress=false --theme=plain
  assert_output --partial "$(cat $TEST_DATA_DIR/expected_long_title.txt)"
  cd -
}

@test "steampipe check short control title" {
  cd $CONTROL_RENDERING_TEST_MOD
  export STEAMPIPE_DISPLAY_WIDTH=100
  run steampipe check control.control_short_title --progress=false --theme=plain
  assert_output --partial "$(cat $TEST_DATA_DIR/expected_short_title.txt)"
  cd -
}

@test "steampipe check unicode control title" {
  cd $CONTROL_RENDERING_TEST_MOD
  export STEAMPIPE_DISPLAY_WIDTH=100
  run steampipe check control.control_unicode_title --progress=false --theme=plain
  assert_output --partial "$(cat $TEST_DATA_DIR/expected_unicode_title.txt)"
  cd -
}

@test "steampipe check reasons(very long, very short, unicode)" {
  cd $CONTROL_RENDERING_TEST_MOD
  export STEAMPIPE_DISPLAY_WIDTH=100
  run steampipe check control.control_long_short_unicode_reasons --progress=false --theme=plain
  assert_output --partial "$(cat $TEST_DATA_DIR/expected_reasons.txt)"
  cd -
}

@test "steampipe check control with all possible statuses(10 OK, 5 ALARM, 2 ERROR, 1 SKIP and 3 INFO)" {
  cd $CONTROL_RENDERING_TEST_MOD
  export STEAMPIPE_DISPLAY_WIDTH=100
  run steampipe check control.sample_control_mixed_results_1 --progress=false --theme=plain
  assert_output --partial "$(cat $TEST_DATA_DIR/expected_mixed_results.txt)"
  cd -
}

@test "steampipe check control with all resources in ALARM" {
  cd $CONTROL_RENDERING_TEST_MOD
  export STEAMPIPE_DISPLAY_WIDTH=100
  run steampipe check control.sample_control_all_alarms --progress=false --theme=plain
  assert_output --partial "$(cat $TEST_DATA_DIR/expected_all_alarm.txt)"
  cd -
}

@test "steampipe check control with blank dimension" {
  cd $BLANK_DIMENSION_VALUE_TEST_MOD
  export STEAMPIPE_DISPLAY_WIDTH=100
  run steampipe check all --progress=false --theme=plain
  assert_output --partial "$(cat $TEST_DATA_DIR/expected_blank_dimension.txt)"
  cd -
}

@test "steampipe check - output csv - no header" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --output=csv --progress=false --header=false
  assert_output --partial "$(cat $TEST_DATA_DIR/expected_check_csv_noheader.csv)"
  cd -
}

@test "steampipe check - output csv(check tags and dimensions sorting)" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_sorted_tags_and_dimensions --output=csv --progress=false
  assert_output --partial "$(cat $TEST_DATA_DIR/expected_check_csv_sorted_tags.csv)"
  cd -
}

@test "steampipe check - output json" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --output json --progress=false --export output.json
  output=""
  run jd "$TEST_DATA_DIR/expected_check_json.json" output.json
  echo $output
  assert_success
  rm -f output.json
  cd -
}

@test "steampipe check - export csv" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export test.csv --progress=false
  assert_equal "$(cat test.csv)" "$(cat $TEST_DATA_DIR/expected_check_csv.csv)"
  rm -f test.csv
  cd -
}

@test "steampipe check - export csv - pipe separator" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export test.csv --separator="|" --progress=false
  assert_equal "$(cat test.csv)" "$(cat $TEST_DATA_DIR/expected_check_csv_pipe_separator.csv)"
  rm -f test.csv
  cd -
}

@test "steampipe check - export csv(check tags and dimensions sorting)" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_sorted_tags_and_dimensions --export test.csv --progress=false
  assert_equal "$(cat test.csv)" "$(cat $TEST_DATA_DIR/expected_check_csv_sorted_tags.csv)"
  rm -f test.csv
  cd -
}

@test "steampipe check - export json" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export test.json --progress=false
  output=""
  run jd "$TEST_DATA_DIR/expected_check_json.json" test.json
  echo $output
  assert_success
  rm -f output.json
  cd -
}

@test "steampipe check - export html" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export test.html --progress=false
  
  # checking for OS type, since sed command is different for linux and OSX
  # removing the 642nd line, since it contains file locations and timestamps
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".html" "642d" test.html
    run sed -i ".html" "642d" test.html
    run sed -i ".html" "642d" test.html
  else
    run sed -i "642d" test.html
    run sed -i "642d" test.html
    run sed -i "642d" test.html
  fi

  assert_equal "$(cat test.html)" "$(cat $TEST_DATA_DIR/expected_check_html.html)"
  rm -rf test.html*
  cd -
}

@test "steampipe check - export md" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export test.md --progress=false
  
  # checking for OS type, since sed command is different for linux and OSX
  # removing the 42nd line, since it contains file locations and timestamps
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".md" "42d" test.md
  else
    run sed -i "42d" test.md
  fi

  assert_equal "$(cat test.md)" "$(cat $TEST_DATA_DIR/expected_check_markdown.md)"
  rm -rf test.md*
  cd -
}

@test "steampipe check - export nunit3" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export test.xml --progress=false

  # checking for OS type, since sed command is different for linux and OSX
  # removing the 6th line, since it contains duration, and duration will be different in each run
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i ".xml" "6d" test.xml
  else
    run sed -i "6d" test.xml
  fi

  assert_equal "$(cat test.xml)" "$(cat $TEST_DATA_DIR/expected_check_nunit3.xml)"
  rm -f test.xml*
  cd -
}

@test "steampipe check - export snapshot" {
  skip "deprecated"
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --export test.sps --progress=false

  # get the patch diff between the two snapshots
  run jd -f patch $TEST_DATA_DIR/expected_check_snapshot.sps test.sps

  # run the script to evaluate the patch
  # returns nothing if there is no diff(except start_time, end_time & search_path)
  diff=$($FILE_PATH/json_patch.sh $output)
  echo $diff
  rm -f test.sps

  # check if there is no diff returned by the script
  assert_equal "$diff" ""
  cd -
}

@test "steampipe check all" {
  cd $CHECK_ALL_MOD
  run steampipe check all --export test.json --progress=false
  output=""
  run jd "$TEST_DATA_DIR/expected_check_all.json" test.json
  echo $output
  assert_success
  rm -f output.json
  cd -
}

@test "check all - export sps" {
  skip "not implemented yet"
}
@test "check all - export file.sps" {
  skip "not implemented yet"
}
@test "check single benchmark - export sps" {
  skip "not implemented yet"
}
@test "check single benchmark - export file.sps" {
  skip "not implemented yet"
}
@test "check multiple benchmark - export sps" {
  skip "not implemented yet"
}
@test "check multiple benchmark - export file.sps" {
  skip "not implemented yet"
}

@test "check all - export json" {
  skip "not implemented yet"
}
@test "check all - export file.json" {
  skip "not implemented yet"
}
@test "check single benchmark - export json" {
  skip "not implemented yet"
}
@test "check single benchmark - export file.json" {
  skip "not implemented yet"
}
@test "check multiple benchmark - export json" {
  skip "not implemented yet"
}
@test "check multiple benchmark - export file.json" {
  skip "not implemented yet"
}

@test "check all - export csv" {
  skip "not implemented yet"
}
@test "check all - export file.csv" {
  skip "not implemented yet"
}
@test "check single benchmark - export csv" {
  skip "not implemented yet"
}
@test "check single benchmark - export file.csv" {
  skip "not implemented yet"
}
@test "check multiple benchmark - export csv" {
  skip "not implemented yet"
}
@test "check multiple benchmark - export file.csv" {
  skip "not implemented yet"
}

## check search_path tests

#@test "steampipe check search_path_prefix when passed through command line" {
#  cd $FUNCTIONALITY_TEST_MOD
#  run steampipe check control.search_path_test_1 --output json --search-path-prefix aws --export test.json
#  assert_equal "$(cat test.json | jq '.controls[0].results[0].status')" '"ok"'
#  rm -f test.json
#}

#@test "steampipe check search_path when passed through command line" {
#  cd $FUNCTIONALITY_TEST_MOD
#  run steampipe check control.search_path_test_2 --output json --search-path chaos,b,c --export test.json
#  assert_equal "$(cat test.json | jq '.controls[0].results[0].status')" '"ok"'
#  rm -f test.json
#}

#@test "steampipe check search_path and search_path_prefix when passed through command line" {
#  cd $FUNCTIONALITY_TEST_MOD
#  run steampipe check control.search_path_test_3 --output json --search-path chaos,b,c --search-path-prefix aws --export test.json
#  assert_equal "$(cat test.json | jq '.controls[0].results[0].status')" '"ok"'
#  rm -f test.json
#}



## plugin crash

@test "check whether the plugin is crashing or not" {
  skip
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check benchmark.check_plugin_crash_benchmark
  echo $output
  [ $(echo $output | grep "ERROR: context canceled" | wc -l | tr -d ' ') -eq 0 ]
}

# testing the check summary output feature in steampipe
@test "check summary output" {
  cd $FUNCTIONALITY_TEST_MOD
  run steampipe check benchmark.control_summary_benchmark --theme plain

  echo $output

  # TODO: Find a way to store the output in a file and match it with the 
  # expected file. For now the work-around is to check whether the output
  # contains `summary`
  assert_output --partial 'Summary'
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
