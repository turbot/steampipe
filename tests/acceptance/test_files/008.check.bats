load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "check check_rendering_benchmark" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check benchmark.control_check_rendering_benchmark
  assert_equal $status 12
  cd -
}

@test "check long control title" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.control_long_title --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_long_title.txt)"
  cd -
}

@test "check short control title" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.control_short_title --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_short_title.txt)"
  cd -
}

@test "check unicode control title" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.control_unicode_title --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_unicode_title.txt)"
  cd -
}

@test "check reasons(very long, very short, unicode)" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.control_long_short_unicode_reasons --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_reasons.txt)"
  cd -
}

@test "check control with all possible statuses(10 OK, 5 ALARM, 2 ERROR, 1 SKIP and 3 INFO)" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_mixed_results_1 --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_mixed_results.txt)"
  cd -
}

@test "check control with all resources in ALARM" {
  cd $CONTROL_RENDERING_TEST_MOD
  run steampipe check control.sample_control_all_alarms --progress=false --theme=plain
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_all_alarm.txt)"
  cd -
}

# @test "steampipe check cis_v130 - output csv - no header" {
#   cd $WORKSPACE_DIR
#   run steampipe check benchmark.cis_v130 --output=csv --progress=false --header=false
#   assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_csv_noheader.csv)"
#   cd -
# }

# @test "steampipe check cis_v130 - output json" {
#   cd $WORKSPACE_DIR
#   run steampipe check benchmark.cis_v130 --output=json --progress=false
#   assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_json.json)"
#   cd -
# }

# @test "steampipe check cis_v130 - export csv" {
#   cd $WORKSPACE_DIR
#   run steampipe check benchmark.cis_v130 --export=csv:./test.csv --progress=false
#   assert_equal "$(cat ./test.csv)" "$(cat $TEST_DATA_DIR/expected_check_csv.csv)"
#   rm -f ./test.csv
#   cd -
# }

# @test "steampipe check cis_v130 - export json" {
#   cd $WORKSPACE_DIR
#   run steampipe check benchmark.cis_v130 --export=json:./test.json --progress=false
#   assert_equal "$(cat ./test.json)" "$(cat $TEST_DATA_DIR/expected_check_json.json)"
#   rm -f ./test.json
#   cd -
# }

# @test "steampipe check cis_v130 - export html" {
#   tmpdir=$(mktemp -d)
#   cp $TEST_DATA_DIR/expected_check_html.html $tmpdir/expected.html

#   cd $WORKSPACE_DIR
#   run steampipe check benchmark.cis_v130 --export=html:$tmpdir/test.html --progress=false

#   declare -a remove_lines=("4587" "4588" "4589")
  
#   # checking for OS type, since sed command is different for linux and OSX
#   for remove_line in "${remove_lines[@]}"
#   do
#     if [[ "$OSTYPE" == "darwin"* ]]; then
#       run sed -i ".html" "${remove_line}d" $tmpdir/test.html
#       run sed -i ".html" "${remove_line}d" $tmpdir/expected.html
#     else
#       run sed -i "${remove_line}d" $tmpdir/test.html
#       run sed -i "${remove_line}d" $tmpdir/expected.html
#     fi
#   done

#   assert_equal "$(cat $tmpdir/test.html)" "$(cat $tmpdir/expected.html)"
#   rm -rf $tmpdir
#   cd -
# }

# @test "steampipe check cis_v130 - export markdown" {
#   tmpdir=$(mktemp -d)
#   cp $TEST_DATA_DIR/expected_check_markdown.md $tmpdir/expected.md

#   cd $WORKSPACE_DIR
#   run steampipe check benchmark.cis_v130 --export=markdown:$tmpdir/test.md --progress=false

#   # checking for OS type, since sed command is different for linux and OSX
#   if [[ "$OSTYPE" == "darwin"* ]]; then
#     run sed -i '.md' '834d' $tmpdir/test.md
#     run sed -i '.md' '834d' $tmpdir/expected.md
#   else
#     run sed -i '834d' $tmpdir/test.md
#     run sed -i '834d' $tmpdir/expected.md
#   fi

#   assert_equal "$(cat $tmpdir/test.md)" "$(cat $tmpdir/expected.md)"
#   rm -rf $tmpdir
#   cd -
# }
