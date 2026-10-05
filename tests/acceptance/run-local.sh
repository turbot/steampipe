#!/bin/bash

MY_PATH="`dirname \"$0\"`"              # relative
MY_PATH="`( cd \"$MY_PATH\" && pwd )`"  # absolutized and normalized
REPO_ROOT="`( cd \"$MY_PATH/../..\" && pwd )`"

export TIME_TO_QUERY=3                  # overriding since it takes more than 2secs to run locally
export TZ=UTC
export STEAMPIPE_UPDATE_CHECK=false
export STEAMPIPE_LOG=info

# the test files CI runs, in CI's order (see .github/workflows/11-test-acceptance.yaml)
ALL_FILES="migration brew installation plugin connection_config service settings ssl blank_aggregators search_path chaos_and_query date_time_types dynamic_schema dynamic_aggregators cache performance config_precedence cloud schema_cloning exit_codes force_stop"
# CI does not run these on macOS
if [ "$(uname)" = "Darwin" ]; then
  ALL_FILES="${ALL_FILES/migration /}"
  ALL_FILES="${ALL_FILES/ force_stop/}"
fi

FILE_TIMEOUT=1800   # seconds
IDLE_LIMIT=30       # seconds without output, once every planned test has reported

BIN_DIR=$(mktemp -d)
LOG_DIR=$(mktemp -d)
INSTALL_DIRS=""     # every install dir this script has created
CUR_PID=""
TAIL_PID=""

# mtime of a file in epoch seconds
mtime() {
  if [ "$(uname)" = "Darwin" ]; then stat -f %m "$1"; else stat -c %Y "$1"; fi
}

# send a signal to a process and all its descendants
kill_tree() {
  local pid=$1 sig=$2 child
  for child in $(pgrep -P $pid 2>/dev/null); do
    kill_tree $child $sig
  done
  kill -$sig $pid 2>/dev/null
}

# stop whatever is still running under an install dir this script created, or from the binary it built.
# The service is asked to stop first; the plugin manager ignores SIGTERM, so the rest get SIGKILL.
stop_leftovers() {
  local dir=$1
  if [ -n "$dir" ]; then
    STEAMPIPE_INSTALL_DIR=$dir steampipe service stop > /dev/null 2>&1
    pkill -f "$dir" 2>/dev/null
  fi
  sleep 2
  [ -n "$dir" ] && pkill -KILL -f "$dir" 2>/dev/null
  pkill -KILL -f "$BIN_DIR/steampipe" 2>/dev/null
  return 0
}

cleanup() {
  local code=$?
  trap '' EXIT INT TERM
  [ -n "$TAIL_PID" ] && kill $TAIL_PID 2>/dev/null
  [ -n "$CUR_PID" ] && kill_tree $CUR_PID KILL
  local dir
  for dir in $INSTALL_DIRS; do
    stop_leftovers $dir
    rm -rf $dir
  done
  rm -rf $BIN_DIR
  exit $code
}
trap cleanup EXIT
trap 'exit 130' INT TERM

echo "Building steampipe from $REPO_ROOT"
(cd $REPO_ROOT && CGO_ENABLED=0 go build -o $BIN_DIR/steampipe .)
if [ $? -ne 0 ]; then
  echo "Build failed"
  exit 1
fi
export PATH=$BIN_DIR:$PATH

if [ $# -eq 0 ]; then
  FILES=$ALL_FILES
else
  FILES=""
  for arg in "$@"; do FILES="$FILES ${arg%.bats}"; done
fi

echo "Logs: $LOG_DIR"
SUMMARY=""
FAILED=0

for f in $FILES; do
  log=$LOG_DIR/$f.tap
  export STEAMPIPE_INSTALL_DIR=$(mktemp -d)
  INSTALL_DIRS="$INSTALL_DIRS $STEAMPIPE_INSTALL_DIR"
  WD=$(mktemp -d)

  echo
  echo "=== $f"
  start=$(date +%s)
  result=""
  leftovers=""

  # a fresh installation per file, warmed up the way CI does
  (
    cd $WD
    echo "Working directory: $WD"
    echo "Install directory: $STEAMPIPE_INSTALL_DIR"
    steampipe query "select 1 as setup_complete" || exit 1
    steampipe plugin install chaos chaosdynamic --progress=false || exit 1
  ) > $log.setup 2>&1
  if [ $? -ne 0 ]; then
    cat $log.setup
    result="SETUP-FAILED"
  fi

  if [ -z "$result" ]; then
    (
      cd $WD
      MY_PATH=$MY_PATH exec $MY_PATH/run.sh $f.bats
    ) > $log 2>&1 &
    CUR_PID=$!
    tail -n +1 -f $log &
    TAIL_PID=$!

    while kill -0 $CUR_PID 2>/dev/null; do
      # A service or plugin manager left running by the file keeps bats's output pipe open, so bats
      # never exits. Once every planned test has reported and the output has gone quiet, stop what
      # the file left behind and bats then exits on its own.
      plan=$(grep -m1 -E '^1\.\.[0-9]+' $log | cut -d. -f3)
      reported=$(grep -cE '^(ok|not ok) ' $log)
      idle=$(( $(date +%s) - $(mtime $log) ))
      if [ -n "$plan" ] && [ "$reported" -ge "$plan" ] && [ $idle -ge $IDLE_LIMIT ] && [ -z "$leftovers" ]; then
        leftovers=" (left a service running)"
        stop_leftovers $STEAMPIPE_INSTALL_DIR
      fi
      if [ $(( $(date +%s) - start )) -ge $FILE_TIMEOUT ]; then
        kill_tree $CUR_PID TERM
        sleep 5
        kill_tree $CUR_PID KILL
        result="TIMEOUT"
        break
      fi
      sleep 2
    done
    if [ -z "$result" ]; then
      wait $CUR_PID
      status=$?
      if [ $status -eq 0 ]; then result="PASS"; else result="FAIL (exit $status)"; fi
    fi
    CUR_PID=""
    kill $TAIL_PID 2>/dev/null
    TAIL_PID=""
  fi

  # leave nothing running for the next file
  stop_leftovers $STEAMPIPE_INSTALL_DIR
  rm -rf $STEAMPIPE_INSTALL_DIR $WD

  [ "${result%% *}" = "PASS" ] || FAILED=$((FAILED + 1))
  ok=$(grep -c '^ok ' $log 2>/dev/null)
  notok=$(grep -c '^not ok ' $log 2>/dev/null)
  SUMMARY="$SUMMARY$(printf '%-24s %s%s  ok=%s not ok=%s  %ss' $f "$result" "$leftovers" "${ok:-0}" "${notok:-0}" $(( $(date +%s) - start )))
"
done

echo
echo "=== Summary (logs in $LOG_DIR)"
printf '%s' "$SUMMARY"
if [ $FAILED -ne 0 ]; then
  echo "$FAILED file(s) failed"
  exit 1
fi
echo "All files passed"
