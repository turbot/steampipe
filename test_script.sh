#!/bin/sh

./install.sh v0.13.6
rm -rf ~/.steampipe
steampipe service start

steampipe plugin install github
rm -f ~/.steampipe/config/github.spc
cp github_token.spc ~/.steampipe/config/github.spc

cat ~/.steampipe/config/github.spc
sleep 10
steampipe query "select 1"

steampipe query "create table repo_names as select full_name from github_search_repository where query = 'org:turbot' limit 5;"

steampipe query "create function repo_names_fn() returns table (full_name text) as 'select full_name from github_my_repository limit 5' language sql;"

steampipe query "create materialized view m_repo_names_from_fn as select * from repo_names_fn();"

steampipe service stop

./install.sh v0.14.0-rc.1

steampipe service start

steampipe query "select * from repo_names"
steampipe query "select * from repo_names_fn()"
steampipe query "select * from m_repo_names_from_fn"

steampipe service stop
