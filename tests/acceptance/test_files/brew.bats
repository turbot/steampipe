load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# Homebrew-core runs a set of tests in their release workflows. These tests replicate the 
# tests that they run on steampipe. This is to make sure that there are no unknown failures
# in their workflows

@test "steampipe completion should not create INSTALL DIRs" {
  export STEAMPIPE_LOG=info
  # create a fresh target install dir
  target_install_directory=$(mktemp -d)

  run steampipe completion zsh --install-dir $target_install_directory

  # check no steampipe install directories are created at target_install_directory
  cd $target_install_directory
  directory_count=$(ls | wc -l)
  echo $directory_count

  # steampipe completion should not create INSTALL DIRs
  assert_equal $directory_count 0
}

# This is to test that the steampipe binary can be symlinked and still function correctly.
# This is important for Homebrew and other package managers that may symlink the binary.
# We had a failure in v2.0.0 where the symlinked binary left over steampipe plugin processes
# running in the background, due to a pluginmanager bug. 
# This test ensures that the symlinked binary works properly and does not leave any processes 
# running in the background.
@test "symlinked steampipe binary should work" {
  export STEAMPIPE_LOG=info
  # create a fresh target dir
  target_directory=$(mktemp -d)

  # create a symlink to the steampipe binary
  ln -s $(which steampipe) $target_directory/sp

  # add the target directory to PATH
  export PATH=$target_directory:$PATH

  # run a steampipe command to verify the symlink has been created correctly
  run $target_directory/sp --version
  assert_success

  # check if querying is successful
  run $target_directory/sp query "select * from chaos_all_column_types"
  assert_success
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
