#!/usr/bin/env bash

# chown steampipe:0 /home/steampipe/.steampipe/db/14.2.0/data/

steampipe -v
pwd
cat /etc/sudoers
sudo git clone https://github.com/turbot/steampipe.git
ls -al
cd steampipe
git init
git submodule update --init
git submodule update --recursive

./tests/acceptance/run.sh
