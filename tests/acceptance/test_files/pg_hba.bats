load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"
bats_require_minimum_version 1.5.0

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

  # steampipe query has no --database-port flag
  cat > "$STEAMPIPE_INSTALL_DIR/config/pg_hba_port.spc" <<EOF
options "database" {
  port = $PG_PORT
}
EOF
}

teardown_file() {
  steampipe service stop
  rm -f "$STEAMPIPE_INSTALL_DIR/config/pg_hba_port.spc"
}

start_service() {
  STEAMPIPE_DATABASE_PASSWORD=$PG_PASSWORD steampipe service start --database-port $PG_PORT
}

connect_as_root_by_socket() {
  psql -w -h "$HBA_DATA_DIR" -p $PG_PORT -U root postgres -tAc "select 1"
}

# connect over TCP from this machine and never ask for a password: -w fails instead of prompting
connect_as() {
  PGPASSWORD="$3" psql -w -h 127.0.0.1 -p $PG_PORT -U "$1" "$2" -tAc "select 1"
}

query_as_steampipe() {
  psql -w -h 127.0.0.1 -p $PG_PORT -U steampipe steampipe -tAc "$1"
}

# postgres runs with -D <data directory>, so its command line names the data directory
assert_no_postgres() {
  run pgrep -f "$HBA_DATA_DIR"
  assert_failure
  run steampipe service status
  assert_output --partial "Service is not running"
}

# a pg_hba.conf with no header and no checksum, as the user makes it
write_hba_from_body() {
  sed '1,/^# checksum: /d' "$HBA_FILE" > "$BATS_FILE_TMPDIR/body.conf"
  mv "$BATS_FILE_TMPDIR/body.conf" "$HBA_FILE"
}

@test "a fresh install has Steampipe's pg_hba.conf before the first start" {
  # the install writes it at once after initdb, so pg_hba.conf is never missing
  run head -n 1 "$HBA_FILE"
  assert_output "$HBA_HEADER"
  [ -d "$HBA_DIR" ]
}

@test "after a service start pg_hba.conf starts with Steampipe's header and ends with Steampipe's rules" {
  run start_service
  assert_success

  run head -n 1 "$HBA_FILE"
  assert_output "$HBA_HEADER"

  run tail -n 3 "$HBA_FILE"
  assert_line --index 0 --regexp '^local +all +root +trust$'
  assert_line --index 1 --regexp '^host +all +steampipe +samehost +trust$'
  assert_line --index 2 --regexp '^host +all +steampipe +all +scram-sha-256$'
}

@test "root connects through the unix socket and is rejected over TCP" {
  run connect_as_root_by_socket
  assert_success

  run connect_as root postgres ""
  assert_failure
}

@test "the socket is in the data directory, only its owner can open it, and there is none in /tmp" {
  local socket="$HBA_DATA_DIR/.s.PGSQL.$PG_PORT"
  [ -S "$socket" ]
  run ls -l "$socket"
  assert_output --regexp '^srwx------'
  [ ! -e /tmp/.s.PGSQL.$PG_PORT ]
}

@test "steampipe connects over TCP from this machine with no password" {
  run connect_as steampipe steampipe ""
  assert_success
}

@test "another OS user cannot connect as root through the socket" {
  sudo -n true || skip "sudo -n does not work here"

  # the service is up and root works for this OS user
  run connect_as_root_by_socket
  assert_success

  run sudo -n -u nobody psql -w -h "$HBA_DATA_DIR" -p $PG_PORT -U root postgres -tAc "select 1"
  assert_failure
  assert_output --partial "Permission denied"
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
}

@test "a hand-edited older pg_hba.conf with no local rule for root stops the start, and is not changed" {
  steampipe service stop
  cat > "$HBA_FILE" <<'EOF'
host all steampipe samehost scram-sha-256
host all root samehost trust
EOF
  cp "$HBA_FILE" "$BATS_FILE_TMPDIR/edited.conf"

  run start_service
  assert_equal "$status" 32
  assert_output --partial "has rules that Steampipe did not write, and they do not let Steampipe connect"
  assert_output --partial "local all root trust"
  # the user's rule is listed, the rule that Steampipe 2.4 wrote is not
  assert_output --partial "Your rules (the rules that older Steampipe versions did not write):
    host all steampipe samehost scram-sha-256
"
  run cmp "$HBA_FILE" "$BATS_FILE_TMPDIR/edited.conf"
  assert_success
  assert_no_postgres

  # a batch query fails the same way: the message on stderr, nothing on stdout
  run --separate-stderr steampipe query "select 1"
  assert_failure
  assert_equal "$output" ""
  [[ "$stderr" == *"has rules that Steampipe did not write"* ]]
  run cmp "$HBA_FILE" "$BATS_FILE_TMPDIR/edited.conf"
  assert_success
  assert_no_postgres
}

@test "the same file with local all root trust added starts, is not changed, and its rules apply" {
  { echo "local all root trust"; cat "$BATS_FILE_TMPDIR/edited.conf"; } > "$HBA_FILE"
  cp "$HBA_FILE" "$BATS_FILE_TMPDIR/yours.conf"

  run start_service
  assert_success
  run cmp "$HBA_FILE" "$BATS_FILE_TMPDIR/yours.conf"
  assert_success

  run connect_as_root_by_socket
  assert_success
  # the user's rule asks for a password
  run connect_as steampipe steampipe ""
  assert_failure
  run connect_as steampipe steampipe "$PG_PASSWORD"
  assert_success
  # the user's file still lets root connect over TCP
  run connect_as root postgres ""
  assert_success
}

@test "a Steampipe pg_hba.conf with its header lines deleted is the user's, and pg_hba.conf.d is not used" {
  steampipe service stop
  # give the file back to Steampipe, and start once to write it
  rm "$HBA_FILE"
  run start_service
  assert_success
  steampipe service stop

  write_hba_from_body
  cp "$HBA_FILE" "$BATS_FILE_TMPDIR/yours.conf"
  echo "host all steampipe samehost scram-sha-256" > "$HBA_DIR/10-unused.conf"

  run start_service
  assert_success
  assert_output --partial "is your own file, so Steampipe does not use the files in $HBA_DIR: 10-unused.conf"
  run cmp "$HBA_FILE" "$BATS_FILE_TMPDIR/yours.conf"
  assert_success
  # the unused rule would ask for a password
  run connect_as steampipe steampipe ""
  assert_success

  steampipe service stop
  rm "$HBA_DIR/10-unused.conf" "$HBA_FILE"
}

@test "an edited Steampipe pg_hba.conf stops a restart and a start, and is not changed" {
  run start_service
  assert_success
  run head -n 1 "$HBA_FILE"
  assert_output "$HBA_HEADER"
  echo "host all bob all md5" >> "$HBA_FILE"
  cp "$HBA_FILE" "$BATS_FILE_TMPDIR/edited.conf"

  run steampipe service restart
  assert_equal "$status" 32
  assert_output --partial "pg_hba.conf was changed after Steampipe wrote it"
  assert_output --partial "Rules in the file that Steampipe would not write:
    host all bob all md5
"
  # the running service keeps running
  run connect_as_root_by_socket
  assert_success

  steampipe service stop
  run start_service
  assert_equal "$status" 32
  assert_output --partial "pg_hba.conf was changed after Steampipe wrote it"
  run cmp "$HBA_FILE" "$BATS_FILE_TMPDIR/edited.conf"
  assert_success
  assert_no_postgres

  # discard the change: Steampipe writes its file again
  rm "$HBA_FILE"
  run start_service
  assert_success
  run grep -c "bob" "$HBA_FILE"
  assert_output "0"
}

@test "a rule in pg_hba.conf.d that refuses root stops the start at once" {
  steampipe service stop
  echo "local all root reject" > "$HBA_DIR/05-no-root.conf"

  # a refused login is not retried until the start time-out
  local started=$SECONDS
  STEAMPIPE_DATABASE_START_TIMEOUT=120 run start_service
  local took=$((SECONDS - started))
  assert_equal "$status" 32
  assert_output --partial "a rule in $HBA_DIR stops Steampipe from connecting as root through the unix socket"
  assert_output --partial "05-no-root.conf"
  [ "$took" -lt 60 ] || fail "the start took ${took}s"
  assert_no_postgres

  rm "$HBA_DIR/05-no-root.conf"
}

@test "a downgrade to v2.4.9 keeps the data with one rule added, and the next upgrade keeps it too" {
  # v2.4.9's backup fails on macOS before it connects: it cannot read the command line of every process
  [ "$(uname -s)" = Linux ] || skip "v2.4.9's database backup works only on Linux"
  local arch
  case "$(uname -m)" in x86_64) arch=amd64 ;; aarch64) arch=arm64 ;; *) skip "no v2.4.9 build for $(uname -m)" ;; esac
  curl -fsSL -o "$BATS_FILE_TMPDIR/v249.tar.gz" "https://github.com/turbot/steampipe/releases/download/v2.4.9/steampipe_linux_$arch.tar.gz"
  mkdir -p "$BATS_FILE_TMPDIR/v249"
  tar -xzf "$BATS_FILE_TMPDIR/v249.tar.gz" -C "$BATS_FILE_TMPDIR/v249"
  local v249="$BATS_FILE_TMPDIR/v249/steampipe"
  # v2.4.9 uses Postgres 14.19.0, this version 14.24.0: a downgrade is a dump and restore
  local v249_data="$STEAMPIPE_INSTALL_DIR/db/14.19.0/data"

  run start_service
  assert_success
  run query_as_steampipe "create table public.downgrade_check (x int); insert into public.downgrade_check values (42)"
  assert_success
  steampipe service stop

  # with no step, v2.4.9 starts with an empty public schema: its backup connects as root over TCP
  STEAMPIPE_DATABASE_START_TIMEOUT=15 run "$v249" service start --database-port $PG_PORT
  assert_success
  run query_as_steampipe "select x from public.downgrade_check"
  assert_failure
  "$v249" service stop
  # the data is still in this version's data directory
  [ -f "$HBA_DATA_DIR/PG_VERSION" ]

  # the step: let root connect over TCP, and remove the empty v2.4.9 install so that it runs the backup again
  echo "host all root samehost trust" >> "$HBA_FILE"
  rm -rf "$STEAMPIPE_INSTALL_DIR/db/14.19.0"
  run "$v249" service start --database-port $PG_PORT
  assert_success
  run query_as_steampipe "select x from public.downgrade_check"
  assert_output "42"
  "$v249" service stop
  # v2.4.9 removes the installation that it restored from
  [ ! -e "$HBA_DATA_DIR" ]
  [ -f "$v249_data/PG_VERSION" ]

  # upgrade again: this version moves v2.4.9's data directory, and writes its pg_hba.conf over v2.4.9's
  run start_service
  assert_success
  run query_as_steampipe "select x from public.downgrade_check"
  assert_output "42"
  run head -n 1 "$HBA_FILE"
  assert_output "$HBA_HEADER"
  [ ! -e "$STEAMPIPE_INSTALL_DIR/db/14.19.0" ]
  run query_as_steampipe "drop table public.downgrade_check"
  assert_success
  steampipe service stop
}

@test "service start refuses an install directory that makes the socket path too long" {
  steampipe service stop
  # the same install, through a link with a long name: the socket path is over the OS's limit
  local long_dir="$BATS_FILE_TMPDIR/$(printf 'x%.0s' $(seq 100))"
  ln -s "$STEAMPIPE_INSTALL_DIR" "$long_dir"

  STEAMPIPE_INSTALL_DIR="$long_dir" run steampipe service start --database-port $PG_PORT
  assert_equal "$status" 32
  assert_output --partial "Set STEAMPIPE_INSTALL_DIR to a shorter path"

  # postgres did not start
  run pgrep -f "$long_dir"
  assert_failure
  rm "$long_dir"
}

# this test is last: it leaves no service running, and stops this install's processes itself
@test "a running service with no socket in the data directory says to stop it once with --force" {
  run start_service
  assert_success
  # a service that Steampipe 2.4 or earlier started has no socket in the data directory
  rm "$HBA_DATA_DIR/.s.PGSQL.$PG_PORT"

  run steampipe service stop
  assert_failure
  assert_output --partial "started by an older version of Steampipe"
  assert_output --partial "steampipe service stop --force"

  # stop only this install's processes: service stop --force would stop every service on the machine.
  # Postgres, the plugin manager (--install-dir) and the plugins all name the install on their command lines.
  pkill -f "$STEAMPIPE_INSTALL_DIR" || true
  for _ in $(seq 50); do pgrep -f "$STEAMPIPE_INSTALL_DIR" > /dev/null || break; sleep 0.2; done
  run pgrep -f "$STEAMPIPE_INSTALL_DIR"
  assert_failure
  rm -f "$STEAMPIPE_INSTALL_DIR/internal/steampipe.json" "$STEAMPIPE_INSTALL_DIR/internal/plugin_manager.json"
}
