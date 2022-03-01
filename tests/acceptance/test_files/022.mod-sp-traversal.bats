load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# This test consists of a mod with nested folders, with mod.sp file within one of them(folder11).
# Running steampipe check from folder111 should give us the result since the mod.sp file is present somewhere
# up the directory tree
@test "load a mod from an arbitrarily nested sub folder - PASS" {
  # go to the nested sub directory within the mod
  cd $FILE_PATH/test_data/nested_mod/folder1/folder11/folder111

  run steampipe check all
  assert_success
  cd -
}

# This test consists of a mod with nested folders, with mod.sp file within one of them(folder11).
# Running steampipe check from folder1 should return an error, since the mod.sp file is present nowhere
# up the directory tree
@test "load a mod from an arbitrarily nested sub folder - FAIL" {
  # go to the nested sub directory within the mod
  cd $FILE_PATH/test_data/nested_mod/folder1

  run steampipe check all
  assert_equal "$output" "Error: this command requires a mod definition file - could not find in the current directory tree"
  cd -
}

# This test consists of a mod with nested folders, with no mod.sp file in any of them.
# Running steampipe check from folder11 should return an error, since the mod.sp file is present nowhere
# up the directory tree
# Running steampipe query from folder11 should give us the result since query is independent of mod.sp file.
@test "check and query from an arbitrarily nested sub folder - PASS & FAIL" {
  # go to the nested sub directory within the mod
  cd $FILE_PATH/test_data/nested_mod_no_mod_file/folder1/folder11

  run steampipe check all
  assert_equal "$output" "Error: this command requires a mod definition file - could not find in the current directory tree"

  run steampipe query control.check_1
  assert_success
  cd -
}
