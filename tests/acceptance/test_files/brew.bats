load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# Homebrew-core runs a set of tests in their release workflows. These tests replicate the 
# tests that they run on steampipe. This is to make sure that there are no unknown failures
# in their workflows

@test "steampipe completion should not create INSTALL DIRs" {
  # create a fresh target install dir
  target_install_directory=$(mktemp -d)

  run steampipe completion zsh --install-dir $target_install_directory

  # check no steampipe install directories are created at target_install_directory
  directory_count=$(ls | wc -l)
  echo $directory_count

  # steampipe completion should not create INSTALL DIRs
  assert_equal $directory_count 0
}
