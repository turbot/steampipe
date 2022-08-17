[<img width="524px" src="https://steampipe.io/images/steampipe_logo_wordmark_color.svg" />](https://steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

[![plugins](https://img.shields.io/badge/plugins-83-green)](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp; [![mods](https://img.shields.io/badge/controls-3000-green)](https://hub.steampipe.io/mods?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp; [![dashboards](https://img.shields.io/badge/dashboards-744-green)](https://hub.steampipe.io/mods?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp; [![slack](https://img.shields.io/badge/slack-800-E01563)](https://steampipe.io/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp; [![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-gold)](https://turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

Steampipe reads live data from APIs into Postgres. Data sources ([plugins](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)) include the major clouds (AWS, Azure, GCP), SaaS (GitHub, Snowflake, CrowdStrike), Containers (Docker, Kubernetes), IaC (Terraform, CloudFormation), logs (Splunk, Algolia), and [more](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme). 

With [Steampipe](https://steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) you can:

- **Query** → Write SQL queries that report on (and join across!) clouds, business apps, code, logs, and more.

- **Report** → Verify that your cloud resources comply with security & compliance benchmarks such as CIS, NIST, and HIPAA.

- **View** → Explore query results on dashboards.

## Steampipe CLI - The SQL console for API queries

Steampipe provides a growing suite of [plugins](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) that map APIs to Postgres tables. The [interactive query shell](https://steampipe.io/docs/query/query-shell?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) is one way you can query those tables. 

<br/>

<img marginTop="200" width="524" src="https://steampipe.io/images/steampipe-sql-demo.gif" />

<br/>

You can also use psql, pgcli, Metabase, Tableau, or [any client](https://steampipe.io/docs/cloud/integrations/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) that can connect to Postgres.

### Get started with the CLI

[Install](https://steampipe.io/downloads?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) Steampipe for:

Linux or WSL

```
sudo /bin/sh -c "$(curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh)"
```

MacOS

```
brew tap turbot/tap
brew install steampipe
```

Add your first [plugin](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).
```
steampipe plugin install net
```

Run `steampipe query` to launch the interactive shell.
```
steampipe query
```

Run your first query!
```
select
  *
from
  net_certificate
where
  domain = 'google.com';
```

### Learn more about the CLI

- It's [just SQL](https://steampipe.io/docs/sql/steampipe-sql?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)!

- You can run queries [on the command line](https://steampipe.io/docs/query/overview#non-interactive-batch-query-mode?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) and include them in scripts.

- Other [commands](https://steampipe.io/docs/reference/cli/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) run benchmarks, launch Steampipe as a service, and start the dashboard server.

- [Meta-commands](https://steampipe.io/docs/reference/dot-commands/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) control caching, [environment variables](https://steampipe.io/docs/reference/env-vars/overview), the [search path](https://steampipe.io/docs/guides/search-path?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme), and more.

- Queries can run in [batch mode](https://steampipe.io/docs/query/batch-query?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

- You can bundle connections (e.g. for many AWS accounts) using an [aggregator](https://steampipe.io/docs/managing/connections?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#using-aggregators).
  
## Steampipe mods: Benchmarks and dashboards

Steampipe [mods](https://hub.steampipe.io/mods?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) build on the plugins. Some run suites of controls that check for compliance with security & compliance benchmarks (e.g. CIS, NIST, HIPAA). 

<br/>


![readme-aws-cis-1 4](https://user-images.githubusercontent.com/46509/185204883-54311f57-759d-410f-92bb-d1e92373a35b.gif)



<br/>

Others visualize query results using charts, tables, and [other widgets](https://steampipe.io/docs/reference/mod-resources/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

<br/>

![readme-aws-vpc-dashboard](https://user-images.githubusercontent.com/46509/185204989-05d594e7-8f6b-4998-8550-0934f6ace522.gif)

<br/>

### Get started with benchmarks and dashboards

The [Net Insights](https://hub.steampipe.io/mods/turbot/net_insights?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) mod works with the Net plugin shown above. To run it, first clone its repo and change to that directory.

```
git clone https://github.com/turbot/steampipe-mod-net-insights
cd steampipe-mod-net-insights
```

### Run benchmarks in the CLI

All the benchmarks:

```sh
steampipe check all
```

A single benchmark:

```sh
steampipe check benchmark.dns_best_practices
```

A single control:

```sh
steampipe check control.dns_ns_name_valid
```

### Run benchmarks as dashboards

Launch the dashboard server: `steampipe dashboard`

Open `http://localhost:9194` in your browser. The home page lists available dashboards. Click `DNS Best Practices` to view that dashboard.

Note that the default domains are `microsoft.com` and `github.com`. You can [change those defaults](https://hub.steampipe.io/mods/turbot/net_insights#configuration?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) to check other domains.

### Explore query results on dashboards

Dashboards use charts, tables, and interactive widgets to help you explore and visualize your resources. The [AWS Insights](https://hub.steampipe.io/mods/turbot/aws_insights?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme), for example, provides dozens of dashboards that exercise the full set of widgets. To explore these dashboards, first install the [AWS plugin](https://hub.steampipe.io/plugins/turbot/aws?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) and [authenticate](https://hub.steampipe.io/plugins/turbot/aws#configuration?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

Then clone `AWS Insights`, change to its directory, launch `steampipe dashboard`, and open `localhost:9194`.

### Learn more about benchmarks and dashboards

 **Benchmarks**

- Use [search paths](https://steampipe.io/docs/check/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#more-examples) to specify connections.

- Review `steampipe check` [commands](https://steampipe.io/docs/reference/cli/check?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) to filter on control tags or set  variables.

- Output results in a variety of [formats](https://steampipe.io/docs/reference/cli/check?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#output-formats).

- Write your own [custom ouput template](https://steampipe.io/docs/develop/writing-control-output-templates?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

- Build your own [custom benchmark](https://steampipe.io/docs/mods/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

**Dashboards**

- Reveal the [HCL and SQL source code](https://steampipe.io/docs/dashboard/panel?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) for dashboard panels, and download the data.

- Review [dashboard commands](https://steampipe.io/docs/reference/cli/dashboard?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) to set search paths, dashboard ports, and variables.

- Explore [dashboard widgets](https://steampipe.io/docs/reference/mod-resources/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#dashboards).

- Build your own [custom dashboard](https://steampipe.io/docs/mods/writing-dashboards?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

## Community

We thrive on feedback and community involvement!

**Have a question?** → Join our [Slack community](https://steampipe.io/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) or open a [GitHub issue](https://github.com/turbot/steampipe/issues/new/choose).

**Want to get involved?** → Learn how to [contribute](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md).

**Want to work with the team?** → We are [hiring](https://turbot.com/careers?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)!

## Steampipe Cloud

Want a hosted version of Steampipe? Bring your team to [Steampipe Cloud](https://cloud.steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).  

