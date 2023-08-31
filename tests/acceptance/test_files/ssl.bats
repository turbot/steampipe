load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "expiry year of root.crt should be 9999 and server.crt should be 3yrs from now" {
  current_year=$(date +"%Y")
  steampipe service start

  run openssl x509 -enddate -noout -in $STEAMPIPE_INSTALL_DIR/db/14.2.0/data/root.crt
  echo $output
  assert_output --partial "notAfter=Dec 31 23:59:59 9999 GMT"

  server_expiry=$((current_year + 3))
  echo $server_expiry

  run openssl x509 -enddate -noout -in $STEAMPIPE_INSTALL_DIR/db/14.2.0/data/server.crt
  echo $output
  assert_output --partial "$server_expiry"
}

@test "restarting service should not rotate root and server certificates" {
  steampipe service start

  run cksum $STEAMPIPE_INSTALL_DIR/db/14.2.0/data/root.crt
  id_root=$(echo $output | awk '{print $1}')
  echo $id_root

  run cksum $STEAMPIPE_INSTALL_DIR/db/14.2.0/data/server.crt
  id_server=$(echo $output | awk '{print $1}')
  echo $id_server

  steampipe service restart
  
  run cksum $STEAMPIPE_INSTALL_DIR/db/14.2.0/data/root.crt
  id_root_new=$(echo $output | awk '{print $1}')
  echo $id_root_new
  assert_equal $id_root $id_root_new

  run cksum $STEAMPIPE_INSTALL_DIR/db/14.2.0/data/server.crt
  id_server_new=$(echo $output | awk '{print $1}')
  echo $id_server_new
  assert_equal $id_server $id_server_new

}

function teardown() {
  steampipe service stop
}