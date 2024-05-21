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

/usr/local/bin/steampipe service start # verify service start
/usr/local/bin/steampipe service status # verify service status
/usr/local/bin/steampipe service stop # verify service stop

/usr/local/bin/steampipe plugin install steampipe

# if block to check the OS and run specific commands
if [ "$(uname -s)" = "Darwin" ]; then
  /usr/local/bin/steampipe query "select name from steampipe_registry_plugin limit 1;" --export /Users/runner/query.sps # verify file export
  cat /Users/runner/query.sps | jq '.end_time' # verify file created is readable
else
  /usr/local/bin/steampipe query "select name from steampipe_registry_plugin limit 1;" --export /home/steampipe/query.sps # verify file export
  cat /home/steampipe/query.sps | jq '.end_time' # verify file created is readable
fi
