![image](https://steampipe.io/images/steampipe-social-preview-4.png)

# Steampipe CLI quick start

- **[Get started →](https://steampipe.io/downloads)**
- Install your favorite [plugins](https://hub.steampipe.io/plugins)
- Documentation: [Table definitions & examples](https://steampipe.io/docs)
- Community: [Slack Channel](https://join.slack.com/t/steampipe/shared_invite/zt-oij778tv-lYyRTWOTMQYBVAbtPSWs3g)
- Get involved: [Issues](https://github.com/turbot/steampipe/issues)

# Developing

Prerequisites:

- [Golang](https://golang.org/doc/install) Version 1.16 or higher.

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
steampipe version 0.7.0
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

## Contributing

Please see the [contribution guidelines](https://github.com/turbot/steampipe/blob/main/CONTRIBUTING.md) and our [code of conduct](https://github.com/turbot/steampipe/blob/main/CODE_OF_CONDUCT.md). All contributions are subject to the [AGPLv3 open source license](https://github.com/turbot/steampipe-plugin-shodan/blob/main/LICENSE).

Guides:

- [Writing plugins](https://steampipe.io/docs/develop/writing-plugins)
- [Writing your first table](https://steampipe.io/docs/develop/writing-your-first-table)

`help wanted` issues:

- [Steampipe](https://github.com/turbot/steampipe/labels/help%20wanted)
- [Plugin Repos](https://github.com/topics/steampipe-plugin)
