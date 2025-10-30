load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "check postgres database, fdw are correctly installed" {
  # create a fresh target install dir
  target_install_directory=$(mktemp -d)

  # running steampipe - this would install the postgres database and the FDW from the registry
  steampipe query "select 1" --install-dir $target_install_directory

  # check postgres binary is present in correct location
  run file $target_install_directory/db/14.19.0/postgres/bin/postgres
  if [[ "$arch" == "x86_64" && "$os" == "Darwin" ]]; then
    assert_output --partial 'Mach-O 64-bit executable x86_64'
  elif [[ "$arch" == "arm64" && "$os" == "Darwin" ]]; then
    assert_output --partial 'Mach-O 64-bit executable arm64'
  elif [[ "$arch" == "x86_64" && "$os" == "Linux" ]]; then
    assert_output --partial 'ELF 64-bit LSB pie executable, x86-64'
  elif [[ "$arch" == "aarch64" && "$os" == "Linux" ]]; then
    assert_output --partial 'ELF 64-bit LSB executable, ARM aarch64'
  fi

  # check initdb binary is present in the correct location
  run file $target_install_directory/db/14.19.0/postgres/bin/initdb
  if [[ "$arch" == "arm64" && "$os" == "Darwin" ]]; then
    assert_output --partial 'Mach-O 64-bit executable arm64'
  elif [[ "$arch" == "x86_64" && "$os" == "Darwin" ]]; then
    assert_output --partial 'Mach-O 64-bit executable x86_64'
  elif [[ "$arch" == "x86_64" && "$os" == "Linux" ]]; then
    assert_output --partial 'ELF 64-bit LSB pie executable, x86-64'
  elif [[ "$arch" == "aarch64" && "$os" == "Linux" ]]; then
    assert_output --partial 'ELF 64-bit LSB executable, ARM aarch64'
  fi

  # check fdw binary(steampipe_postgres_fdw.so) is present in the correct location
  run file $target_install_directory/db/14.19.0/postgres/lib/postgresql/steampipe_postgres_fdw.so
  if [[ "$arch" == "arm64" && "$os" == "Darwin" ]]; then
    assert_output --partial 'Mach-O 64-bit bundle arm64'
  elif [[ "$arch" == "x86_64" && "$os" == "Darwin" ]]; then
    assert_output --partial 'Mach-O 64-bit bundle x86_64'
  elif [[ "$arch" == "x86_64" && "$os" == "Linux" ]]; then
    assert_output --partial 'ELF 64-bit LSB shared object, x86-64'
  elif [[ "$arch" == "aarch64" && "$os" == "Linux" ]]; then
    assert_output --partial 'ELF 64-bit LSB shared object, ARM aarch64'
  fi

  # check fdw extension(steampipe_postgres_fdw.control) is present in the correct location
  run file $target_install_directory/db/14.19.0/postgres/share/postgresql/extension/steampipe_postgres_fdw.control
  assert_output --partial 'ASCII text'
}

@test "check plugin is correctly installed" {
  # create a fresh target install dir
  target_install_directory=$(mktemp -d)

  # running steampipe - this would install the postgres database and the FDW from the registry
  steampipe query "select 1" --install-dir $target_install_directory

  # install a plugin
  steampipe plugin install chaos --install-dir $target_install_directory --progress=false

  # check plugin binary is present in correct location
  run file $target_install_directory/plugins/hub.steampipe.io/plugins/turbot/chaos@latest/steampipe-plugin-chaos.plugin
  if [[ "$arch" == "arm64" && "$os" == "Darwin" ]]; then
    assert_output --partial 'Mach-O 64-bit executable arm64'
  elif [[ "$arch" == "x86_64" && "$os" == "Darwin" ]]; then
    assert_output --partial 'Mach-O 64-bit executable x86_64'
  elif [[ "$arch" == "x86_64" && "$os" == "Linux" ]]; then
    assert_output --partial 'ELF 64-bit LSB executable, x86-64'
  elif [[ "$arch" == "aarch64" && "$os" == "Linux" ]]; then
    assert_output --partial 'ELF 64-bit LSB executable, ARM aarch64'
  fi

  # check spc config file is present in correct location
  run file $target_install_directory/config/chaos.spc
  assert_output --partial 'ASCII text'
}

function setup() {
  arch=$(uname -m)
  os=$(uname -s)
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
