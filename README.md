[<img width="524px" src="https://steampipe.io/images/steampipe_logo_wordmark_color.svg" />](https://steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

[![plugins](https://img.shields.io/badge/plugins-83-lightblue)](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp; [![mods](https://img.shields.io/badge/controls-3,000-lightblue)](https://hub.steampipe.io/mods?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp; [![dashboards](https://img.shields.io/badge/dashboards-744-lightblue)](https://hub.steampipe.io/mods?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp; [![slack](https://img.shields.io/badge/slack_community-825-lightblue)](https://steampipe.io/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp; [![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-gold)](https://turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

Steampipe is the universal API for cloud. Query across infrastructure, SaaS, code, logs, and more. SQL is the common interface to data everywhere.

With [Steampipe](https://steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) you can:

- **Query** → Write SQL queries that report on (and join across!) clouds, SaaS, code, logs, and more.

- **Report** → Verify that your cloud resources comply with security benchmarks such as CIS, NIST, SOC2, etc.

- **Visualize** → Leverage a huge library of interactive dashboards, or build your own dashboards as code.

- **Collaborate** → Clone. Edit. Query. Steampipe's design promotes community and sharing. 

## Steampipe CLI - The SQL console for API queries
 
Steampipe's community have grown a suite of [plugins](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) that map APIs to Postgres tables. There are plugins for:
- **Public Cloud:** AWS, Azure, GCP, Oracle, Alibaba, IBM, etc
- **SaaS:** Airtable, GitHub, Slack, Stripe, Zoom, Salesforce, etc
- **Containers:** Docker, Kubernetes, Prometheus, Grafana, etc
- **Security:** CrowdStrike, VirusTotal, Shodan, PAN-OS, etc
- **Identity:** Azure AD, LDAP, Duo, Google Directory, etc
- **Web:** Cloudflare, Gandi, updown.io, WHOIS, Requests, etc 
- **Code:** Terrafrom, Docker, Code Secrets, CSV, etc
- **Logs:** AWS CloudWatch, Splunk, Datadog, Algolia, etc
- **Social:** Twitter, Reddit, HackerNews, RSS, etc
- **Custom:** Build your own [custom plugins](https://steampipe.io/docs/develop/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)
 
The [interactive query shell](https://steampipe.io/docs/query/query-shell?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) is one way you can query those tables. 

<img marginTop="200" width="524" src="https://steampipe.io/images/steampipe-sql-demo.gif" />

You can also use psql, pgcli, Metabase, Tableau, or [any client](https://steampipe.io/docs/cloud/integrations/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) that can connect to Postgres.

### Get started with the CLI

[Install](https://steampipe.io/downloads?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) Steampipe for:

Linux or WSL

```sh
sudo /bin/sh -c "$(curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh)"
```

MacOS

```sh
brew tap turbot/tap
brew install steampipe
```

Add your first [plugin](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

```sh
steampipe plugin install net
```

Run `steampipe query` to launch the interactive shell.

```sh
steampipe query
```

Run your first query!

```sql
select
  *
from
  net_certificate
where
  domain = 'google.com';
```

### Learn more about the CLI

- It's [just SQL](https://steampipe.io/docs/sql/steampipe-sql?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)!

- You can run queries [on the command line](https://steampipe.io/docs/query/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#non-interactive-batch-query-mode) and include them in scripts.

- Other [commands](https://steampipe.io/docs/reference/cli/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) run benchmarks, launch Steampipe as a service, and start the dashboard server.

- [Meta-commands](https://steampipe.io/docs/reference/dot-commands/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) control caching, [environment variables](https://steampipe.io/docs/reference/env-vars/overview), the [search path](https://steampipe.io/docs/guides/search-path?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme), and more.

- Queries can run in [batch mode](https://steampipe.io/docs/query/batch-query?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

- You can bundle connections (e.g. for many AWS accounts) using an [aggregator](https://steampipe.io/docs/managing/connections?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#using-aggregators).
  
## Steampipe Mods - Developer centric benchmarks & dashboards

Steampipe's community have grown a suite of [mods](https://hub.steampipe.io/mods?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) which are collections of named `queries`, `benchmarks` & `controls` that can be used to test current configurations of your cloud resources, and `dashboards` that organize and display key pieces of information. Within Steampipe Mods there are a huge library of interactive dashboards, and thousands of industry benchmarks to assess your compliance, security and cost controls.  There are mods for:

- **Insights:** view dashboards and reports for your resources across AWS, GCP, Kubernetes, etc 
- **Compliance:** assess compliance for HIPAA, PCI, GxP and more across AWS, Azure, etc 
- **Security:** run security benchmarks for CIS, NIST, FedRAMP, and more across AWS, OCI, Terraform, etc.
- **Tagging:** review tagging controls across all your AWS, Azure and GCP accounts
- **Cost:** check for unused and under utilized resources across AWS, OCI, Digital Ocean, etc.
- **Custom:** Build your [custom mods, benchmarks & dashboards](https://steampipe.io/docs/mods/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)


Steampipe controls and benchmarks provide a generic mechanism for defining and running compliance, security, tagging, and cost controls, as well as your own customized groups of controls.


![readme-aws-cis-1 4](https://user-images.githubusercontent.com/46509/185204883-54311f57-759d-410f-92bb-d1e92373a35b.gif)

<br/>

Mods also provide dashboards that report status, display charts and tables, and visualize relationships among resources.  

![aws_s3_bucket_dashboard](https://user-images.githubusercontent.com/17007758/185409103-4eeaccd7-29e6-415c-94f7-dcab01a351c0.png)

Steampipe's [dashboards-as-code](https://steampipe.io/blog/dashboards-as-code?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) approach enables developers to extend these dashboards, and create their own, using SQL to gather data and HCL to flow the data into [widgets](https://steampipe.io/docs/reference/mod-resources/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

### Get started with benchmarks and dashboards

The [Net Insights](https://hub.steampipe.io/mods/turbot/net_insights?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) mod works with the Net plugin shown above. To run it, first clone its repo and change to that directory.

```sh
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

Results can be outputed into a variety of [formats](https://steampipe.io/docs/reference/cli/check?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#output-formats) such as JSON, CSV, HTML, etc. [Custom ouput templates](https://steampipe.io/docs/develop/writing-control-output-templates?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) can be created as well.


### Run benchmarks as dashboards

Launch the dashboard server: `steampipe dashboard`

Open `http://localhost:9194` in your browser. The home page lists available dashboards. Click `DNS Best Practices` to view that dashboard.

Note that the default domains are `microsoft.com` and `github.com`. You can [change those defaults](https://hub.steampipe.io/mods/turbot/net_insights#configuration?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) to check other domains.

### Explore query results on dashboards

Dashboards use charts, tables, and interactive [widgets](https://steampipe.io/docs/reference/mod-resources/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#dashboards) to help you explore and visualize your resources. The [AWS Insights](https://hub.steampipe.io/mods/turbot/aws_insights?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme), for example, provides dozens of dashboards that exercise the full set of widgets. To explore these dashboards, first install the [AWS plugin](https://hub.steampipe.io/plugins/turbot/aws?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) and [authenticate](https://hub.steampipe.io/plugins/turbot/aws?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#configuration).

Then clone `AWS Insights`, change to its directory, launch `steampipe dashboard`, and open `localhost:9194`.

## Community

We thrive on feedback and community involvement!

**Have a question?** → Join our [Slack community](https://steampipe.io/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) or open a [GitHub issue](https://github.com/turbot/steampipe/issues/new/choose).

**Want to get involved?** → Learn how to [contribute](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md).

**Want to work with the team?** → We are [hiring](https://turbot.com/careers?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)!

## Steampipe Cloud

Want a hosted version of Steampipe? Bring your team to [Steampipe Cloud](https://cloud.steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).  

