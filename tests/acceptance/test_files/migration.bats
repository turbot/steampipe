load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

## public schema migration

@test "verify data is properly migrated when upgrading from v1.0.3" {
  # setup sql statements
  setup_sql[0]="create table sample(sample_col_1 char(10), sample_col_2 char(10))"
  setup_sql[1]="insert into sample(sample_col_1,sample_col_2) values ('foo','bar')"
  setup_sql[2]="insert into sample(sample_col_1,sample_col_2) values ('foo1','bar1')"
  setup_sql[3]="create function sample_func() returns integer as 'select 1 as result;' language sql;"

  # verify sql statements
  verify_sql[0]="select * from sample"
  verify_sql[1]="select * from sample_func()"

  # create a temp directory to install steampipe(1.0.3)
  tmpdir="$(mktemp -d)"
  mkdir -p "${tmpdir}"
  tmpdir="${tmpdir%/}"
    
  # find the name of the zip file as per OS and arch
  case $(uname -sm) in
	"Darwin x86_64") target="darwin_amd64.zip" ;;
	"Darwin arm64") target="darwin_arm64.zip" ;;
	"Linux x86_64") target="linux_amd64.tar.gz" ;;
	"Linux aarch64") target="linux_arm64.tar.gz" ;;
	*) echo "Error: '$(uname -sm)' is not supported yet." 1>&2;exit 1 ;;
	esac
    
  # download the zip and extract
  steampipe_uri="https://github.com/turbot/steampipe/releases/download/v1.0.3/steampipe_${target}"
  case $(uname -s) in
    "Darwin") zip_location="${tmpdir}/steampipe.zip" ;;
    "Linux") zip_location="${tmpdir}/steampipe.tar.gz" ;;
    *) echo "Error: steampipe is not supported on '$(uname -s)' yet." 1>&2;exit 1 ;;
  esac
  curl --fail --location --progress-bar --output "$zip_location" "$steampipe_uri"
  tar -xf "$zip_location" -C "$tmpdir"
  
  # start the service
  $tmpdir/steampipe --install-dir $tmpdir service start

  # execute the setup sql statements
  for ((i = 0; i < ${#setup_sql[@]}; i++)); do
    $tmpdir/steampipe --install-dir $tmpdir query "${setup_sql[$i]}"
  done

  # store the result of the verification statements(1.0.3)
  for ((i = 0; i < ${#verify_sql[@]}; i++)); do
    $tmpdir/steampipe --install-dir $tmpdir query "${verify_sql[$i]}" > verify$i.txt
  done

  # stop the service
  $tmpdir/steampipe --install-dir $tmpdir service stop
  
  # Now run this version - which should migrate the data
  steampipe --install-dir $tmpdir service start
  
  # store the result of the verification statements(0.14.*)
  for ((i = 0; i < ${#verify_sql[@]}; i++)); do
    echo "VerifySQL: ${verify_sql[$i]}"
    steampipe --install-dir $tmpdir query "${verify_sql[$i]}" > verify$i$i.txt
  done

  # stop the service
  steampipe --install-dir $tmpdir service stop

  # verify data is migrated correctly
  for ((i = 0; i < ${#verify_sql[@]}; i++)); do
    assert_equal "$(cat verify$i.txt)" "$(cat verify$i$i.txt)"
  done

  rm -rf $tmpdir
  rm -f verify*
}

@test "verify data is properly migrated when upgrading from v2.2.0" {
  # setup sql statements
  setup_sql[0]="create table sample(sample_col_1 char(10), sample_col_2 char(10))"
  setup_sql[1]="insert into sample(sample_col_1,sample_col_2) values ('foo','bar')"
  setup_sql[2]="insert into sample(sample_col_1,sample_col_2) values ('foo1','bar1')"
  setup_sql[3]="create function sample_func() returns integer as 'select 1 as result;' language sql;"

  # verify sql statements
  verify_sql[0]="select * from sample"
  verify_sql[1]="select * from sample_func()"

  # create a temp directory to install steampipe(2.2.0)
  tmpdir="$(mktemp -d)"
  mkdir -p "${tmpdir}"
  tmpdir="${tmpdir%/}"
    
  # find the name of the zip file as per OS and arch
  case $(uname -sm) in
	"Darwin x86_64") target="darwin_amd64.zip" ;;
	"Darwin arm64") target="darwin_arm64.zip" ;;
	"Linux x86_64") target="linux_amd64.tar.gz" ;;
	"Linux aarch64") target="linux_arm64.tar.gz" ;;
	*) echo "Error: '$(uname -sm)' is not supported yet." 1>&2;exit 1 ;;
	esac
    
  # download the zip and extract
  steampipe_uri="https://github.com/turbot/steampipe/releases/download/v2.2.0/steampipe_${target}"
  case $(uname -s) in
    "Darwin") zip_location="${tmpdir}/steampipe.zip" ;;
    "Linux") zip_location="${tmpdir}/steampipe.tar.gz" ;;
    *) echo "Error: steampipe is not supported on '$(uname -s)' yet." 1>&2;exit 1 ;;
  esac
  curl --fail --location --progress-bar --output "$zip_location" "$steampipe_uri"
  tar -xf "$zip_location" -C "$tmpdir"
  
  # start the service
  $tmpdir/steampipe --install-dir $tmpdir service start

  # execute the setup sql statements
  for ((i = 0; i < ${#setup_sql[@]}; i++)); do
    $tmpdir/steampipe --install-dir $tmpdir query "${setup_sql[$i]}"
  done

  # store the result of the verification statements(1.0.3)
  for ((i = 0; i < ${#verify_sql[@]}; i++)); do
    $tmpdir/steampipe --install-dir $tmpdir query "${verify_sql[$i]}" > verify$i.txt
  done

  # stop the service
  $tmpdir/steampipe --install-dir $tmpdir service stop
  
  # Now run this version - which should migrate the data
  steampipe --install-dir $tmpdir service start
  
  # store the result of the verification statements(0.14.*)
  for ((i = 0; i < ${#verify_sql[@]}; i++)); do
    echo "VerifySQL: ${verify_sql[$i]}"
    steampipe --install-dir $tmpdir query "${verify_sql[$i]}" > verify$i$i.txt
  done

  # stop the service
  steampipe --install-dir $tmpdir service stop

  # verify data is migrated correctly
  for ((i = 0; i < ${#verify_sql[@]}; i++)); do
    assert_equal "$(cat verify$i.txt)" "$(cat verify$i$i.txt)"
  done

  rm -rf $tmpdir
  rm -f verify*
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}

function setup() {
  # skip if this test is run on Linux ARM64, since there is no linux_arm binary available
  # for v0.13.6 to run this test
  sys=$(uname -sm)
  if [[ "$sys" == "Linux aarch64" ]]; then
    skip
  else
    echo "Running migration test..."
  fi
}
