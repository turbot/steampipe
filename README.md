[<picture><source media="(prefers-color-scheme: dark)" srcset="https://steampipe.io/images/steampipe-color-logo-and-wordmark-with-white-bubble.svg"><source media="(prefers-color-scheme: light)" srcset="https://steampipe.io/images/steampipe-color-logo-and-wordmark-with-white-bubble.svg"><img width="67%" alt="Steampipe Logo" src="https://steampipe.io/images/steampipe-color-logo-and-wordmark-with-white-bubble.svg"></picture>](https://steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

[![plugins](https://img.shields.io/badge/apis_supported-145-blue)](https://hub.steampipe.io/) &nbsp; 
[![slack](https://img.shields.io/badge/slack-2695-blue)](https://turbot.com/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp;
[![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-blue)](https://turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)


Steampipe is the zero-ETL solution for getting data directly from APIs and services. We offer these Steampipe engines:

- Steampipe CLI. Use the Steampipe core engine to translate APIs to tables in the Postgres instance that's bundled with Steampipe.

- Steampipe Postgres FDWs. Use [native Postgres Foreign Data Wrappers](https://steampipe.io/docs/steampipe_postgres/overview) to translate APIs to foreign tables.

- Steampipe SQLite extensions. Use [SQLite extensions](https://steampipe.io/docs/steampipe_sqlite/overview) to translate APIS to SQLite virtual tables.

- Steampipe export tools. Use [standalone binaries](https://steampipe.io/docs/steampipe_export/overview) that export data from APIs, no database required.

- Turbot Pipes. Use [Turbot Pipes](https://turbot.com/pipes) to run Steampipe in the cloud.

## Steampipe API plugins

The Steampipe community has grown a suite of [plugins](https://hub.steampipe.io/) that map APIs to database tables. They work with all Steampipe engines.


## Install Steampipe

 The <a href="https://steampipe.io/downloads?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">downloads</a> page shows you how but tl;dr:
 
Linux or WSL

```sh
sudo /bin/sh -c "$(curl -fsSL https://steampipe.io/install/steampipe.sh)"
```

MacOS

```sh
brew tap turbot/tap
brew install steampipe
```

## Install a plugin

 Choose a plugin from the [hub](https://hub.steampipe.io/), for example: [Hacker News](https://hub.steampipe.io/plugins/turbot/hackernews).

 Run the `steampipe plugin` command to install it.

```sh
steampipe plugin install hackernews
```

## Query with a plugin

Run a query using `psql` — or another Postgres client , or [Powerpipe](https://powerpipe.io) — to query a table provided by the installed plugin.


```
psql -h localhost -p 9193 -d steampipe -U steampipe
```

```
steampipe=> select * from hackernews_new limit 10
```

## Developing

Prerequisites:

- [Golang](https://golang.org/doc/install) Version 1.19 or higher.

Clone the repo.

```sh
git clone https://github.com/turbot/steampipe
cd steampipe
```

Build, which automatically installs the new version to your `/usr/local/bin/steampipe` directory:

```
make
```

Check the version.

```
$ steampipe -v
steampipe version 0.21.1
```

Install a plugin and run a query as per above.

## Open source & contributing

This repository is published under the [AGPL 3.0](https://www.gnu.org/licenses/agpl-3.0.html) license. Please see our [code of conduct](https://github.com/turbot/.github/blob/main/CODE_OF_CONDUCT.md). Contributors must sign our [Contributor License Agreement](https://turbot.com/open-source#cla) as part of their first pull request. We look forward to collaborating with you!

[Steampipe](https://steampipe.io) is a product produced from this open source software, exclusively by [Turbot HQ, Inc](https://turbot.com). It is distributed under our commercial terms. Others are allowed to make their own distribution of the software, but cannot use any of the Turbot trademarks, cloud services, etc. You can learn more in our [Open Source FAQ](https://turbot.com/open-source).

## Turbot Pipes

Want a hosted version of Steampipe? Bring your team to [Turbot Pipes](https://pipes.turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

## Get involved

**[Join #steampipe on Slack →](https://turbot.com/community/join)**

Want to help but don't know where to start? Pick up one of the `help wanted` issues:
* [Steampipe](https://github.com/turbot/steampipe/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22)