#!/bin/sh
# This is a script with set of commands to smoke test a steampipe build.
# The plan is to gradually add more tests to this script.

/usr/local/bin/steampipe --version
/usr/local/bin/steampipe query "select 1 as installed"
/usr/local/bin/steampipe plugin install steampipe
/usr/local/bin/steampipe plugin list
/usr/local/bin/steampipe query "select name from steampipe_registry_plugin limit 10;"