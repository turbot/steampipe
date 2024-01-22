[<picture><source media="(prefers-color-scheme: dark)" srcset="https://steampipe.io/images/steampipe-color-logo-and-wordmark-with-white-bubble.svg"><source media="(prefers-color-scheme: light)" srcset="https://steampipe.io/images/steampipe-color-logo-and-wordmark-with-white-bubble.svg"><img width="67%" alt="Steampipe Logo" src="https://steampipe.io/images/steampipe-color-logo-and-wordmark-with-white-bubble.svg"></picture>](https://steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

[![plugins](https://img.shields.io/badge/apis_supported-140-blue)](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp; 
[![benchmarks](https://img.shields.io/badge/controls-5873-blue)](https://hub.steampipe.io/mods?objectives=compliance?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp;
[![dashboards](https://img.shields.io/badge/dashboards-710-blue)](https://hub.steampipe.io/mods?objectives=dashboard?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp;
[![slack](https://img.shields.io/badge/slack-2017-blue)](https://turbot.com/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp;
[![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-blue)](https://turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)


Steampipe is the zero-ETL solution for getting data directly from APIs and services. We offer these Steampipe engines:

- [Steampipe CLI](#steampipe-cli). Query APIs, check compliance, visualize on dashboards.

- [Steampipe Postgres FDWs](https://steampipe.io/docs/steampipe_postgres/index). Use native Postgres Foreign Data Wrappers that translate APIs to foreign tables.

- [Steampipe SQLite extensions](https://steampipe.io/docs/steampipe_sqlite/index). Use SQLite extensions that translate APIS to SQLite virtual tables.

- [Steampipe export tools](https://steampipe.io/docs/steampipe_export/index). Use standalone binaries that export data from APIs, no database required.

- [Turbot Pipes](https://turbot.com/pipes). Query, check, and visualize with your team using the only intelligence, automation & security platform built specifically for DevOps. 

## A common suite of API plugins

The Steampipe community has grown a suite of [plugins](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) that map APIs to tables. They work with all Steampipe engines.

<table>
  <tr>
   <td><b>Cloud</b></td>
   <td><a href="https://hub.steampipe.io/plugins/turbot/aws?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">AWS</a>, <a href="https://hub.steampipe.io/plugins/turbot/alicloud?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Alibaba</a>, <a href="https://hub.steampipe.io/plugins/turbot/azure?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Azure</a>, <a href="https://hub.steampipe.io/plugins/turbot/gcp?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">GCP</a>, <a href="https://hub.steampipe.io/plugins/turbot/ibm?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">IBM</a>, <a href="https://hub.steampipe.io/plugins/turbot/oci?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Oracle</a> …</td>
  </tr>
  <tr>
   <td><b>SaaS</b></td>
   <td><a href="https://hub.steampipe.io/plugins/francois2metz/airtable?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Airtable</a>, <a href="https://hub.steampipe.io/plugins/turbot/jira?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Jira</a>, <a href="https://hub.steampipe.io/plugins/turbot/github?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">GitHub</a>, <a href="https://hub.steampipe.io/plugins/turbot/googleworkspace?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Google Workspace</a>, <a href="https://hub.steampipe.io/plugins/turbot/microsoft365?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Microsoft 365</a>, <a href="https://hub.steampipe.io/plugins/turbot/salesforce?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Salesforce</a>, <a href="https://hub.steampipe.io/plugins/turbot/slack?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Slack</a>, <a href="https://hub.steampipe.io/plugins/turbot/stripe?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Stripe</a>, <a href="https://hub.steampipe.io/plugins/turbot/zoom?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Zoom</a> …</td>
  </tr>
  <tr>
   <td><b>Security</b></td>
   <td><a href="https://hub.steampipe.io/plugins/turbot/crowdstrike?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">CrowdStrike</a>, <a href="https://hub.steampipe.io/plugins/francois2metz/gitguardian?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">GitGuardian</a>, <a href="https://hub.steampipe.io/plugins/turbot/hibp?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Have I Been Pwned</a>, <a href="https://hub.steampipe.io/plugins/turbot/panos?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">PAN-OS</a>, <a href="https://hub.steampipe.io/plugins/turbot/shodan?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Shodan</a>, <a href="https://hub.steampipe.io/plugins/turbot/trivy?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Trivy</a>, <a href="https://hub.steampipe.io/plugins/turbot/virustotal?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">VirusTotal</a> …</td>
  </tr>
  <tr>
   <td><b>Identity</b></td>
   <td><a href="https://hub.steampipe.io/plugins/turbot/azuread?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Azure AD</a>, <a href="https://hub.steampipe.io/plugins/turbot/duo?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Duo</a>, <a href="https://hub.steampipe.io/plugins/theapsgroup/keycloak?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Keycloak</a>, <a href="https://hub.steampipe.io/plugins/turbot/googledirectory?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Google Directory</a>, <a href="https://hub.steampipe.io/plugins/turbot/ldap?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">LDAP</a> …</td>
  </tr>
  <tr>
   <td><b>DevOps</b></td>
   <td><a href="https://hub.steampipe.io/plugins/turbot/docker?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Docker</a>, <a href="https://hub.steampipe.io/plugins/turbot/grafana?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Grafana</a>, <a href="https://hub.steampipe.io/plugins/turbot/kubernetes?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Kubernetes</a>, <a href="https://hub.steampipe.io/plugins/turbot/prometheus?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Prometheus</a> …</td>
  </tr>
  <tr>
   <td><b>Net</b></td>
   <td><a href="https://hub.steampipe.io/plugins/francois2metz/baleen?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Baleen</a>, <a href="https://hub.steampipe.io/plugins/turbot/cloudflare?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Cloudflare</a>, <a href="https://hub.steampipe.io/plugins/turbot/crtsh?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">crt.sh</a>, <a href="https://hub.steampipe.io/plugins/francois2metz/gandi?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Gandi</a>, <a href="https://hub.steampipe.io/plugins/turbot/imap?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">IMAP</a>, <a href="https://hub.steampipe.io/plugins/turbot/ipstack?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">ipstack</a>, <a href="https://hub.steampipe.io/plugins/turbot/tailscale?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Tailscale</a>, <a href="https://hub.steampipe.io/plugins/turbot/updown?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">updown.io</a>, <a href="https://hub.steampipe.io/plugins/turbot/whois?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">WHOIS</a> …</td>
</tr>
<tr>
   <td><b>IaC</b></td>
   <td><a href="https://hub.steampipe.io/plugins/turbot/awscfn?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">CloudFormation</a>, <a href="https://hub.steampipe.io/plugins/turbot/terraform?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Terraform</a> …</td>
</tr>
<tr>
   <td><b>Logs</b></td>
   <td><a href="https://hub.steampipe.io/plugins/turbot/algolia?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Algolia</a>, <a href="https://hub.steampipe.io/plugins/turbot/aws/tables/aws_cloudwatch_log_event?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">AWS CloudWatch</a>, <a href="https://hub.steampipe.io/plugins/turbot/datadog?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Datadog</a>, <a href="https://hub.steampipe.io/plugins/turbot/splunk?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Splunk</a> …</td>
  </tr>
  <tr>
   <td><b>Social</b></td>
   <td><a href="https://hub.steampipe.io/plugins/turbot/hackernews?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">HackerNews</a>, <a href="https://hub.steampipe.io/plugins/turbot/twitter?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Twitter</a>, <a href="https://hub.steampipe.io/plugins/turbot/reddit?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Reddit</a>, <a href="https://hub.steampipe.io/plugins/turbot/rss?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">RSS</a> …</td>
  </tr>
  <tr>
   <td><b>Your API</b></td>
   <td>Build your own <a href="https://steampipe.io/docs/develop/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">custom plugins</a></td>
  </tr>
</table>
  

## Steampipe CLI

With [Steampipe CLI](https://steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) you can:

- **Query** → Use SQL to [query](https://steampipe.io/docs/query/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) (and join across!) [APIs](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

- **Check** → Ensure that cloud resources comply with [security benchmarks](https://steampipe.io/docs/check/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) such as CIS, NIST, and SOC2.

- **Visualize** → View [prebuilt dashboards](https://steampipe.io/docs/dashboard/overview?objectives=dashboard&utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) or [build your own](https://steampipe.io/docs/mods/writing-dashboards?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).
 


The [interactive query shell](https://steampipe.io/docs/query/query-shell?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) is one way you can query those tables. 

<img width="524" src="https://steampipe.io/images/steampipe-sql-demo.gif" />

You can also use psql, pgcli, Metabase, Tableau, or [any client](https://steampipe.io/docs/cloud/integrations/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) that can connect to Postgres.

### Get started with the CLI

<details>

 <summary>Install Steampipe</summary>
 <br/>

 The <a href="https://steampipe.io/downloads?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">downloads</a> page shows you how but tl;dr:
 
Linux or WSL

```sh
sudo /bin/sh -c "$(curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh)"
```

MacOS

```sh
brew tap turbot/tap
brew install steampipe
```

</details>

<details>
 <summary>Add a plugin</summary>
 <br>
 
 Choose a plugin from the [hub](https://hub.steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme), for example: [Net](https://hub.steampipe.io/plugins/turbot/net?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

 Run the `steampipe plugin` command to install it.

```sh
steampipe plugin install net
```
 </details>
 
 <details>
 <summary>Run <tt>steampipe query</tt></summary>
<br/>
Launch the interactive shell.

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
</details>

<details>
 <summary>Learn more about the CLI</summary>

- It's [just SQL](https://steampipe.io/docs/sql/steampipe-sql?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)!

- You can run queries [on the command line](https://steampipe.io/docs/query/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#non-interactive-batch-query-mode) and include them in scripts.

- Other [commands](https://steampipe.io/docs/reference/cli/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) run benchmarks, launch Steampipe as a service, and start the dashboard server.

- [Meta-commands](https://steampipe.io/docs/reference/dot-commands/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) control caching, [environment variables](https://steampipe.io/docs/reference/env-vars/overview), the [search path](https://steampipe.io/docs/guides/search-path?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme), and more.

- Queries can run in [batch mode](https://steampipe.io/docs/query/batch-query?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

- You can bundle connections (e.g. for many AWS accounts) using an [aggregator](https://steampipe.io/docs/managing/connections?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#using-aggregators).
 
 </details>

 <details>
 <summary>Build and develop the CLI</summary>

Prerequisites:

- [Golang](https://golang.org/doc/install) Version 1.19 or higher.

Clone:

```sh
git clone git@github.com:turbot/steampipe
cd steampipe
```

Build, which automatically installs the new version to your `/usr/local/bin/steampipe` directory:

```
make
```

Check the verison

```
$ steampipe -v
steampipe version 0.18.1
```

Install a plugin

```
$ steampipe plugin install steampipe
```

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
  
### Steampipe Mods: Dashboards and benchmarks

The Steampipe community has also grown a suite of [mods](https://hub.steampipe.io/mods?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) which are sets of **dashboards** that visualize your resources and **benchmarks** that check your cloud resources for compliance.

<table>
  <tr>
   <td><b>Compliance</b></td>
   <td>Check AWS, Azure, GCP, etc for compliance with HIPAA, PCI, etc
  </tr>
  <tr>
   <td><b>Cost</b></td>
   <td>Review what AWS, Azure, GCP, and other clouds are costing you</td>
  </tr>
  <tr>
   <td><b>Insights</b></td>
   <td>Visualize cloud resources with charts, tables, and interactive widgets</td>
  </tr>
  <tr>
   <td><b>Security</b></td>
   <td>Use CIS, NIST, FedRAMP etc to assess the security of AWS, Azure, GCP, etc</td>
  </tr>
  <tr>
   <td><b>Tags</b></td>
   <td>Verify the consistency of tags applied to AWS, Azure, and GCP resources</td>
  </tr>
  <tr>
   <td><b>Your mod</b></td>
   <td>Build your own <a href="https://steampipe.io/docs/mods/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">benchmarks and dashboards</a></td>
  </tr>
 </table>
<!--
<table>
  <tr>
   <td><b><a href="https://hub.steampipe.io/mods?objectives=compliance?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Compliance</a></b></td>
   <td>Check AWS, Azure, GCP, etc for compliance with HIPAA, PCI, etc
  </tr>
  <tr>
   <td><b><a href="https://hub.steampipe.io/mods?objectives=cost?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Cost</a></b></td>
   <td>Review what AWS, Azure, GCP, and other clouds are costing you</td>
  </tr>
  <tr>
   <td><b><a href="https://hub.steampipe.io/mods?objectives=dashboard?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Insights</a></b></td>
   <td>Visualize cloud resources with charts, tables, and interactive widgets</td>
  </tr>
  <tr>
   <td><b><a href="https://hub.steampipe.io/mods?objectives=security?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Security</a></b></td>
   <td>Use CIS, NIST, FedRAMP etc to assess the security of AWS, Azure, GCP, etc</td>
  </tr>
  <tr>
   <td><b><a href="https://hub.steampipe.io/mods?objectives=tags?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Tags</a></b></td>
   <td>Verify the consistency of tags applied to AWS, Azure, and GCP resources</td>
  </tr>
  <tr>
   <td><b>Your mod</b></td>
   <td>Build your own <a href="https://steampipe.io/docs/mods/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">benchmarks and dashboards</a></td>
  </tr>
 </table>
-->


### Running dashboards and benchmarks

![benchmarks-and-dashboards](https://user-images.githubusercontent.com/46509/193875366-7d10ca8b-601a-4d93-a333-5c62ea86374b.gif)
 
Dashboards and benchmarks use SQL to gather data and HCL to flow the data into [dashboard widgets](https://steampipe.io/blog/dashboards-as-code?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) and [benchmark controls](https://steampipe.io/blog/release-0-11-0?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#composable-mods). You can use the existing suites of benchmarks and dashboards, or build derivative versions, or create your own. 

### Get started with dashboards and benchmarks

<details>
<summary>Install the Net Insights mod</summary>
<br/>
The <a href="https://hub.steampipe.io/mods/turbot/net_insights?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">Net Insights</a> mod works with the Net plugin shown above. To run it, first clone its repo and change to that directory.

```sh
git clone https://github.com/turbot/steampipe-mod-net-insights
cd steampipe-mod-net-insights
```
</details>

<details>
<br/>
<summary>Run benchmarks in the CLI</summary>

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
 
Available <a href="https://steampipe.io/docs/reference/cli/check?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#output-formats">formats</a> include JSON, CSV, HTML, and ASFF. 
You can use <a href="https://steampipe.io/docs/develop/writing-control-output-templates?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">custom output templates</a> to create new output formats.
</details>

<details>
<summary>Run benchmarks as dashboards</summary>
<br/>
Launch the dashboard server: `steampipe dashboard`, then open `http://localhost:9194` in your browser. 

The home page lists available dashboards. Click `DNS Best Practices` to view that dashboard.

Note that the default domains are `microsoft.com` and `github.com`. You can <a href="https://hub.steampipe.io/mods/turbot/net_insights?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#configuration">change those defaults</a> to check other domains.
</details>

<details>
<summary>Use dashboards to explore your resources</summary>
<br/>
Dashboards use charts, tables, and interactive <a href="https://steampipe.io/docs/reference/mod-resources/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#dashboards">widgets</a> to help you explore and visualize your resources. 

The <a href="https://hub.steampipe.io/mods/turbot/aws_insights?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">AWS Insights</a> mod, for example, provides dozens of dashboards that exercise the full set of widgets. To use these dashboards, first install the <a href="https://hub.steampipe.io/plugins/turbot/aws?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">AWS plugin</a> and <a href="https://hub.steampipe.io/plugins/turbot/aws?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#configuration">authenticate</a>. Then clone `AWS Insights`, change to its directory, launch `steampipe dashboard`, and open `localhost:9194`.
</details>

## Open Source & Contributing

This repository is published under the [AGPL 3.0](https://www.gnu.org/licenses/agpl-3.0.html) license. Please see our [code of conduct](https://github.com/turbot/.github/blob/main/CODE_OF_CONDUCT.md). Contributors must sign our [Contributor License Agreement](https://turbot.com/open-source#cla) as part of their first pull request. We look forward to collaborating with you!

[Steampipe](https://steampipe.io) is a product produced from this open source software, exclusively by [Turbot HQ, Inc](https://turbot.com). It is distributed under our commercial terms. Others are allowed to make their own distribution of the software, but cannot use any of the Turbot trademarks, cloud services, etc. You can learn more in our [Open Source FAQ](https://turbot.com/open-source).

## Get Involved

**[Join #steampipe on Slack →](https://turbot.com/community/join)**

Want to help but don't know where to start? Pick up one of the `help wanted` issues:
* [Steampipe](https://github.com/turbot/steampipe/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22)

## Turbot Pipes

Want a hosted version of Steampipe? Bring your team to [Turbot Pipes](https://pipes.turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

## License

Steampipe is distributed under [AGPL-3.0](https://github.com/turbot/steampipe/blob/main/LICENSE).

