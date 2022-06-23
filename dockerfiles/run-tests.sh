#!/usr/bin/env bash

# chown steampipe:0 /home/steampipe/.steampipe/db/14.2.0/data/
# chown steampipe:0 /workspace
steampipe -v
pwd
whoami
uname -a
git clone https://github.com/turbot/steampipe.git
ls -al
cd steampipe
git init
git submodule update --init
git submodule update --recursive

./tests/acceptance/run.sh
