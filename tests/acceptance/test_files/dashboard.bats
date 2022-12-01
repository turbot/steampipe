load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "test 01" {
  run jd -f patch $SRC_DATA_DIR/snapshots/source.json $SRC_DATA_DIR/snapshots/target.json

  a=$(./tests/acceptance/test_files/test_script.sh $output)
  echo $a

  if [[ $a == "" ]]; then
    flag=1
  else
    flag=0
  fi

  assert_equal "$flag" "1"
}