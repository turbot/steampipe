#!/usr/bin/env bash

# check version
steampipe -v

# clone the repo, to run the test suite
git clone https://github.com/turbot/steampipe.git
cd steampipe

# initialize git along with bats submodules
git init
git submodule update --init
git submodule update --recursive

# run test suite
./tests/acceptance/run.sh
