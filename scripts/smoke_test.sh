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

/usr/local/bin/steampipe plugin install steampipe
/usr/local/bin/steampipe query "select name from steampipe_registry_plugin limit 1;" --export sps # verify file export
cat query.*.sps | jq '.end_time' # verify file is readable
