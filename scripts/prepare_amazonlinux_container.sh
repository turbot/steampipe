#!/bin/sh
# This is a a script to install dependencies/packages, create user, and assign necessary permissions in the amazonlinux 2023 container.
# Used in release smoke tests. 

# update yum and install required packages
yum install -y shadow-utils tar gzip ca-certificates jq

# Extract the steampipe binary
tar -xzf /artifacts/linux.tar.gz -C /usr/local/bin

# Create user, since steampipe cannot be run as root
useradd -m steampipe
          
# Ensure the binary is executable and owned by steampipe and is executable
chown steampipe:steampipe /usr/local/bin/steampipe
chmod +x /usr/local/bin/steampipe

# Ensure the script is executable
chown steampipe:steampipe /scripts/smoke_test.sh
chmod +x /scripts/smoke_test.sh
