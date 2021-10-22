load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "steampipe check cis_v130" {
  cd $WORKSPACE_DIR
  run steampipe check benchmark.cis_v130
  assert_equal $status 10
  cd -
}

@test "steampipe check cis_v130 - output csv" {
  cd $WORKSPACE_DIR
  run steampipe check benchmark.cis_v130 --output=csv --progress=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_csv.csv)"
  cd -
}

@test "steampipe check cis_v130 - output csv - | separator" {
  cd $WORKSPACE_DIR
  run steampipe check benchmark.cis_v130 --output=csv --progress=false "--separator=|"
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_separator_csv.csv)"
  cd -
}

@test "steampipe check cis_v130 - output csv - no header" {
  cd $WORKSPACE_DIR
  run steampipe check benchmark.cis_v130 --output=csv --progress=false --header=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_csv_noheader.csv)"
  cd -
}

@test "steampipe check cis_v130 - output json" {
  cd $WORKSPACE_DIR
  run steampipe check benchmark.cis_v130 --output=json --progress=false
  assert_equal "$output" "$(cat $TEST_DATA_DIR/expected_check_json.json)"
  cd -
}

@test "steampipe check cis_v130 - export csv" {
  cd $WORKSPACE_DIR
  run steampipe check benchmark.cis_v130 --export=csv:./test.csv --progress=false
  assert_equal "$(cat ./test.csv)" "$(cat $TEST_DATA_DIR/expected_check_csv.csv)"
  rm -f ./test.csv
  cd -
}

@test "steampipe check cis_v130 - export json" {
  cd $WORKSPACE_DIR
  run steampipe check benchmark.cis_v130 --export=json:./test.json --progress=false
  assert_equal "$(cat ./test.json)" "$(cat $TEST_DATA_DIR/expected_check_json.json)"
  rm -f ./test.json
  cd -
}

# @test "steampipe check cis_v130 - export html" {
#   tmpdir=$(mktemp -d)
#   cp $TEST_DATA_DIR/expected_check_html.html $tmpdir/expected.html

#   cd $WORKSPACE_DIR
#   run steampipe check benchmark.cis_v130 --export=html:$tmpdir/test.html --progress=false

#   # checking for OS type, since sed command is different for linux and OSX
#   if [[ "$OSTYPE" == "darwin"* ]]; then
#     run sed -i '.html' '4478d' $tmpdir/test.html
#     run sed -i '.html' '4478d' $tmpdir/expected.html
#   else
#     run sed -i '4478d' $tmpdir/test.html
#     run sed -i '4478d' $tmpdir/expected.html
#   fi
  
  
#   assert_equal "$(cat $tmpdir/test.html)" "$(cat $tmpdir/expected.html)"
#   rm -rf $tmpdir
#   cd -
# }

@test "steampipe check cis_v130 - export markdown" {
  tmpdir=$(mktemp -d)
  cp $TEST_DATA_DIR/expected_check_markdown.md $tmpdir/expected.md

  cd $WORKSPACE_DIR
  run steampipe check benchmark.cis_v130 --export=markdown:$tmpdir/test.md --progress=false

  # checking for OS type, since sed command is different for linux and OSX
  if [[ "$OSTYPE" == "darwin"* ]]; then
    run sed -i '.md' '834d' $tmpdir/test.md
    run sed -i '.md' '834d' $tmpdir/expected.md
  else
    run sed -i '834d' $tmpdir/test.md
    run sed -i '834d' $tmpdir/expected.md
  fi
  
  
  assert_equal "$(cat $tmpdir/test.md)" "$(cat $tmpdir/expected.md)"
  rm -rf $tmpdir
  cd -
}
