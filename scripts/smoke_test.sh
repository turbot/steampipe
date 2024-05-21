#!/bin/sh
# This is a script with set of commands to smoke test a steampipe build.
# The plan is to gradually add more tests to this script.

/usr/local/bin/steampipe --version # check version
/usr/local/bin/steampipe query "select 1 as installed" # verify installation

/usr/local/bin/steampipe plugin install steampipe # verify plugin install
/usr/local/bin/steampipe plugin list # verify plugin listings

/usr/local/bin/steampipe query "select name from steampipe_registry_plugin limit 10;" # verify simple query

/usr/local/bin/steampipe plugin uninstall steampipe # verify plugin uninstall
/usr/local/bin/steampipe plugin list # verify plugin listing after uninstalling

/usr/local/bin/steampipe plugin install steampipe # re-install for other tests
# the file path is different for darwin and linux
if [ "$(uname -s)" = "Darwin" ]; then
  /usr/local/bin/steampipe query "select name from steampipe_registry_plugin limit 1;" --export /Users/runner/query.sps # verify file export
  cat /Users/runner/query.sps | jq '.end_time' # verify file created is readable
else
  /usr/local/bin/steampipe query "select name from steampipe_registry_plugin limit 1;" --export /home/steampipe/query.sps # verify file export
  cat /home/steampipe/query.sps | jq '.end_time' # verify file created is readable
fi

# Ensure the log file path exists before trying to read it
LOG_PATH="/home/steampipe/.steampipe/logs/steampipe-*.log"
if [ "$(uname -s)" = "Darwin" ]; then
  LOG_PATH="/Users/runner/.steampipe/logs/steampipe-*.log"
fi

# Verify log level in logfile
STEAMPIPE_LOG=info /usr/local/bin/steampipe query "select name from steampipe_registry_plugin limit 1;"

# Check if log file exists before attempting to cat it
if ls $LOG_PATH 1> /dev/null 2>&1; then
  bash -c "cat $LOG_PATH | grep '\[INFO\]'"
else
  echo "Log file not found: $LOG_PATH"
  exit 1
fi
