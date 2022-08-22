[<picture><source media="(prefers-color-scheme: dark)" srcset="https://cloud.steampipe.io/images/steampipe-logo-wordmark-white.svg"><source media="(prefers-color-scheme: light)" srcset="https://cloud.steampipe.io/images/steampipe-logo-wordmark-color.svg"><img width="67%" alt="Steampipe Logo" src="https://cloud.steampipe.io/images/steampipe-logo-wordmark-color.svg"></picture>](https://steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

[![plugins](https://img.shields.io/badge/apis_supported-83-gold)](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp; 
[![benchmarks](https://img.shields.io/badge/benchmarks-2K-gold)](https://hub.steampipe.io/mods?objectives=compliance?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp;
[![dashboards](https://img.shields.io/badge/dashboards-320-gold)](https://hub.steampipe.io/mods?objectives=dashboard?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp;
[![slack](https://img.shields.io/badge/slack-825-gold)](https://steampipe.io/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) &nbsp;
[![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-gold)](https://turbot.com?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)

Steampipe is the universal interface to APIs. Use SQL to query cloud infrastructure, SaaS, code, logs, and more. 

With [Steampipe](https://steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) you can:

- **Query** → Use SQL to [query](https://steampipe.io/docs/query/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) (and join across!) [APIs](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).

- **Check** → Ensure that cloud resources comply with [security benchmarks](https://steampipe.io/docs/check/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) such as CIS, NIST, and SOC2.

- **Visualize** → View [prebuilt dashboards](https://hub.steampipe.io/mods?objectives=dashboard&utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) or [build your own](https://steampipe.io/docs/mods/writing-dashboards?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).
 

## Steampipe CLI - The SQL console for API queries

The Steampipe community has grown a suite of [plugins](https://hub.steampipe.io/plugins?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) that map APIs to tables. 
<table>
  <tr>
   <td><b>Cloud</b></td>
   <td>AWS, Alibaba, Azure, GCP, IBM, Oracle …</td>
  </tr>
  <tr>
   <td><b>SaaS</b></td>
   <td>Airtable, Jira, GitHub, Google Workspace, Salesforce, Slack, Stripe, Zoom …</td>
  </tr>
  <tr>
   <td><b>Security</b></td>
   <td>CrowdStrike, PAN-OS, VirusTotal, Shodan, Trivy …</td>
  </tr>
  <tr>
   <td><b>Identity</b></td>
   <td>Azure AD, Duo, Keycloack, Google Directory, LDAP …</td>
  </tr>
  <tr>
   <td><b>DevOps</b></td>
   <td>Docker, Grafana, Kubernetes, Prometheus …</td>
  </tr>
  <tr>
   <td><b>Net</b></td>
   <td>Baleen, Cloudflare, crt.sh, Gandi, IMAP, ipstack, updown.io, WHOIS …</td>
  </tr>
  <tr>
   <td><b>IaC</b></td>
   <td>CloudFormation, Terraform …</td>
  </tr>
  <tr>
   <td><b>Logs</b></td>
   <td>Algolia, AWS CloudWatch, Splunk, Datadog …</td>
  </tr>
  <tr>
   <td><b>Social</b></td>
   <td>HackerNews, Twitter, Reddit, RSS …</td>
  </tr>
  <tr>
   <td><b>Your API</b></td>
   <td>Build your own <a href="https://steampipe.io/docs/develop/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">custom plugins</a>.</td>
  </tr>
 </table>
  

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
 
 Choose a plugin from the [hub](https://hub.steampipe.io), for example: [Net](https://hub.steampipe.io/plugins/turbot/net).

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
  
## Steampipe Mods - Benchmarks & dashboards

The Steampipe community has also grown a suite of [mods](https://hub.steampipe.io/mods?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) which are sets of **benchmarks** that check your cloud resources for compliance, and **dashboards** that visualize your resources.


<table>
  <tr>
   <td><b>Compliance</b></td>
   <td>Check AWS, Azure, GCP, and other clouds for compliance with HIPAA, PCI, GxP and other standards
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
   <td>Use CIS, NIST, FedRAMP , and other benchmarks to asses the security of AWS, Azure, GCP, and other clouds</td>
  </tr>
  <tr>
   <td><b>Tagging</b></td>
   <td>Verify the consistency of tags applied to AWS, Azure, and GCP resources</td>
  </tr>
  <tr>
   <td><b>Your own mod</b></td>
   <td>Build your own <a href="https://steampipe.io/docs/mods/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">benchmarks and dashboards</a></td>.
  </tr>
 </table>



### The CIS 1.4 benchmark in `AWS Compliance`

![gh-readme-cis-benchmark-in-dashboard](https://user-images.githubusercontent.com/46509/186024940-7ae9f42f-241b-44a7-84d6-244f5d488e1f.gif)
 
### The AWS EC2 Instance dashboard in `AWS Insights`
 
 ![aws-ec2-dashboard-in-cloud](https://user-images.githubusercontent.com/46509/186023273-d2be66c9-c050-4576-a46b-ed3f82f2e14a.jpg)

 
 </details>
 

Benchmarks and dashboards use SQL to gather data and HCL to flow the data into [benchmark controls](https://steampipe.io/blog/release-0-11-0#composable-mods) and  [dashboard widgets](https://steampipe.io/blog/dashboards-as-code). You can use the existing suites of benchmarks and dashboards, or build derivative versions, or create your own. 
### Get started with benchmarks and dashboards

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
</details>

<details>
<summary>Capture output</summary>
<br/>
Available <a href="https://steampipe.io/docs/reference/cli/check?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#output-formats">formats</a> include JSON, CSV, HTML, and ASFF. 

You can use <a href="https://steampipe.io/docs/develop/writing-control-output-templates?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">custom ouput templates</a> to create new output formats
</details>

<details>
<summary>Run benchmarks as dashboards</summary>

Launch the dashboard server: `steampipe dashboard`.

Open `http://localhost:9194` in your browser. 

The home page lists available dashboards. Click `DNS Best Practices` to view that dashboard.

Note that the default domains are `microsoft.com` and `github.com`. You can <a href="https://hub.steampipe.io/mods/turbot/net_insights#configuration?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">change those defaults</a> to check other domains.
</details>

<details>
<summary>Explore your resources</summary>
<br/>
Dashboards use charts, tables, and interactive <a href="https://steampipe.io/docs/reference/mod-resources/overview?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#dashboards">widgets</a> to help you explore and visualize your resources. 

The <a href="https://hub.steampipe.io/mods/turbot/aws_insights?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">AWS Insights</a> mod, for example, provides dozens of dashboards that exercise the full set of widgets. To use these dashboards, first install the <a href="https://hub.steampipe.io/plugins/turbot/aws?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme">AWS plugin</a> and <a href="https://hub.steampipe.io/plugins/turbot/aws?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme#configuration">authenticate</a>. Then clone `AWS Insights`, change to its directory, launch `steampipe dashboard`, and open `localhost:9194`.
</details>


## Community

We thrive on feedback and community involvement!

**Have a question?** → Join our [Slack community](https://steampipe.io/community/join?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme) or open a [GitHub issue](https://github.com/turbot/steampipe/issues/new/choose).

**Want to get involved?** → Learn how to [contribute](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md).

**Want to work with the team?** → We are [hiring](https://turbot.com/careers?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme)!

## Steampipe Cloud

Want a hosted version of Steampipe? Bring your team to [Steampipe Cloud](https://cloud.steampipe.io?utm_id=gspreadme&utm_source=github&utm_medium=repo&utm_campaign=github&utm_content=readme).  

