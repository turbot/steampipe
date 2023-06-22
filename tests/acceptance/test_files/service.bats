load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "service stability" {
  echo "# Setting up"
  steampipe query "select 1"
  echo "# Setup Done"
  echo "# Executing tests"

  # pick up the test definitions
  tests=$(cat $FILE_PATH/test_data/source_files/service.json)

  test_indices=$(echo $tests | jq '. | keys[]')
  
  cd $FILE_PATH/test_data/service_mod
  
  # prepare a sample sql file
  echo 'select 1' > sample.sql

  # loop through the tests
  for i in $test_indices; do
    test_name=$(echo $tests | jq -c ".[${i}]" | jq ".name")
    echo ">>> TEST NAME: '$test_name'"
    # pick up the commands that need to run for this test
    runs=$(echo $tests | jq -c ".[${i}]" | jq ".run")
    
    # get the indices of the commands to run
    run_indices=$(echo $runs | jq '. | keys[]')
    
    for k in 1..10; do
      # loop through the run indices
      for j in $run_indices; do
        cmd=$(echo $runs | jq ".[${j}]" | tr -d '"')
        echo ">>>>>>Command: $cmd"
        # run the command
        run $command
        
        # make sure that the command executed successfully
        assert_success
      done
      
      # make sure that there are no steampipe service processes running
      assert_equal $(ps aux | grep steampipe | grep -v bats |grep -v grep | wc -l | tr -d ' ') 0
    done
  done

  # remove the sample sql file
  rm -f sample.sql
}

@test "verify steampipe_connection_state table is getting properly migrated" {
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
  steampipe_uri="https://github.com/turbot/steampipe/releases/download/v0.20.6/steampipe_${target}"
  case $(uname -s) in
    "Darwin") zip_location="${tmpdir}/steampipe.zip" ;;
    "Linux") zip_location="${tmpdir}/steampipe.tar.gz" ;;
    *) echo "Error: steampipe is not supported on '$(uname -s)' yet." 1>&2;exit 1 ;;
  esac
  curl --fail --location --progress-bar --output "$zip_location" "$steampipe_uri"
  tar -xf "$zip_location" -C "$tmpdir"

  # install a couple of plugins which can work with default config
  $tmpdir/steampipe --install-dir $tmpdir plugin install chaos net  
  $tmpdir/steampipe --install-dir $tmpdir query "select * from steampipe_internal.steampipe_connection_state" --output json

  run steampipe --install-dir $tmpdir query "select * from steampipe_internal.steampipe_connection_state" --output json
  
  rm -rf $tmpdir
    
  assert_success
}

function teardown_file() {
  # list running processes
  ps -ef | grep steampipe

  # check if any processes are running
  num=$(ps aux | grep steampipe | grep -v bats | grep -v grep | grep -v tests/acceptance | wc -l | tr -d ' ')
  assert_equal $num 0
}
