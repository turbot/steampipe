![](https://steampipe.io/images/steampipe_logo_wordmark_color.svg)

With Steampipe you can query your favorite cloud APIs using PostgreSQL, check your AWS/Azure/GCP/Kubernetes infrastructure for compliance with security frameworks, and visualize data in a whole new way.

![](./steampipe-original-screencast.png)

> Note:to make it come alive, open in github editor, drag-drop steampipe-original-screencast.mp4]

## Table of Contents

[Steampipe CLI quick start](#steampipe-cli-quick-start)

[Steampipe architecture](#steampipe-architecture)

[For developers](#for-developers)

[Compliance benchmarks](#compliance-benchmarks)

[By the numbers](#by-the-numbers)


# Steampipe CLI quick start

- **[Get started →](https://steampipe.io/downloads)**

- Install your favorite [plugins](https://hub.steampipe.io/plugins)

- Documentation: [Table definitions & examples](https://steampipe.io/docs)

- Community: [Slack Channel](https://join.slack.com/t/steampipe/shared_invite/zt-oij778tv-lYyRTWOTMQYBVAbtPSWs3g)


# Steampipe architecture

![steampipe architecture](./architecture.png)

# For developers

## Writing queries and controls

Prerequisites: none

Links:

 - [Writing queries](https://steampipe.io/docs/writing-queries)

 - [Writing controls](https://steampipe.io/docs/using-steampipe/writing-controls)

## Developing a plugin

Prerequisites:

- [Golang](https://golang.org/doc/install) Version 1.17 or higher.

Links:

- [Writing plugins](https://steampipe.io/docs/develop/writing-plugins)

- [Writing your first table](https://steampipe.io/docs/develop/writing-your-first-table)

## Steampipe developers

Prerequisites:

- [Golang](https://golang.org/doc/install) Version 1.17 or higher.

Clone:

```sh
git clone git@github.com:turbot/steampipe
cd steampipe
```

Build, which automatically installs the new version to your `/usr/local/bin/steampipe` directory:

```
make
```

Check the version

```
$ steampipe -v
steampipe version 0.10.0
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

## For all contributors

Please see the [contribution guidelines](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md) and our [code of conduct](https://github.com/turbot/steampipe/blob/main/CODE_OF_CONDUCT.md). All contributions are subject to the [AGPLv3 open source license](https://github.com/turbot/steampipe-plugin-shodan/blob/main/LICENSE).

# Compliance benchmarks

[AWS](https://hub.steampipe.io/mods/turbot/aws_compliance): Audit Manager Control Tower, AWS Foundational Security Best Practices, CIS, GDPR, HIPAA, NIST 800-53, NIST CSF, PCI DSS, RBI Cyber Security Framework and SOC 2.

[Azure](https://hub.steampipe.io/mods/turbot/azure_compliance): CIS, HIPAA HITRUST and NIST

[GCP](https://hub.steampipe.io/mods/turbot/gcp_compliance): CIS, Forseti Security and CFT Scorecard

[Kubernetes](https://hub.steampipe.io/mods/turbot/kubernetes_compliance): NSA and CISA Kubernetes Hardening Guidance

# By the numbers (Feb 2022)

APIs supported by [plugins](https://hub.steampipe.io/plugins) in the hub: 62

[Mods](https://hub.steampipe.io/mods) available in the hub: 21

Compliance benchmarks for [AWS](https://hub.steampipe.io/mods/turbot/aws_compliance): 11, [Azure](https://hub.steampipe.io/mods/turbot/azure_compliance): 3, [GCP](https://hub.steampipe.io/mods/turbot/gcp_compliance): 3, [Kubernetes](https://hub.steampipe.io/mods/turbot/kubernetes_compliance): 2

Named resources available on the hub: controls: x000, queries: y000.
