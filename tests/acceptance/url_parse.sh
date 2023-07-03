#!/bin/bash

# The acceptance tests use this script to generate the delete snapshot request URL.

# extract the protocol
proto="$(echo $1 | grep :// | sed -e's,^\(.*://\).*,\1,g')"

# remove the protocol
url="$(echo ${1/$proto/})"

# extract the user (if any)
user="$(echo $url | grep @ | cut -d@ -f1)"

# extract the host and port
hostport="$(echo ${url/$user@/} | cut -d/ -f1)"

# by request host without port    
host="$(echo $hostport | sed -e 's,:.*,,g')"

# by request - try to extract the port
port="$(echo $hostport | sed -e 's,^.*:,:,g' -e 's,.*:\([0-9]*\).*,\1,g' -e 's,[^0-9],,g')"

# extract the path (if any)
path="$(echo $url | grep / | cut -d/ -f2-)"

# echo "  url: $url"
# echo "  proto: $proto"
# echo "  user: $user"
# echo "  host: $host"
# echo "  port: $port"
# echo "  path: $path"

echo "$proto$host/api/v0/$path"