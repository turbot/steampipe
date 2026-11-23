load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "expiry year of root.crt should be 9999 and server.crt should be 3yrs from now" {
  current_year=$(date +"%Y")
  steampipe service start

  run openssl x509 -enddate -noout -in $STEAMPIPE_INSTALL_DIR/db/14.24.0/data/root.crt
  echo $output
  # check enddate
  assert_output --partial "notAfter=Dec 31 23:59:59 9999 GMT"

  server_expiry=$((current_year + 3))
  echo $server_expiry

  run openssl x509 -enddate -noout -in $STEAMPIPE_INSTALL_DIR/db/14.24.0/data/server.crt
  echo $output
  # check enddate
  assert_output --partial "$server_expiry"
}

@test "restarting service should not rotate root and server certificates" {
  steampipe service start

  # save file hash
  run cksum $STEAMPIPE_INSTALL_DIR/db/14.24.0/data/root.crt
  id_root=$(echo $output | awk '{print $1}')
  echo $id_root

  # save file hash
  run cksum $STEAMPIPE_INSTALL_DIR/db/14.24.0/data/server.crt
  id_server=$(echo $output | awk '{print $1}')
  echo $id_server

  steampipe service restart
  
  # check file hash after restart
  run cksum $STEAMPIPE_INSTALL_DIR/db/14.24.0/data/root.crt
  id_root_new=$(echo $output | awk '{print $1}')
  echo $id_root_new
  assert_equal $id_root $id_root_new

  # check file hash after restart
  run cksum $STEAMPIPE_INSTALL_DIR/db/14.24.0/data/server.crt
  id_server_new=$(echo $output | awk '{print $1}')
  echo $id_server_new

  # both hashes should be same - which means file did not get regenerated/rotated
  assert_equal $id_server $id_server_new

}

@test "deleting root certificate, service start should regenerate server and root certs" {
  # save file hash
  run cksum $STEAMPIPE_INSTALL_DIR/db/14.24.0/data/server.crt
  id_server=$(echo $output | awk '{print $1}')
  echo $id_server

  # delete root certificate
  rm -f $STEAMPIPE_INSTALL_DIR/db/14.24.0/data/root.crt

  steampipe service start

  # save new file hash
  run cksum $STEAMPIPE_INSTALL_DIR/db/14.24.0/data/server.crt
  id_server_new=$(echo $output | awk '{print $1}')
  echo $id_server_new

  # old and new file hashes should not be equal - deleting root certificate would regenerate/
  # rotate server certificates too
  if [[ "$id_server" == "$id_server_new" ]]; then
    flag=1
  else
    flag=0
  fi
  assert_equal "$flag" "0"
}

@test "adding an encrypted private key should work fine and service should start successfully" {
  steampipe service start
  steampipe service stop

  # a passphrase that the shell would act on, if postgres ran it as part of a command
  marker="$BATS_TEST_TMPDIR/pwned"
  export STEAMPIPE_DATABASE_SSL_PASSWORD="it's a \"pass\" \$(touch $marker) \\n"

  # encrypt the existing server key, so it still matches the server certificate
  key=$STEAMPIPE_INSTALL_DIR/db/14.24.0/data/server.key
  cp $key $BATS_TEST_TMPDIR/server.key
  # OpenSSL 3 needs -traditional to write the encrypted key format that pgx reads; LibreSSL writes that format
  # already, and has no such option
  traditional=""
  if openssl rsa -help 2>&1 | grep -q -- -traditional; then
    traditional="-traditional"
  fi
  run openssl rsa -in $BATS_TEST_TMPDIR/server.key -aes256 $traditional -passout env:STEAMPIPE_DATABASE_SSL_PASSWORD -out $key
  assert_success

  run steampipe service start
  assert_success

  run steampipe query "select 1 as ok" --output csv
  assert_success
  assert_output --partial "ok"

  # the shell did not run any part of the passphrase
  run test -e "$marker"
  assert_failure

  # the passphrase is not in the postgres command line, which every local user can read
  pid=$(jq .pid $STEAMPIPE_INSTALL_DIR/internal/steampipe.json)
  run ps -o command= -p $pid
  assert_success
  refute_output --partial "touch $marker"

  steampipe service stop
  cp $BATS_TEST_TMPDIR/server.key $key
}

function teardown() {
  steampipe service stop --force
}
