<img width="524px" src="https://steampipe.io/images/steampipe_logo_wordmark_color.svg" />

[![plugins](https://img.shields.io/badge/plugins-83-green)](https://hub.steampipe.io/plugins) &nbsp; [![mods](https://img.shields.io/badge/controls-3000-green)](https://hub.steampipe.io/mods) &nbsp; [![dashboards](https://img.shields.io/badge/dashboards-744-green)](https://hub.steampipe.io/mods) &nbsp; [![slack](https://img.shields.io/badge/slack-800-E01563)](https://steampipe.io/community/join) &nbsp; [![maintained by](https://img.shields.io/badge/maintained%20by-Turbot-gold)](https://turbot.com)

Steampipe from [Turbot](https://turbot.com) exposes APIs and services as a high-performance relational database, giving you the ability to write SQL-based queries to explore dynamic data through [plugins](https://hub.steampipe.io/plugins). [Mods](https://hub.steampipe.io/plugins) extend Steampipe's capabilities with dashboards, reports, and controls built with simple HCL + SQL.

With [Steampipe](https://steampipe.io) you can:

- **Query**, join & report on your cloud, SaaS, containers, code, logs & more in a common SQL interface
- **Visualize** insights of your resource configurations with dashboards
- **Check** for compliance with security benchmarks such as CIS, NIST, HIPAA, and with DevSecOps best practices

## Steampipe CLI - plugins & SQL

Steampipe provides an [interactive query shell](https://steampipe.io/docs/query/query-shell) that provides features like auto-complete, syntax highlighting, and command history to assist you in writing queries.

<img width="524" src="https://steampipe.io/images/steampipe-sql-demo.gif" />

**To get started:**

**[Install Steampipe](https://steampipe.io/downloads) in your terminal:** *(example on Linux)*
```
sudo /bin/sh -c "$(curl -fsSL https://raw.githubusercontent.com/turbot/steampipe/main/install.sh)"
```

**Add your first [plugins](https://hub.steampipe.io/plugins):** *(example to install the Net plugin)*
```
steampipe plugin install net
```

**Open the CLI, run `steampipe query` with no arguments:**
```
steampipe query
```

**Run your first query!**
```
select
  *
from
  net_certificate
where
  domain = 'google.com';
```

What's next? Install more [plugins](https://hub.steampipe.io/plugins), test additional queries, try out [Mods](https://hub.steampipe.io/mods) (more info below).

<details>
  <summary><b>More information about Steampipe CLI</b></summary>
 
- It's just SQL -- [refresh on SQL basics](https://steampipe.io/docs/sql/steampipe-sql)
- You can also run queries in [non-interactive mode](https://steampipe.io/docs/query/overview#non-interactive-batch-query-mode) in your terminal e.g. `steampipe query "select * from aws_account;" `
- Learn more about [Steampipe CLI commands](https://steampipe.io/docs/reference/cli/overview), [meta-commands](https://steampipe.io/docs/reference/dot-commands/overview) for caching & more, setting [environment variables](https://steampipe.io/docs/reference/env-vars/overview), and other [configs](https://steampipe.io/docs/reference/config-files/overview).
- Run [multiple SQL queries](https://steampipe.io/docs/query/batch-query) from `.sql` files
- Steampipe by default queries APIs live, however you can enable [service mode](https://steampipe.io/docs/managing/service) for a Postgres endpoint to connect 3rd party tools.
- In service mode or through [Steampipe Cloud](https://cloud.steampipe.io) you can [connect to Steampipe](https://steampipe.io/docs/cloud/integrations/overview) with any SQL IDE or BI tool
- Get familiar with your [plugin connection configurations](https://steampipe.io/docs/managing/connections) -- learn configure multiple connections and setup aggegators
- Tip and tricks for [managing multiple connections & search_paths](https://steampipe.io/docs/guides/search-path) 
</details>
  
## Steampipe Mods - benchmarks & dashboards

While Steampipe plugins provide an easy way to query your resources, Steampipe [mods](https://hub.steampipe.io/mods) are collections of named queries, cofified controls, and dashboards that organize and display key pieces of information.

(screencast)

<img width="524" src="https://github.com/turbot/steampipe-mod-net-insights/blob/main/docs/images/net_dns_best_practices_dashboard.png" />

**To get started:**

**Download a [Mod](https://hub.steampipe.io/mods) and view its dashboards:** *([Net plugin](https://hub.steampipe.io/mods/turbot/net_insights) example)*
```
git clone https://github.com/turbot/steampipe-mod-net-insights.git
```

**Change to that directory and run `steampipe dashboard`:**
```
cd steampipe-mod-net-insights
steampipe dashboard
```

**View your first Dashboard:**

Steampipe will load the embedded web server on port 9194 and open `http://localhost:9194` in your browser. 

The home page lists the available dashboards, and is searchable by title or tags. Click the `DNS Best Practices` to view your first dashboard. 

*Note that the default domains checked are `microsoft.com` and `github.com`. To check different domains, [configure the variables](https://hub.steampipe.io/mods/turbot/net_insights#configuration).*

**Run individual controls from your terminal:**

Instead of running benchmarks in a dashboard, you can also run them within your terminal with the `steampipe check` command:

Run all benchmarks:

```sh
steampipe check all
```

Run a single benchmark:

```sh
steampipe check benchmark.dns_best_practices
```

Run a specific control:

```sh
steampipe check control.dns_ns_name_valid
```

Different output formats are also available, for more information please see
[Output Formats](https://steampipe.io/docs/reference/cli/check#output-formats).

<details>
  <summary><b>More information about Steampipe Mods</b></summary>


</details>

## Community

We thrive on feedback and community involvement!

**Have a question?** → Join our [Slack community](https://steampipe.io/community/join) or open a [GitHub issue](https://github.com/turbot/steampipe/issues/new/choose)

**Want to get involved?** → Learn how to [contribute](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md)

**Want to work with the team?** → We are [hiring](https://turbot.com/careers)!

## Steampipe Cloud

Want a hosted version of Steampipe? Bring your team to [Steampipe Cloud](https://cloud.steampipe.io).  

