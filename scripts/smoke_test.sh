#!/bin/sh
# This is a script with set of commands to smoke test a steampipe build.
# The plan is to gradually add more tests to this script.
set -e

/usr/local/bin/steampipe --version # check version
/usr/local/bin/steampipe query "select 1 as installed" # verify installation

/usr/local/bin/steampipe plugin install net # verify plugin install
/usr/local/bin/steampipe plugin list # verify plugin listings

/usr/local/bin/steampipe query "select issuer, not_after as exp_date from net_certificate where domain = 'steampipe.io';" # verify simple query

/usr/local/bin/steampipe plugin uninstall net # verify plugin uninstall
/usr/local/bin/steampipe plugin list # verify plugin listing after uninstalling

/usr/local/bin/steampipe plugin install net # re-install for other tests
# the file path is different for darwin and linux
if [ "$(uname -s)" = "Darwin" ]; then
  /usr/local/bin/steampipe query "select issuer, not_after as exp_date from net_certificate where domain = 'steampipe.io';" --export /Users/runner/query.sps # verify file export
  jq '.end_time' /Users/runner/query.sps # verify file created is readable
else
  /usr/local/bin/steampipe query "select issuer, not_after as exp_date from net_certificate where domain = 'steampipe.io';" --export /home/steampipe/query.sps # verify file export
  jq '.end_time' /home/steampipe/query.sps # verify file created is readable
fi

# Ensure the log file path exists before trying to read it
LOG_PATH="/home/steampipe/.steampipe/logs/steampipe-*.log"
if [ "$(uname -s)" = "Darwin" ]; then
  LOG_PATH="/Users/runner/.steampipe/logs/steampipe-*.log"
fi

# Verify log level in logfile
STEAMPIPE_LOG=info /usr/local/bin/steampipe query "select issuer, not_after as exp_date from net_certificate where domain = 'steampipe.io';"

# Check if log file exists before attempting to cat it
if ls $LOG_PATH 1> /dev/null 2>&1; then
  grep '\[INFO\]' $LOG_PATH
else
  echo "Log file not found: $LOG_PATH"
  exit 1
fi
