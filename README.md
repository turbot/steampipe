[<picture><source media="(prefers-color-scheme: dark)" srcset="https://steampipe.io/images/steampipe-color-logo-and-wordmark-with-white-bubble.svg"><source media="(prefers-color-scheme: light)" srcset="https://steampipe.io/images/steampipe-color-logo-and-wordmark-with-white-bubble.svg"><img width="67%" alt="Steampipe Logo" src="https://steampipe.io/images/steampipe-color-logo-and-wordmark-with-white-bubble.svg"></picture>](https://steampipe.io)

[![plugins](https://img.shields.io/endpoint?url=https://turbot.com/api/badge-stats?stat=apis_supported)](https://hub.steampipe.io/) &nbsp; 
[![slack](https://img.shields.io/endpoint?url=https://turbot.com/api/badge-stats?stat=slack)](https://turbot.com/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp;
[![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-blue)](https://turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

## select * from cloud;

[Steampipe](https://steampipe.io) is **the zero-ETL way** to query APIs and services. Use it to expose data sources to SQL.

**SQL**. It's been the data access standard for decades.

**Live data**. Query APIs in real-time.

**Speed**. Query APIs faster than you ever thought possible.

**Concurrency**. Query many data sources in parallel.

**Single binary**. Use it locally, deploy it in CI/CD pipelines.

## Demo time!

<img alt="steampipe demo" width=500 src="https://steampipe.io/images/steampipe-sql-demo.gif" >

## Documentation

See the [documentation](https://steampipe.io/docs) for:

- [Running queries](https://steampipe.io/docs/query/overview)
- [Managing Steampipe](https://steampipe.io/docs/managing/overview)
- [CLI commands](https://steampipe.io/docs/reference/cli/overview)
- [Integrations](https://steampipe.io/docs/integrations/overview)
- [Developing plugins](https://steampipe.io/docs/develop/overview)

## Install Steampipe

Install Steampipe from the [downloads](https://steampipe.io/downloads) page:

```sh
# MacOS
brew install turbot/tap/steampipe
```

```
# Linux or Windows (WSL2)
sudo /bin/sh -c "$(curl -fsSL https://steampipe.io/install/steampipe.sh)"
```

Install a plugin for your favorite service (e.g. [AWS](https://hub.steampipe.io/plugins/turbot/aws), [Azure](https://hub.steampipe.io/plugins/turbot/azure), [GCP](https://hub.steampipe.io/plugins/turbot/gcp), [GitHub](https://hub.steampipe.io/plugins/turbot/github), [Kubernetes](https://hub.steampipe.io/plugins/turbot/kubernetes), [Hacker News](https://hub.steampipe.io/plugins/turbot/hackernews), etc):

```sh
steampipe plugin install hackernews
```

Query!

```sh
steampipe query
> select * from hackernews_new limit 10
```

## Steampipe plugins

The Steampipe community has grown a suite of [plugins](https://hub.steampipe.io/plugins) that map APIs to database tables. Plugins are available for [AWS](https://hub.steampipe.io/plugins/turbot/aws), [Azure](https://hub.steampipe.io/plugins/turbot/azure), [GCP](https://hub.steampipe.io/plugins/turbot/gcp), [Kubernetes](https://hub.steampipe.io/plugins/turbot/kubernetes), [GitHub](https://hub.steampipe.io/plugins/turbot/github), [Microsoft 365](https://hub.steampipe.io/plugins/turbot/microsoft365), [Salesforce](https://hub.steampipe.io/plugins/turbot/salesforce), and many more.

There are more than 2000 tables in all, each clearly documented with copy/paste/run examples.

## Steampipe distributions

Plugins are available in these distributions.

**Steampipe CLI**. Run [queries](https://steampipe.io/docs/query/overview) that translate APIs to tables in the Postgres instance that's bundled with Steampipe.

**Steampipe Postgres FDWs**. Use [native Postgres Foreign Data Wrappers](https://steampipe.io/docs/steampipe_postgres/overview) to translate APIs to foreign tables.

**Steampipe SQLite extensions**. Use [SQLite extensions](https://steampipe.io/docs/steampipe_sqlite/overview) to translate APIS to SQLite virtual tables.

**Steampipe export tools**. Use [standalone binaries](https://steampipe.io/docs/steampipe_export/overview) that export data from APIs, no database required.

**Turbot Pipes**. Use [Turbot Pipes](https://turbot.com/pipes) to run Steampipe in the cloud.

## Developing

If you want to help develop the core Steampipe binary, these are the steps to build it.

<details>
<summary>Clone</summary>

```sh
git clone git@github.com:turbot/steampipe
```
</details>

<details>
<summary>Build</summary>

```
cd steampipe
make
```

The Steampipe binary lands in `/usr/local/bin/steampipe` directory unless you specify an alternate `OUTPUT_DIR`.
</details>

<details>
<summary>Check the version</summary>

```
$ steampipe --version
steampipe version 0.22.0
```
</details>

<details>
<summary>Install a plugin</summary>

```
$ steampipe plugin install steampipe
```
</details>

<details>
<summary>Run your first query</summary>
 
Try it!

```
steampipe query
> .inspect steampipe
+-----------------------------------+-----------------------------------+
| TABLE                             | DESCRIPTION                       |
+-----------------------------------+-----------------------------------+
| steampipe_registry_plugin         | Steampipe Registry Plugins        |
| steampipe_registry_plugin_version | Steampipe Registry Plugin Version |
+-----------------------------------+-----------------------------------+

> select * from steampipe_registry_plugin;
```
</details>

If you're interested in developing [Steampipe plugins](https://hub.steampipe.io), see our [documentation for plugin developers](https://steampipe.io/docs/develop/overview).

## Turbot Pipes

Bring your team to [Turbot Pipes](https://turbot.com/pipes) to use Steampipe together in the cloud. In a Pipes workspace you can use Steampipe for data access, [Powerpipe](https://github.com/turbot/powerpipe) to visualize query results, and [Flowpipe](https://github.com/turbot/flowpipe) to automate workflow. 

## Open source and contributing

This repository is published under the [AGPL 3.0](https://www.gnu.org/licenses/agpl-3.0.html) license. Please see our [code of conduct](https://github.com/turbot/.github/blob/main/CODE_OF_CONDUCT.md). Contributors must sign our [Contributor License Agreement](https://turbot.com/open-source#cla) as part of their first pull request. We look forward to collaborating with you!

[Steampipe](https://steampipe.io) is a product produced from this open source software, exclusively by [Turbot HQ, Inc](https://turbot.com). It is distributed under our commercial terms. Others are allowed to make their own distribution of the software, but cannot use any of the Turbot trademarks, cloud services, etc. You can learn more in our [Open Source FAQ](https://turbot.com/open-source).

## Get involved

**[Join #steampipe on Slack →](https://turbot.com/community/join)**


