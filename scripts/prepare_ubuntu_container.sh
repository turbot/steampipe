#!/bin/sh
# This is a a script to install dependencies/packages, create user, and assign necessary permissions in the ubuntu 24 container.
# Used in release smoke tests.

# update apt and install required packages
apt-get update
apt-get install -y tar ca-certificates jq

# Extract the steampipe binary
tar -xzf /artifacts/linux.tar.gz -C /usr/local/bin

# Make the binary executable
chmod +x /usr/local/bin/steampipe

# Create user, since steampipe cannot be run as root
useradd -m steampipe

# Make the scripts executable
chown steampipe:steampipe /scripts/smoke_test.sh
chmod +x /scripts/smoke_test.sh
