load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"
load ../helpers/steampipe_processes

@test "check postgres database, fdw are correctly installed" {
  # create a fresh target install dir
  target_install_directory="$BATS_TEST_TMPDIR/install"
  mkdir -p "$target_install_directory"

  # running steampipe - this would install the postgres database and the FDW from the registry
  steampipe query "select 1" --install-dir $target_install_directory

  # check postgres binary is present in correct location
  run file $target_install_directory/db/14.24.0/postgres/bin/postgres
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
  run file $target_install_directory/db/14.24.0/postgres/bin/initdb
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
  run file $target_install_directory/db/14.24.0/postgres/lib/postgresql/steampipe_postgres_fdw.so
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
  run file $target_install_directory/db/14.24.0/postgres/share/postgresql/extension/steampipe_postgres_fdw.control
  assert_output --partial 'ASCII text'
}

@test "check plugin is correctly installed" {
  # create a fresh target install dir
  target_install_directory="$BATS_TEST_TMPDIR/install"
  mkdir -p "$target_install_directory"

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

@test "reinstalling the programs of the current version keeps its database" {
  target_install_directory="$BATS_TEST_TMPDIR/install"
  mkdir -p "$target_install_directory"

  # install the database, then put data in it
  steampipe query "create table keep_me(c text)" --install-dir $target_install_directory
  steampipe query "insert into keep_me values ('still here')" --install-dir $target_install_directory

  # lose the programs, but not the data folder
  rm -rf $target_install_directory/db/*/postgres
  [ -f $target_install_directory/db/*/data/PG_VERSION ]

  # starting again reinstalls the programs and uses the existing database
  run steampipe query "select c from keep_me" --output csv --install-dir $target_install_directory
  assert_success
  assert_output --partial 'still here'
}

@test "a leftover backup is not restored into the kept database" {
  target_install_directory="$BATS_TEST_TMPDIR/install"
  mkdir -p "$target_install_directory"

  # a running service lets us take a real dump holding a table that the kept database will not have
  steampipe service start --install-dir $target_install_directory
  steampipe query "create table stale_t(c text)" --install-dir $target_install_directory
  $(ls $target_install_directory/db/*/postgres/bin/pg_dump) --file=$target_install_directory/db/backup.bk --format=custom --schema=public --dbname=steampipe --host=$(ls -d $target_install_directory/db/*/data) --port=9193 --username=root
  steampipe query "drop table stale_t" --install-dir $target_install_directory
  steampipe query "create table keep_me(c text)" --install-dir $target_install_directory
  steampipe query "insert into keep_me values ('still here')" --install-dir $target_install_directory
  steampipe service stop --install-dir $target_install_directory

  # lose the programs, but not the data folder or the dump
  rm -rf $target_install_directory/db/*/postgres
  [ -f $target_install_directory/db/*/data/PG_VERSION ]
  [ -f $target_install_directory/db/backup.bk ]

  run steampipe query "select c from keep_me" --output csv --install-dir $target_install_directory
  assert_success
  assert_output --partial 'still here'

  run steampipe query "select * from stale_t" --install-dir $target_install_directory
  assert_failure
  assert_output --partial 'stale_t'

  [ ! -f $target_install_directory/db/backup.bk ]
  [ -n "$(ls $target_install_directory/backups/database-*.dump)" ]
}

@test "an upgrade refused because the data folder holds a database dumps nothing that is restored later" {
  target_install_directory="$BATS_TEST_TMPDIR/install"
  mkdir -p "$target_install_directory"

  steampipe service start --install-dir $target_install_directory
  steampipe query "create table old_t(c text)" --install-dir $target_install_directory
  steampipe service stop --install-dir $target_install_directory

  # an older install of a different major version, holding the database with old_t
  current_install=$(ls -d $target_install_directory/db/*/)
  cp -R $current_install $target_install_directory/db/13.0.0

  # this version's data folder holds a database of its own
  steampipe service start --install-dir $target_install_directory
  steampipe query "drop table old_t" --install-dir $target_install_directory
  steampipe query "create table keep_me(c text)" --install-dir $target_install_directory
  steampipe query "insert into keep_me values ('still here')" --install-dir $target_install_directory
  steampipe service stop --install-dir $target_install_directory
  rm -rf $current_install/postgres

  run steampipe query "select 1" --install-dir $target_install_directory
  assert_failure
  assert_output --partial 'already holds a database'
  [ ! -f $target_install_directory/db/backup.bk ]

  rm -rf $target_install_directory/db/13.0.0

  run steampipe query "select c from keep_me" --output csv --install-dir $target_install_directory
  assert_success
  assert_output --partial 'still here'

  run steampipe query "select * from old_t" --install-dir $target_install_directory
  assert_failure
  assert_output --partial 'old_t'

  [ ! -f $target_install_directory/db/backup.bk ]
}

@test "a leftover backup is not restored into a database moved from an older install of the same major version" {
  target_install_directory="$BATS_TEST_TMPDIR/install"
  mkdir -p "$target_install_directory"

  steampipe service start --install-dir $target_install_directory
  steampipe query "create table keep_me(c text)" --install-dir $target_install_directory
  steampipe query "insert into keep_me values ('still here')" --install-dir $target_install_directory

  # a dump of the database as it is now: restoring it into this database would clash with its tables
  $(ls $target_install_directory/db/*/postgres/bin/pg_dump) --file=$target_install_directory/db/backup.bk --format=custom --schema=public --dbname=steampipe --host=$(ls -d $target_install_directory/db/*/data) --port=9193 --username=root
  steampipe service stop --install-dir $target_install_directory
  [ -f $target_install_directory/db/backup.bk ]

  # the database now lives in an older install of the same major version
  current_install=$(ls -d $target_install_directory/db/*/)
  current_major=$(basename $current_install | cut -d. -f1)
  mv $current_install $target_install_directory/db/$current_major.0.1

  run steampipe query "select c from keep_me" --output csv --install-dir $target_install_directory
  assert_success
  assert_output --partial 'still here'

  [ ! -f $target_install_directory/db/backup.bk ]
  [ -n "$(ls $target_install_directory/backups/database-*.dump)" ]
  [ ! -d $target_install_directory/db/$current_major.0.1 ]
}

function setup() {
  arch=$(uname -m)
  os=$(uname -s)
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(count_steampipe_processes)
  assert_equal $num 0
}
