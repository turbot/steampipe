load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "verify data is properly migrated when upgrading from v0.13.6" {
  #change 
  # create a temp directory to install steampipe(0.13.6)
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
  steampipe_uri="https://github.com/turbot/steampipe/releases/download/v0.13.6/steampipe_${target}"
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
    echo "executing setup_sql$i on old DB"
    $tmpdir/steampipe --install-dir $tmpdir query "${setup_sql[$i]}"
  done

  # store the result of the verification statements(0.13.6)
  for ((i = 0; i < ${#verify_sql[@]}; i++)); do
    echo "extracting verify_sql$i on old DB"
    $tmpdir/steampipe --install-dir $tmpdir query "${verify_sql[$i]}" --output json > verify$i.json
  done

  # stop the service
  $tmpdir/steampipe --install-dir $tmpdir service stop
  
  echo "installing new DB"
  # Now run this version(0.14.*) - which should migrate the data
  steampipe --install-dir $tmpdir service start
  
  # store the result of the verification statements(0.14.*)
  for ((i = 0; i < ${#verify_sql[@]}; i++)); do
    echo "extracting verify_sql$i on new DB"
    steampipe --install-dir $tmpdir query "${verify_sql[$i]}" --output json > verify$i$i.json
  done

  # stop the service
  steampipe --install-dir $tmpdir service stop

  # verify data is migrated correctly
  for ((i = 0; i < ${#verify_sql[@]}; i++)); do
    assert_equal "$(cat verify$i.json)" "$(cat verify$i$i.json)"
  done

}

# add the setup and verify sql statements here
function setup() {
  # setup sql statements
  setup_sql[0]="create table sample(sample_col_1 char(10), sample_col_2 char(10))"
  setup_sql[1]="insert into sample(sample_col_1,sample_col_2) values ('foo','bar')"
  setup_sql[2]="insert into sample(sample_col_1,sample_col_2) values ('foo1','bar1')"
  setup_sql[3]="create function sample_func() returns integer as 'select 1 as result;' language sql;"

  # verify sql statements
  verify_sql[0]="select * from sample"
  verify_sql[1]="select * from sample_func()"
}

function teardown() {
  rm -rf $tmpdir
  rm -f verify*
}
