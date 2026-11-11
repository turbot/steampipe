load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

# These tests only start and stop the service of $STEAMPIPE_INSTALL_DIR, on their own port.
# They never use `service stop --force`, so they leave every other Steampipe service alone.

setup_file() {
  command -v psql > /dev/null || { echo "psql is required for pg_hba.bats" >&3; return 1; }

  export PG_PORT=9293
  export PG_PASSWORD="abcd-efgh-ijkl"
  export HBA_DATA_DIR=$(echo $STEAMPIPE_INSTALL_DIR/db/*/data)
  export HBA_FILE=$HBA_DATA_DIR/pg_hba.conf
  export HBA_DIR=$HBA_DATA_DIR/pg_hba.conf.d
  export HBA_HEADER="# Written by Steampipe each time the database starts. DO NOT EDIT."
}

teardown_file() {
  steampipe service stop
}

start_service() {
  STEAMPIPE_DATABASE_PASSWORD=$PG_PASSWORD steampipe service start --database-port $PG_PORT
}

# connect over TCP from this machine and never ask for a password: -w fails instead of prompting
connect_as() {
  PGPASSWORD="$3" psql -w -h 127.0.0.1 -p $PG_PORT -U "$1" "$2" -tAc "select 1"
}

@test "a fresh install has no preserved pg_hba.conf.d/00-preserved.conf" {
  [ -d "$HBA_DATA_DIR" ]
  [ ! -e "$HBA_DIR/00-preserved.conf" ]
}

@test "after a service start pg_hba.conf starts with Steampipe's header and ends with Steampipe's rules" {
  run start_service
  assert_success

  run head -n 1 "$HBA_FILE"
  assert_output "$HBA_HEADER"

  run tail -n 3 "$HBA_FILE"
  assert_line --index 0 --regexp '^host +all +root +samehost +trust$'
  assert_line --index 1 --regexp '^host +all +steampipe +samehost +trust$'
  assert_line --index 2 --regexp '^host +all +steampipe +all +scram-sha-256$'

  [ ! -e "$HBA_DIR/00-preserved.conf" ]
}

@test "root and steampipe connect from this machine with no password" {
  run connect_as root postgres ""
  assert_success
  run connect_as steampipe steampipe ""
  assert_success
}

@test "steampipe cannot connect to the postgres database" {
  run connect_as steampipe postgres ""
  assert_failure
}

@test "a rule in pg_hba.conf.d wins over Steampipe's rule" {
  steampipe service stop
  mkdir -p "$HBA_DIR"
  echo "host all steampipe samehost scram-sha-256" > "$HBA_DIR/10-password.conf"

  run start_service
  assert_success

  run connect_as steampipe steampipe ""
  assert_failure
  run connect_as steampipe steampipe "$PG_PASSWORD"
  assert_success

  # the rule is gone with its file at the next start
  steampipe service stop
  rm "$HBA_DIR/10-password.conf"
  run start_service
  assert_success
  run connect_as steampipe steampipe ""
  assert_success
}

@test "an install with the old default pg_hba.conf gets the new file at its next start" {
  steampipe service stop
  cat > "$HBA_FILE" <<'EOF'
hostssl all root samehost trust
host    all root samehost trust

hostssl steampipe steampipe samehost trust
host    steampipe steampipe samehost trust
hostssl steampipe steampipe all scram-sha-256
host    steampipe steampipe all scram-sha-256
EOF

  run start_service
  assert_success

  run head -n 1 "$HBA_FILE"
  assert_output "$HBA_HEADER"
  run grep -c hostssl "$HBA_FILE"
  assert_output "0"
  [ ! -e "$HBA_DIR/00-preserved.conf" ]
}

@test "a hand-edited pg_hba.conf is kept as pg_hba.conf.d/00-preserved.conf and its rules apply" {
  steampipe service stop
  cat > "$HBA_FILE" <<'EOF'
host all steampipe samehost scram-sha-256
host all root samehost trust
EOF
  cp "$HBA_FILE" "$BATS_FILE_TMPDIR/edited.conf"

  run start_service
  assert_success

  run cmp "$HBA_DIR/00-preserved.conf" "$BATS_FILE_TMPDIR/edited.conf"
  assert_success

  run grep -l "00-preserved.conf" $STEAMPIPE_INSTALL_DIR/logs/*.log
  assert_success

  # the preserved rule asks for a password and wins over Steampipe's own rule
  run connect_as steampipe steampipe ""
  assert_failure
  run connect_as steampipe steampipe "$PG_PASSWORD"
  assert_success
  run connect_as root postgres ""
  assert_success
}
