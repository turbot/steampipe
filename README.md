<img width="524px" src="https://steampipe.io/images/steampipe_logo_wordmark_color.svg" />

<img src="https://img.shields.io/badge/api_plugins-83-green"> &nbsp; 
<img src="https://img.shields.io/badge/compliance_frameworks-21-green"> &nbsp;
<img src="https://img.shields.io/badge/controls-12,257-green"> &nbsp;
<img src="https://img.shields.io/badge/dashboards-153-green"> 

Steampipe reads live data from APIs into Postgres. Data sources include the major clouds (AWS, Azure, GCP), security services and tools (PAN-OS, Trivy), business apps (Google Workspace, Salesforce, Slack), infrastructure definitions (CloudFormation, Terraform), and more. 

With [Steampipe](https://steampipe.io) you can:

- **Query** → Write SQL queries that report on -- and join across! -- clouds, business apps, code, logs, and more.

- **Check** → Verify that your cloud resources comply with frameworks such as CIS, NIST, and HIPAA.

- **View** → Explore query results on dashboards.

## Steampipe CLI: The console for API queries

Steampipe provides a growing suite of [plugins](https://hub.steampipe.io) that map APIs to Postgres tables. The [interactive query shell](https://steampipe.io/docs/query/query-shell) is one way you can query those tables. 

<br/>

<img marginTop="200" width="524" src="https://steampipe.io/images/steampipe-sql-demo.gif" />

<br/>

You can also use psql, pgcli, Metabase, Tableau, or any client can connect to Postgres.

### Get started with the CLI

[Install](https://steampipe.io/downloads) Steampipe.
```
sudo /bin/sh -c "$(curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh)"
```

Add your first [plugin](https://hub.steampipe.io/plugins).
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

- It's [just SQL](https://steampipe.io/docs/sql/steampipe-sql)!

- You can run queries [on the command line](https://steampipe.io/docs/query/overview#non-interactive-batch-query-mode) and include them in scripts.

- Other [commands](https://steampipe.io/docs/reference/cli/overview) run benchmarks, launch Steampipe as a service, and start the dashboard server.

- [meta-commands](https://steampipe.io/docs/reference/dot-commands/overview) control caching, [environment variables](https://steampipe.io/docs/reference/env-vars/overview), the [search path](https://steampipe.io/docs/guides/search-path), and more.

- Queries can run in [batch mode](https://steampipe.io/docs/query/batch-query).

- You can bundle connections (e.g. for many AWS accounts) using an [aggregator](https://steampipe.io/docs/managing/connections#using-aggregators).

  
## Steampipe mods: Benchmarks and dashboards

Steampipe [mods](https://hub.steampipe.io/mods) build on the query layer. Some run suites of controls that check for compliance with frameworks (e.g. CIS, NIST, GDPR). 

<br/>

![readme-dns-best-practices](https://user-images.githubusercontent.com/46509/184996087-06b9dcaa-8833-4908-8d3d-58d76ac81e0d.gif)

<br/>

Others visualize query results as charts, tables, and [other widgets](https://steampipe.io/docs/reference/mod-resources/overview).

<br/>

![readme-aws-s3-bucket-dashboard](https://user-images.githubusercontent.com/46509/184996162-bf7e3c23-2a5e-4118-a9ba-ba90d3d7ea2c.gif)

<br/>
 
### Get started with benchmarks and dashboards

The [Net Insights](https://hub.steampipe.io/mods/turbot/net_insights) mod works with the Net plugin shown above. To run it, first clone its repo and change to that directory.

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

Note that the default domains are `microsoft.com` and `github.com`. You can [change those defaults](https://hub.steampipe.io/mods/turbot/net_insights#configuration) to check other domains.

### Explore AWS dashboards

The [AWS Insights](https://hub.steampipe.io/mods/turbot/aws_insights) provides dozens of dashboards that exercise the full set of widgets. To explore these dashboards, first install the [AWS plugin](https://hub.steampipe.io/plugins/turbot/aws) and [authenticate](https://hub.steampipe.io/plugins/turbot/aws#configuration).

Then clone `AWS Insights`, change to its directory, launch `steampipe dashboard`, and open `localhost:9194`.

### Learn more about benchmarks and dashboards

 **Benchmarks**

- Use [search paths](https://steampipe.io/docs/check/overview#more-examples) to specify connections.

- Review `steampipe check` [commands](https://steampipe.io/docs/reference/cli/check) to filter on control tags or set  variables.

- Output results in a variety of [formats](https://steampipe.io/docs/reference/cli/check#output-formats).

- Write your own [custom ouput template](https://steampipe.io/docs/develop/writing-control-output-templates).

- Build your own [custom benchmark](https://steampipe.io/docs/mods/overview).

**Dashboards**

- Reveal the [HCL and SQL source code](https://steampipe.io/docs/dashboard/panel) for dashboard panels, and download the data.

- Review [dashboard commands](https://steampipe.io/docs/reference/cli/dashboard) to search paths, dashboard ports, and variables.

- Explore [dashboard widgets](https://steampipe.io/docs/reference/mod-resources/overview#dashboards).

- Build your own [custom dashboard](https://steampipe.io/docs/mods/writing-dashboards).
 
## Community

We thrive on feedback and community involvement!

**Have a question?** → Join our [Slack community](https://steampipe.io/community/join) or open a [GitHub issue](https://github.com/turbot/steampipe/issues/new/choose)

**Want to get involved?** → Learn how to [contribute](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md)

**Want to work with the team?** → We are [hiring](https://turbot.com/careers)!

## Steampipe Cloud

Want a hosted version of Steampipe? Bring your team to [Steampipe Cloud](https://cloud.steampipe.io).  

