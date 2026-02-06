## v2.3.5 [2026-02-06]
_Bug fixes_
- Fix autocomplete regression where suggestions disappear when typing a table name after `from `. ([#4928](https://github.com/turbot/steampipe/issues/4928))

_Dependencies_
- Updated `golang.org/x/crypto` package to remediate security vulnerabilities.

## v2.3.4 [2025-12-16]
_Bug fixes_
- Fix database client deadlocks caused by concurrent session map access during connection pool cleanup. ([#4917](https://github.com/turbot/steampipe/issues/4917))

## v2.3.3 [2025-12-15]

**Memory and Resource Management**
- Fix query history memory leak due to unbounded growth. ([#4811](https://github.com/turbot/steampipe/issues/4811))
- Fix unbounded growth in autocomplete suggestions maps. ([#4812](https://github.com/turbot/steampipe/issues/4812))
- Fix goroutine leak in snapshot functionality. ([#4768](https://github.com/turbot/steampipe/issues/4768))

**Context and Synchronization**
- Fix RunBatchSession blocking when initData.Loaded never closes. ([#4781](https://github.com/turbot/steampipe/issues/4781))

**File Operations and Installation**
- Fix atomic write to prevent partial files during export. ([#4718](https://github.com/turbot/steampipe/issues/4718))
- Fix atomic OCI installations to prevent inconsistent states. ([#4758](https://github.com/turbot/steampipe/issues/4758))
- Fix atomic FDW binary replacement. ([#4753](https://github.com/turbot/steampipe/issues/4753))
- Fix disk space validation before OCI installation. ([#4754](https://github.com/turbot/steampipe/issues/4754))

**General Fixes**
- Improved SQL query parameterization in connection state management to prevent SQL injections. ([#4748](https://github.com/turbot/steampipe/issues/4748))
- Increase snapshot row streaming timeout from 5s to 30s. ([#4866](https://github.com/turbot/steampipe/issues/4866))

**Dependencies**
- Updated `containerd` and `crypto` packages to remediate vulnerabilities.

## v2.3.2 [2025-11-03]
_Bug fixes_
- Fix Linux builds by aligning the glibc baseline with supported distros to restore compatibility. ([#4691](https://github.com/turbot/steampipe/issues/4691))

## v2.3.1 [2025-10-31]
_Bug fixes_
- Fix issue where MacOS binaries failed to run due to absolute openssl paths. ([#4679](https://github.com/turbot/steampipe/issues/4679))

## v2.3.0 [2025-10-30]
_Whats new_
- Update database version to PostgreSQL 14.19. ([#4644](https://github.com/turbot/steampipe/issues/4644))

_Bug fixes_
- Fix issue where the truncation message was not showing in batch queries for table output format. ([#4674](https://github.com/turbot/steampipe/issues/4674))
- Improve truncation message for datasets exceeding 10k rows in table output format. ([#4674](https://github.com/turbot/steampipe/issues/4674))

## v2.2.0 [2025-09-24]
_Whats new_
- Add support for using context functions in steampipe connection config. ([#4433](https://github.com/turbot/steampipe/issues/4433))
- Show message during startup indicating whether Steampipe launched its own Postgres or connected to an existing service. ([#4427](https://github.com/turbot/steampipe/issues/4427))

_Bug fixes_
- Fix issue where running `plugin update` was creating the default config file, if it did not exist. ([#4628](https://github.com/turbot/steampipe/issues/4628))
- Fix help message after uninstalling plugins. ([#4483](https://github.com/turbot/steampipe/issues/4483))
- Fix issue where steampipe login was not respecting `PIPES_INSTALL_DIR` env var. ([#4402](https://github.com/turbot/steampipe/issues/4402))

## v2.1.0 [2025-07-09]
_Whats new_
- Compiled with Go 1.24.
- The versioning mechanism has been changed to use GoReleaser for automated version management during the build process.

_Breaking changes_
- The [version](https://pkg.go.dev/github.com/turbot/steampipe@v1.1.4/pkg/version) package, which was previously used to control CLI versioning, has been removed in this version. This change only affects users who were importing the Steampipe version package in their Go code. Regular CLI usage is not impacted.

_Bug fixes_
- Bump module to v2. ([#4593](https://github.com/turbot/steampipe/issues/4593))

_Dependencies_
- Update `go-viper` package to remediate moderate vulnerabilities.

## v2.0.1 [2025-06-11]
_Bug fixes_
- Fix `plugin manager is not running` error when starting steampipe via a symlink. ([#4573](https://github.com/turbot/steampipe/issues/4573))

## v2.0.0 [2025-06-11]
_Breaking changes_
- Increased the minimum required `glibc` version to `2.34` for the FDW, due to the upgrade of the Linux build environment from Ubuntu 20.04 to Ubuntu 22.04 GitHub runners. As a result, Steampipe no longer supports older Linux distributions such as Ubuntu 20.04 and Amazon Linux 2.

_Bug fixes_
- Fix issue where the FDW did not correctly provide planning cost information for key-columns with an `any-of` requirement. This led the Postgres planner to choose query plans that do not include filters on those columns, even when filters were present in the query. ([#558](https://github.com/turbot/steampipe-postgres-fdw/issues/558))
- Fix issue where Steampipe was returning a 0 exit code even when a wrong sub-command was run. ([#4563](https://github.com/turbot/steampipe/issues/4563))

## v1.1.4 [2025-06-04]
_Bug fixes_
- Fix issue where steampipe was returning 0 exit-code in batch mode even incase of API failures. ([#4551](https://github.com/turbot/steampipe/issues/4551))

_Dependencies_
- Update FDW to 1.12.7 to remediate high vulnerabilities.

## v1.1.3 [2025-05-15]
_Bug fixes_
- Fix intermittent `Reattachment process not found` error when starting steampipe service. ([#4507](https://github.com/turbot/steampipe/issues/4507))

## v1.1.2 [2025-05-06]
_Bug fixes_
- Fix issue where system-ingestible output format(csv) was humanised(comma separated) leading to a breaking change in query outputs. ([#4525](https://github.com/turbot/steampipe/issues/4525))

## v1.1.1 [2025-04-25]
_Bug fixes_
- Fix issue where query batch mode outputs(json, csv, line) were not printing the rows received to stdout when any of the other rows returned an API error. ([#4516](https://github.com/turbot/steampipe/issues/4516))
- Fix issue where query batch mode table output always returned a 0 row count when timing was enabled. ([#4520](https://github.com/turbot/steampipe/issues/4520))

## v1.1.0 [2025-04-10]
_Whats new_
- Update database version to PostgreSQL 14.17. ([#4461](https://github.com/turbot/steampipe/issues/4461))

_Bug fixes_
- Fix issue where plugin start timeout was getting limited to 60s. ([#4477](https://github.com/turbot/steampipe/issues/4477))

## v1.0.3 [2025-02-03]
_Bug fixes_
- Update FDW to 1.12.2 to remediate critical and high vulnerabilities. ([#533](https://github.com/turbot/steampipe-postgres-fdw/issues/533))

## v1.0.2 [2025-01-20]
_Dependencies_
- Upgrade `crypto`, `net` and `go-git` packages to remediate critical and high vulnerabilities.

## v1.0.1 [2024-11-21]
_Bug fixes_
- Fix issue where the steampipe interactive meta-command `.cache clear` was not clearing the cache. ([#4443](https://github.com/turbot/steampipe/issues/4443))

## v1.0.0 [2024-10-22]
_Breaking changes_

The mod functionality, which was previously deprecated and moved to Powerpipe, has been removed in this version.  

- Removed the `check`, `dashboard`, `mod`, and `variable` commands. ([#4413](https://github.com/turbot/steampipe/issues/4413))
- Removed support for running named queries. ([#4416](https://github.com/turbot/steampipe/issues/4416))
- Removed the `watch` and `mod-location` CLI args from the `query` command. ([#4417](https://github.com/turbot/steampipe/issues/4417))
- Removed the `dashboard`, `dashboard-listen`, and `dashboard-port` CLI args from the `service` command. ([#4418](https://github.com/turbot/steampipe/issues/4418))
- Removed the `STEAMPIPE_MOD_LOCATION` and `STEAMPIPE_INTROSPECTION` env vars. ([#4419](https://github.com/turbot/steampipe/issues/4419))
- Removed support for deprecated `STEAMPIPE_CLOUD_HOST` and `STEAMPIPE_CLOUD_TOKEN` env vars. ([#4420](https://github.com/turbot/steampipe/issues/4420))
- Removed the `watch`, `introspection`, and `mod-location` workspace profile args. ([#4421](https://github.com/turbot/steampipe/issues/4421))
- Removed the `check` and `dashboard` options from workspace profiles. ([#4422](https://github.com/turbot/steampipe/issues/4422))
- Removed the `dashboard` option from global options (`default.spc`). ([#4423](https://github.com/turbot/steampipe/issues/4423))

## v0.24.2 [2024-09-13]
_Bug fixes_
- Fix incorrect versioning in v0.24.1. ([#4388](https://github.com/turbot/steampipe/issues/4388))

## v0.24.1 [2024-09-13]
_Bug fixes_
- Fix issue where steampipe failed to download embedded PostgreSQL database and FDW during installation. ([#4382](https://github.com/turbot/steampipe/issues/4382))

## v0.24.0 [2024-09-05]
_Whats new_
- Add ability to configure plugin startup timeout. ([#4320](https://github.com/turbot/steampipe/issues/4320))
- Install FDW and embedded postgres database from GHCR instead of GCP. ([#4344](https://github.com/turbot/steampipe/issues/4344))
- Update query JSON output format to add a `columns` property containing the column information. This allows us to handle duplicate column names by appending a unique suffix to duplicate column name ([#4317](https://github.com/turbot/steampipe/issues/4317))

Existing query JSON format:
```
$ steampipe query "select account_id, arn from aws_account" --output json
{
 "rows": [
  {
   "account_id": "123456789012",
   "arn": "arn:aws:::123456789012"
  }
 ]
}
```
New query JSON format(with new `columns` property):
```
$ steampipe query "select account_id, arn from aws_account" --output json
{
 "columns": [
  {
   "name": "account_id",
   "data_type": "text"
  },
  {
   "name": "arn",
   "data_type": "text"
  }
 ],
 "rows": [
  {
   "account_id": "123456789012",
   "arn": "arn:aws:::123456789012"
  }
 ]
}
```

_Bug fixes_
- Fix issue where plugin manager was incorrectly reporting a shutdown. ([#4365](https://github.com/turbot/steampipe/issues/4365))

## v0.23.5 [2024-08-21]
_Bug fixes_
- Fix issue where refresh connections was not creating a new connection if it was not in the search path. ([#4353](https://github.com/turbot/steampipe/issues/4353))

## v0.23.4 [2024-08-13]
_Whats new_
- Compiled with Go 1.22. ([#4340](https://github.com/turbot/steampipe/issues/4340))

_Bug fixes_
- Fix query error message to not include internal function names. ([#4335](https://github.com/turbot/steampipe/issues/4335))

## v0.23.3 [2024-07-17]
_Bug fixes_
- When installing plugins, do not use local docker config for credential store if the plugin is being installed from GHCR, enabling installation from GHCR to work even if docker-credential-desktop not in PATH. ([#4323](https://github.com/turbot/steampipe/issues/4323))
- Fix issue where steampipe returned 0 exit code even if failed to export snapshot. ([#4276](https://github.com/turbot/steampipe/issues/4276))
- Query command should support legacy 'true' and 'false' for --timing flag. ([#4282](https://github.com/turbot/steampipe/issues/4282))
- Fix issue where sps output is not working. ([#4297](https://github.com/turbot/steampipe/issues/4297))
- When loading creating connection plugins, return connections successfully created even if some connections fail, due to config not being available. ([#474](https://github.com/turbot/steampipe-postgres-fdw/issues/474))
- Show scan info in query JSON output only when timing config is verbose. ([#4292](https://github.com/turbot/steampipe/issues/4292))

## v0.23.2 [2024-05-17]
_Bug fixes_
- Update FDW to 1.11.2 to remove unnecessary NOTICE level log messages. ([#469](https://github.com/turbot/steampipe-postgres-fdw/issues/469))

## v0.23.1 [2024-05-11]
_Bug fixes_
- Update FDW to 1.11.1 to fix bad Linux Arm build. ([#4271](https://github.com/turbot/steampipe/issues/4271))
- Update hydrates count in timing verbose mode to use integer formatting(e.g. 119,138).  ([#4270](https://github.com/turbot/steampipe/issues/4270))

## v0.23.0 [2024-05-09]
_Whats new_
- Add support for connection key columns. ([#768](https://github.com/turbot/steampipe-plugin-sdk/issues/768))
  

A `ConnectionKeyColumn` defines a column that has a value which maps 1-1 to a Steampipe connection
    and so can be used to filter connections when executing an aggregator query. 

These columns are treated as (optional) KeyColumns. This means they are taken into account in the query planning.

- Add support for pushing down sort order. ([#447](https://github.com/turbot/steampipe-postgres-fdw/issues/447))
- Update limit pushdown logic to push down the limit if all sort clauses are pushed down. ([#458](https://github.com/turbot/steampipe-postgres-fdw/issues/458))
- Add support for `WHERE column=val1 OR column=val2 OR column=val3...`
- Adds support for verbose timing information. ([#4244](https://github.com/turbot/steampipe/issues/4244))
- Migrate from plugin registry from GCP to GHCR. ([#4232](https://github.com/turbot/steampipe/issues/4232))

_Bug fixes_
- Fix hang when timing disabled. ([#4237](https://github.com/turbot/steampipe/issues/4237))
- Add signal handler for signal 16 to avoid FDW crash.  ([#457](https://github.com/turbot/steampipe-postgres-fdw/issues/457))

_Breaking changes_
- JSON query output has changed from a JSON array of result rows to a JSON object with a `rows` property containing the result rows, and (optionally) a metadata property containing timing information.  

## v0.22.2 [2024-04-05]
_Bug fixes_
* Fix issue where daily update check message showed a <nil> when there was no message to show. ([#4206](https://github.com/turbot/steampipe/issues/4206))
* Fix issue where local plugins are not being loaded. ([#4196](https://github.com/turbot/steampipe/issues/4196))
* Re-add support for 'implicit' local plugins (i.e. the plugin binary exists but there is no entry in the `versions.json`). ([#4223](https://github.com/turbot/steampipe/issues/4223))
* Add support for nested dashboards. ([#4208](https://github.com/turbot/steampipe/issues/4208))

## v0.22.1 [2024-03-15]
_Whats new_
* Improve startup performance with high plugin count - parallelize plugin startup. ([#4183](https://github.com/turbot/steampipe/issues/4183))
* Add database SSL password support for encrypted private key in order to handle your own certificates. ([#4149](https://github.com/turbot/steampipe/issues/4149))

_Bug fixes_
* Fix issue where plugin list cannot re-create top-level versions.json file if the file has been corrupted or empty. ([#4191](https://github.com/turbot/steampipe/issues/4191))

## v0.22.0 [2024-03-06]

_Steampipe unbundled, introducing Powerpipe_

[Powerpipe](https://powerpipe.io) is now the recommended way to run dashboards and benchmarks!

Mods still work as normal in Steampipe for now, but they are deprecated and will be removed in a future release:
* [Steampipe unbundled →](https://steampipe.io/blog/steampipe-unbundled)
* [Powerpipe for Steampipe users →](https://powerpipe.io/blog/migrating-from-steampipe)

_Whats new_

* Added `version` column to `steampipe_plugin` table. ([#4141](https://github.com/turbot/steampipe/issues/4141))
* Direct all errors and warnings to standard error (stderr). ([4162](https://github.com/turbot/steampipe/issues/4162))

_Bug fixes_

* Fixed the issue where `search_path_prefix` set in `database options` does not alter the search path. ([#4160](https://github.com/turbot/steampipe/issues/4160))
* Fix issue where `asff` output was always missing the first row. ([#4157](https://github.com/turbot/steampipe/pull/4157))

_Deprecations and migrations_

* Steampipe mods and dashboards are now separately available in [Powerpipe](https://powerpipe.io), a new [open-source project](https://github.com/turbot/powerpipe). The steampipe mod, check and dashboard commands have been deprecated and will be removed in a future version. [Migration guide](https://powerpipe.io/blog/migrating-from-steampipe).
* Deprecated `cloud-host` and `cloud-token` CLI args, and replaced them with `pipes-host` and `pipes-token` respectively. ([#4137](https://github.com/turbot/steampipe/issues/4137))
* Deprecated `STEAMPIPE_CLOUD_HOST` and `STEAMPIPE_CLOUD_TOKEN` env vars, replaced with `PIPES_HOST` and `PIPES_TOKEN` respectively. ([#4137](https://github.com/turbot/steampipe/issues/4137))
* Deprecated `cloud_host` and `cloud_token` workspace args, replaced with `pipes_host` and `pipes_token` respectively. ([#4137](https://github.com/turbot/steampipe/issues/4137))
* Removed support for deprecated `terminal options`. ([#3751](https://github.com/turbot/steampipe/issues/3751))
* Removed support for deprecated `max_parallel` property in `general options`. ([#4132](https://github.com/turbot/steampipe/issues/4132))
* Removed support for deprecated `connection options`. ([#4131](https://github.com/turbot/steampipe/issues/4131))
* Removed deprecated `version` property from the mod `require` block. ([#3750](https://github.com/turbot/steampipe/issues/3750))

## v0.21.8 [2024-02-23]
_Bug fixes_
* Fix growing memory usage following file watching events when running dashboard server. ([#4150](https://github.com/turbot/steampipe/issues/4150))

## v0.21.7 [2024-02-09]
_Bug fixes_
* Fix variables not being reloaded after file watch event. ([#4123](https://github.com/turbot/steampipe/issues/4123))
* Fix modfile being left invalid after mod uninstall. Fix variables not being reloaded after file watch event. ([#4124](https://github.com/turbot/steampipe/issues/4124))

## v0.21.6 [2024-02-06]
_Bug fixes_
* Fix `HomeDirectoryModfileCheck` returning false positive, causing errors when executing steampipe out of the home directory. ([#4118](https://github.com/turbot/steampipe/issues/4118))

## v0.21.5 [2024-02-05]
_Bug fixes_
* Fix dependency variable validation - was failing if dependency variable value was set in the vars file. ([#4110](https://github.com/turbot/steampipe/issues/4110))
* Fix UI freeze when prompting for workspace variables. ([#4105](https://github.com/turbot/steampipe/issues/4105))

## v0.21.4 [2024-01-23]
_Bug fixes_
* Fixed schema clone function failing if table has an LTREE column. ([#4079](https://github.com/turbot/steampipe/issues/4079))
* Maintain the order of execution when running multiple queries in batch mode. ([#3728](https://github.com/turbot/steampipe/issues/3728))
* Fixes issue where using any meta-command would load connection state even if not required. ([#3614](https://github.com/turbot/steampipe/issues/3614))
* Fixes issue where plugin version backfilling would write versions.json to cwd if the plugin folder is not found. ([#4073](https://github.com/turbot/steampipe/issues/4073))
* Simplifies and fix available port check. ([#4030](https://github.com/turbot/steampipe/issues/4030))


## v0.21.3 [2023-12-22]
_Whats new_
* Allow using pprof on FDW when STEAMPIPE_FDW_PPROF environment variable is set. ([#368](https://github.com/turbot/steampipe-postgres-fdw/issues/368))

_Bug fixes_
* Set connection state to error if plugin load fails. ([#4043](https://github.com/turbot/steampipe/issues/4043))
* Fixes incorrect row count in timing output for aggregator connections. ([#402](https://github.com/turbot/steampipe-postgres-fdw/issues/402))
* OpenTelemetry metric names must only contain [A-Za-z0-9_.-]. ([#369](https://github.com/turbot/steampipe-postgres-fdw/issues/369))
* Maintain the order of execution when running multiple queries in batch mode. ([#3728](https://github.com/turbot/steampipe/issues/3728))

## v0.21.2 [2023-12-12]
_Whats new_
* Add `steampipe_plugin_column` introspection table to the `steampipe_internal` schema. ([#4003](https://github.com/turbot/steampipe/issues/4003))

_Bug fixes_
* Fixes issue where a query would return 'null' for an empty result set when output is set to json. ([#3955](https://github.com/turbot/steampipe/issues/3955))
* Fix custom registries bugs 
* Clean up apt temporary files in Dockerfile 

## v0.21.1 [2023-10-03]
_Bug fixes_
* Added support for the missing `mod-location` flag to the `steampipe variable list` command. ([#3942](https://github.com/turbot/steampipe/issues/3942))

## v0.21.0 [2023-10-02]
_Whats new?_
* Define [rate and concurrency limits](https://steampipe.io/docs/guides/limiter#concurrency--rate-limiting) for plugin execution. ([#3746](https://github.com/turbot/steampipe/issues/3746))
* Define multiple instances of a plugin version using a `plugin` connection config block. ([#3807](https://github.com/turbot/steampipe/issues/3807))
* The maximum memory used by plugins and the CLI can now be specified either in `plugin` instance definitions or the new `plugin` options block. ([#3807](https://github.com/turbot/steampipe/issues/3807))
* New introspection tables `steampipe_plugin` and `steampipe_plugin_limiter` containing all configured plugin instances and limiters. ([#3746](https://github.com/turbot/steampipe/issues/3746))
* New introspection table `steampipe_server_settings` populated with server settings data during service startup. ([#3462](https://github.com/turbot/steampipe/issues/3462))
* Running `plugin install` with no arguments installs all referenced plugins. ([#3451](https://github.com/turbot/steampipe/issues/3451))
* New `--output` flag for `plugin list` cmd allows selection between `json` and `table` output. ([#3368](https://github.com/turbot/steampipe/issues/3368))
* Each plugin directory ncontains a `version.json` which can be used to recompose the global plugin `versions.json` if it is missing or corrupt. ([#3492](https://github.com/turbot/steampipe/issues/3492))
* Typing `.cache` in interactive prompt shows the current value of cache. ([#2439](https://github.com/turbot/steampipe/issues/2439))
* Steampipe commands bypass plugin requirement check if installed plugin is locally built. ([#3643](https://github.com/turbot/steampipe/issues/3643))
* New `skip-config` flag disables writing of default plugin config during plugin installation. ([#3531](https://github.com/turbot/steampipe/issues/3531), [#2206](https://github.com/turbot/steampipe/issues/2206))
* Logs are now written to file instead of console. ([#2916](https://github.com/turbot/steampipe/issues/2916))
* When plugin startup fails, report useful message in the CLI. ([#3732](https://github.com/turbot/steampipe/issues/3732))
* Users are warned to not have mod.sp files in home directory. ([#2321](https://github.com/turbot/steampipe/issues/2321))
* Updated messaging when service is started on an unavailable port. ([#623](https://github.com/turbot/steampipe/issues/623))
* Log files are rotated if the process is active across date boundaries. ([#125](https://github.com/turbot/steampipe/issues/125), [#3825](https://github.com/turbot/steampipe/issues/3825))
* Listen hosts may be selected when starting steampipe service. ([#3505](https://github.com/turbot/steampipe/issues/3505))
* Initialisation behaviour for the sample options has been changed: always copy a sample file (`default.spc.sample`), but only overwrite the `default.spc` file with the sample content if the existing file has not been modified.  ([#3431](https://github.com/turbot/steampipe/issues/3431))
* Validation for the workspace profile `cache` settings. ([#3646](https://github.com/turbot/steampipe/issues/3646))
* Support OCI registries requiring authentication. ([#2819](https://github.com/turbot/steampipe/issues/2819))
* Compiled with Go 1.21. ([#3763](https://github.com/turbot/steampipe/issues/3763))

_Bug fixes_
* Plugin manager shutdown stalling intermittently due to deadlocks. ([#3818](https://github.com/turbot/steampipe/issues/3818))
* Temporary tables dropped in interactive prompt when pool connections recycled. ([#3781](https://github.com/turbot/steampipe/issues/3781),[#3543](https://github.com/turbot/steampipe/issues/3543))
* `service start` was not listening on `network` by default. ([#3593](https://github.com/turbot/steampipe/issues/3593))
* Multi line logs from plugins not rendered correctly in plugin logs. ([#3678](https://github.com/turbot/steampipe/issues/3678))
* `.inspect` panicking for long column descriptions. ([#3709](https://github.com/turbot/steampipe/issues/3709))
* Interactive prompt crashing when there is a code panic. ([#3713](https://github.com/turbot/steampipe/issues/3713))
* Incorrect zsh completion instructions.
* Steampipe should not create export files for cancelled control runs. ([#3578](https://github.com/turbot/steampipe/issues/3578))
* `BuildFullResourceName` not validating non empty arguments. ([#3601](https://github.com/turbot/steampipe/issues/3601))
* Spinner not showing when exporting check results. ([#3577](https://github.com/turbot/steampipe/issues/3577))
* `stdin` was consumed by `query` command even if there are arguments. ([#1985](https://github.com/turbot/steampipe/issues/1985))
* When exporting multiple benchmarks, results now merged the results into a single export. ([#2380](https://github.com/turbot/steampipe/issues/2380))
* Raise warning when pseudo-resources are ignored because of named HCL resources. ([#1328](https://github.com/turbot/steampipe/issues/1328))
* Database reinstalled unnecessarily if any FDW files were missing. ([#2040](https://github.com/turbot/steampipe/issues/2040))
* Improved error message when steampipe fails to parse a mod definition file because mod block does not exist. ([#1198](https://github.com/turbot/steampipe/issues/1198))
* Only `install-dir` and `workspace` flags should be global flags. All other flags should only apply to specific command. ([#3542](https://github.com/turbot/steampipe/issues/3542))
* Passing an empty list for list variables was not working. ([#2094](https://github.com/turbot/steampipe/issues/2094))
* Show deprecation warning for `version` field in `require` block of mod definition.
* Temporary directories were not always being cleaned  up after plugin commands.
* `plugin list` returned nothing if no plugins were installed. ([#3927](https://github.com/turbot/steampipe/issues/3927))

_Deprecations and migrations_
* Table `steampipe_connection_state` renamed to `steampipe_connection`
* Removed migration and backward compatibility of data files from v0.13.0. ([#3517](https://github.com/turbot/steampipe/issues/3517))
* Removed deprecated `workspace-chdir` flag. ([#3925](https://github.com/turbot/steampipe/issues/3925))
* Migrated from `cloud.steampipe.io` to `pipes.turbot.com`. ([#3724](https://github.com/turbot/steampipe/issues/3724))
* Removed support for plugins which do not support multiple connections (i.e. using SDK < v4.0.0).
* Deprecated `terminal options`.

## v0.20.12 [2023-09-14]
_Whats new?_
* Updated help outputs for steampipe mod commands. ([#1817](https://github.com/turbot/steampipe/issues/1817))

_Bug fixes_
* Fixes issue where expired root and server SSL certificates were not getting rotated. ([#3596](https://github.com/turbot/steampipe/issues/3596))
* Fixes issue where steampipe was returning an `index out of range` error when the `children` property of a `benchmark` contains an invalid name. ([#3563](https://github.com/turbot/steampipe/issues/3563))
* Steampipe should not validate locally installed plugins when connecting to remote database. ([#3516](https://github.com/turbot/steampipe/issues/3516))

## v0.20.11 [2023-08-28]
_Bug fixes_
* Fix validation error for `input` blocks using `base` inheritance. ([#3755](https://github.com/turbot/steampipe/issues/3755))
* Fix support for mixed case schema names. ([#3753](https://github.com/turbot/steampipe/issues/3753))
* If the SQL file passed as an argument to `steampipe query` does not exist, display the `file does not exist` error. ([#1752](https://github.com/turbot/steampipe/issues/1752))

## v0.20.10 [2023-08-11]
_Bug fixes_
* Fixes issue where CAPITAL arguments to '.cache' meta command were not getting recognised.  ([#3670](https://github.com/turbot/steampipe/issues/3670))
* Fixes issue where `port` property in dashboard options was not respected.  ([#3664](https://github.com/turbot/steampipe/issues/3685))
* Fixes issue where using a bad workspace-database with a valid token gives invalid token as the error.  ([#3610](https://github.com/turbot/steampipe/issues/3610)) 
* Fixes timing issue where refresh connections was sometimes not run when starting service.  ([#3734](https://github.com/turbot/steampipe/issues/3734))
* Fixes issue where db connections are not closed after sending postgres notification.  ([#3744](https://github.com/turbot/steampipe/issues/3744))

## v0.20.9 [2023-07-11]
_Bug fixes_
* Fix aggregator connections being dropped intermittently when refreshing connections. ([#3664](https://github.com/turbot/steampipe/issues/3664))
* Ensure dynamic aggregator schema is updated if connections are added. ([#3645](https://github.com/turbot/steampipe/issues/3645))

## v0.20.8 [2023-07-03]
_Bug fixes_
* Fixes issue where setting cache ttl from the CLI results in cache being disabled for that session. ([#3639](https://github.com/turbot/steampipe/issues/3639))

## v0.20.7 [2023-06-22]
_Bug fixes_
* Fixes issue where aggregator connections are updated every time RefreshConnections runs. ([#3582](https://github.com/turbot/steampipe/issues/3582))
* Add `connections` column to steampipe_connection_state table. ([#3582](https://github.com/turbot/steampipe/issues/3582))
* Fixes issue where exporting check all yields a badly formatted filename. ([#3591](https://github.com/turbot/steampipe/issues/3591))
* Fix variable value validation not taking into account command line variable values. ([#3606](https://github.com/turbot/steampipe/issues/3606))

## v0.20.6 [2023-06-14]
_Bug fixes_
* Fix variable validation ([#3546](https://github.com/turbot/steampipe/issues/3546)):
  * Raise warning or error when setting a value for a variable which is not found or inaccessible (e.g. because it is in a transitive dependency). 
  * Validate that mod require `args` properties can be resolved. 
* Support resolution of variables for transitive dependencies using parent mod `require` block `args` property. ([#3549](https://github.com/turbot/steampipe/issues/3549))
* `steampipe mod update` now updates transitive mods. ([#3547](https://github.com/turbot/steampipe/issues/3547))
* It is now be possible to set values for variables in the current mod using fully qualified variable names. ([#3551](https://github.com/turbot/steampipe/issues/3551))
* Only variables for root mod and top level dependency mods can be set by user.  ([#3550](https://github.com/turbot/steampipe/issues/3550))
* Avoid orphan plugin processes when running short batch queries. ([#3514](https://github.com/turbot/steampipe/issues/3514))
* Delete dynamic schemas before updating them to avoid a timing issue showing incorrect schema. ([#3510](https://github.com/turbot/steampipe/issues/3510))
* Fixes issue where blank dimension values are leaving extra spaces in 'table' rendering. ([#3474](https://github.com/turbot/steampipe/issues/3474))
* Fixes issue when steampipe fails to startup if plugin version file is blank. ([#3518](https://github.com/turbot/steampipe/issues/3518))
* Fixes issue where OS specific metadata directories were being considered as check templates. ([#3523](https://github.com/turbot/steampipe/issues/3523))
* Fixes issue where prefixing a 'v' on a version stream during plugin install would come back with 'not found'. ([#3513](https://github.com/turbot/steampipe/issues/3513))
* Increase plugin load timeout to 20s. ([#3564](https://github.com/turbot/steampipe/issues/3564))
  Fixes issue where timing is not shown in interactive prompt even if .timing is on. ([#3557](https://github.com/turbot/steampipe/issues/3557))
* Fixes issue where 'dot' commands in interactive prompt fail to execute if there's a file/folder by the same name in the working directory. ([#3558](https://github.com/turbot/steampipe/issues/3558))
* Fixes issue where 'plugin list' hangs if there are connections with 'import_schema = "disabled"'. ([#3561](https://github.com/turbot/steampipe/issues/3561))

## v0.20.5 [2023-05-31]
_Bug fixes_
* Set incomplete connections to `Incomplete` before setting ready connections to `Pending` to avoid ready connections ending up `Incomplete`. ([#3507](https://github.com/turbot/steampipe/issues/3507))

## v0.20.4 [2023-05-31]
_Bug fixes_
* Ensure `Ready` connections are set to `Pending` state on startup. This makes sure connection changes are reflected in the connection schema if a query is executed soon after startup. ([#3483](https://github.com/turbot/steampipe/issues/3483))

## v0.20.3 [2023-05-30]
_Whats new?_
* Update refresh connections to execute updates serially by default.  ([#3498](https://github.com/turbot/steampipe/issues/3498))

_Bug fixes_
* Fix issue where result counter spinner was not showing up in interactive when timing was enabled. ([#3481](https://github.com/turbot/steampipe/issues/3481))
* Fixes issue where dependency mods are installed even if there is an installed mod which satisfies requirement. ([#3475](https://github.com/turbot/steampipe/issues/3475))
* Ensure a schema is created for blank aggregators when connections are added. ([#3488](https://github.com/turbot/steampipe/issues/3488))
* Fix issue where `steampipe completion` command was creating install directories. ([#3485](https://github.com/turbot/steampipe/issues/3485))
* Don't use custom theme color `yellow` for severity cards, to avoid clashing with Tailwind's yellow palette. ([#3501](https://github.com/turbot/steampipe/issues/3501))

## v0.20.2 [2023-05-19]
_Whats new?_
* Re-add support for legacy command-schema. ([#3457](https://github.com/turbot/steampipe/issues/3457))

_Bug fixes_
* Cleanup temp plugin files when killing plugin manager. ([#3292](https://github.com/turbot/steampipe/issues/3292))

## v0.20.1 [2023-05-19]
_Bug fixes_
- Update FDW version to v1.7.1 to work around bad Linux Arm build of FDW v1.70. ([#3455](https://github.com/turbot/steampipe/issues/3455), [#311](https://github.com/turbot/steampipe-postgres-fdw/issues/311))

## v0.20.0 [2023-05-18]

#### Connection Management 
- Optimise connection initialisation for high connection count ([#3394](https://github.com/turbot/steampipe/issues/3394),[#3267](https://github.com/turbot/steampipe/issues/3267),[#3236](https://github.com/turbot/steampipe/issues/3236),[#3229](https://github.com/turbot/steampipe/issues/3229),[#3413](https://github.com/turbot/steampipe/issues/3413))
  - Execute RefreshConnections asyncronously in service startup
  - Start executing queries without waiting for connections to load, add smart error handling to wait for required connection
  - Optimise autocomplete for high connection count
  - Autocomplete and inspect data available before all conections are refreshed
  - Add `steampipe_connection_state` table to indicate the loading state of connections
  - Add support for `import_schema` property in connection config, controlling whether to create a postgres schema for a steampipe connection. Closes #3407
  - Optimise schema creation by cloning connection schemas
  - Add locking to ensure only a single instance of RefreshConnections runs
  - Update refresh connections to write comments for exemplar schemas first, followed by remaining schemas.  

- Update connection and plugin validation during refreshConnections. ([#3432](https://github.com/turbot/steampipe/issues/3432),[#3402](https://github.com/turbot/steampipe/issues/3402))
  - ensure failed connections are set to 'error' in connection state.  
  - Schema names starting with steampipe_ are to be reserved for steampipe. 

#### Mod Dependency Management
- Support mods requiring different versions of the same depdency mod. ([#3302](https://github.com/turbot/steampipe/issues/3302))
- Support transitive dependencies referencing variables from different versions of same mod.([#3337](https://github.com/turbot/steampipe/issues/3337))
- Resource references in dependency mods must be fully qualified. ([#3335](https://github.com/turbot/steampipe/issues/3335))
- Locals in dependency mods cannot be referenced. ([#3336](https://github.com/turbot/steampipe/issues/3336))
- Fix issue where 'mod install' on an existing mod would sometimes corrupt the 'mod.sp' file. ([#3376](https://github.com/turbot/steampipe/issues/3376))
- Fix issue where mod installation would fail silently for unmet dependencies in top mod in force mode. ([#3358](https://github.com/turbot/steampipe/issues/3358))
- Fix issue where mod list output is not printed in a specific order. ([#3349](https://github.com/turbot/steampipe/issues/3349))
- Fix issue where a mod would install even if plugin dependencies are not met. ([#3041](https://github.com/turbot/steampipe/issues/3041))
- Fix issue where running mods with unmet dependencies does not raise warnings. ([#3324](https://github.com/turbot/steampipe/issues/3324))
- Fix mod commands failing when using a `https` prefix. ([#3257](https://github.com/turbot/steampipe/issues/3257))
- Fix issue where mod install/update continues installation even with unsatisfied requirements. ([#3291](https://github.com/turbot/steampipe/issues/3291))
- Fix nil reference exception when loading a mod using the legacy `requires` property. ([#3347](https://github.com/turbot/steampipe/issues/3347))

#### Caching

- Updates in cache configuration to allow disabling of all caching on server. ([#3258](https://github.com/turbot/steampipe/issues/3258))
  - STEAMPIPE_CACHE environment variable controls both *service* cache-enabled and  *client* cache-enabled
  - *service* cache enabled is used by the plugin manager to enable/disable caching on the plugins during startup.
  - *client* cache enabled is used to enable/disable the cache on the database session. 
- Introduce SQL functions to easily manipulate caching functionality - `meta_cache()` and `meta_cache_ttl()`. ([#3442](https://github.com/turbot/steampipe/issues/3442))

_What's new?_
- Add support for time-series charts. ([#1389](https://github.com/turbot/steampipe/issues/1389))
- Updates to workspace profile - add additional properties and command specific options blocks. ([#3223](https://github.com/turbot/steampipe/issues/3223))
- Adds a `--progress` flag to `plugin install` to disable progress bars. ([#2953](https://github.com/turbot/steampipe/issues/2953))
- Detect older versions of MacOS and warn that Steampipe does not support them. ([#3256](https://github.com/turbot/steampipe/issues/3256))
- Updates the default content written to 'default.spc' and remove deprecated blocks. ([#3391](https://github.com/turbot/steampipe/issues/3391))
- Show plugin name with stream (if not latest) in the progress bar during plugin update. ([#3241](https://github.com/turbot/steampipe/issues/3241),[#3330](https://github.com/turbot/steampipe/issues/3330))
- Replace all '...' with ellipsis … in terminal output. ([#3441](https://github.com/turbot/steampipe/issues/3441))
- Add check to the mod init function so users are aware if it's run in the home directory or if there are a large number of non-mod files in the path. ([#2562](https://github.com/turbot/steampipe/issues/2562))
- Add query column in introspection tables to populate FullName if a QueryProvider references a named query. ([#3161](https://github.com/turbot/steampipe/issues/3161))
- Improve error message when running steampipe check/dashboard outside a mod. ([#3215](https://github.com/turbot/steampipe/issues/3215))

_Bug fixes_
- Fixes issue where not being able to open the browser results in a fatal error during login. ([#3437](https://github.com/turbot/steampipe/issues/3437))
- Fixes issue where 'internal' would be added twice in the search_path if one is mentioned in the non default search path. ([#3397](https://github.com/turbot/steampipe/issues/3397))
- Set mod name in resource metadata for pseudo-resources. ([#3405](https://github.com/turbot/steampipe/issues/3405))
- Fix error message when connecting to steampipe cloud if login token has expired or become corrupted. ([#3418](https://github.com/turbot/steampipe/issues/3418))
- Fix `invalid output format` error when running dashboard if `output` is set in terminal options. ([#3293](https://github.com/turbot/steampipe/issues/3293))
- Fixes issue where execution continues even if there's an unexpected error in parsing config. ([#3286](https://github.com/turbot/steampipe/issues/3286))
- Fix rendering issues when running .inspect. ([#3268](https://github.com/turbot/steampipe/issues/3268))
- Fixes issue where spinner was not showing up in interactive prompt while a query was executing. ([#3259](https://github.com/turbot/steampipe/issues/3259))
- Fix crash on shutdown if init not complete. ([#3352](https://github.com/turbot/steampipe/issues/3352))
- Fixes issue where workspace introspection option was boolean instead of control/info/none. ([#3389](https://github.com/turbot/steampipe/issues/3389))
- Fixes issue where network failures during plugin install was returning 0 exit code. ([#3367](https://github.com/turbot/steampipe/issues/3367))
- Ensure successful shutdown after dashboard service start failure. ([#3354](https://github.com/turbot/steampipe/issues/3354))
- Ensure plugin-manager command does not execute scheduled tasks - avoid deprecation warnings which make the plugin manager GRPC startup fail. ([#3410](https://github.com/turbot/steampipe/issues/3410)


## v0.19.5 [2023-04-27]
_Bug fixes_
* Fix plugin manager to crash with unhandled signal caused by connection validation warning following a file watcher event. ([#3371](https://github.com/turbot/steampipe/issues/3371))
* Fix array bounds error when querying an aggregator with no children. Show useful error instead. ([#303](https://github.com/turbot/steampipe-postgres-fdw/issues/303))
* Fixes issue where having non graphic code points in output would mess up table output in interactive. ([#3205](https://github.com/turbot/steampipe/issues/3205))

## v0.19.4 [2023-04-06]
_What's new?_
* Dashboard snapshot href links now work for external URLs. ([#3278](https://github.com/turbot/steampipe/issues/3278))
* Numeric dashboard benchmark summary card values should render using locale string. ([#3299](https://github.com/turbot/steampipe/issues/3299))
* Improve hover title grammar of critical/high severity dashboard benchmark badges. ([#3300](https://github.com/turbot/steampipe/issues/3300))

* _Bug fixes_
* Fix issue where installing transitive mod dependencies leaves the lock file with an entry with an incorrect key. ([#3285](https://github.com/turbot/steampipe/issues/3285))
* Fix duplicate dashboard UI benchmark nodes being rendered for deep benchmark hierarchies with mixture of benchmark and child controls. ([#3298](https://github.com/turbot/steampipe/issues/3298))

## v0.19.3 [2023-03-24]
_Bug fixes_
* Fix issue where the json output of variable list command was returning wrong values for `value` and `value_default` fields. ([#3265](https://github.com/turbot/steampipe/issues/3265))
* Fix dashboard UI crash when select inputs return null labels or values. ([#3244](https://github.com/turbot/steampipe/issues/3244))

## v0.19.2 [2023-03-16]
_Bug fixes_
* When creating a query snapshot, respect the `snapshot-title` arg when assigning a title to the dashboard. ([#3233](https://github.com/turbot/steampipe/issues/3233))

## v0.19.1 [2023-03-09]
_Bug fixes_
* Fix `service stop` failing if invoked directly after a schema change notification. ([#3206](https://github.com/turbot/steampipe/issues/3206))
 
## v0.19.0 [2023-03-09]
_What's new?_
* Add support for aggregator connections with dynamic tables. ([#2886](https://github.com/turbot/steampipe/issues/2886))
* Support updating of dynamic plugin schemas based on file watching events (e.g. a new csv file is created in a watched location) ([#2767](https://github.com/turbot/steampipe/issues/2767))
* Make workspace loading asynchronous. ([#3123](https://github.com/turbot/steampipe/issues/3123))
* Make database start timeout configurable. ([#3038](https://github.com/turbot/steampipe/issues/3038))
* When initialising interactive mode, instead of showing `Initializing...`, show the current status. ([#3077](https://github.com/turbot/steampipe/issues/3077))
* Show the exported file location when `--progress` flag is enabled. ([#2860](https://github.com/turbot/steampipe/issues/2860))
* For aggregator connections, add child connection names to connections.json. ([#3079](https://github.com/turbot/steampipe/issues/3079))
* Aggregator connection with no child connections should only be a warning - not an error. ([#3155](https://github.com/turbot/steampipe/issues/3155))
* Cleanup connection state file to remove legacy properties. ([#3086](https://github.com/turbot/steampipe/issues/3086))
* Dashboard server should emit updated dashboard metadata when available dashboards changes. ([#3182](https://github.com/turbot/steampipe/issues/3182))
* Update interactive prompt `.inspect` output and autocomplete based on changes to connection config or dynamic schema updates. ([#3184](https://github.com/turbot/steampipe/issues/3184))

_Bug fixes_
* Steampipe config validation failure no longer prevents Steampipe commands from running - instead invalid connections are removed. ([#3156](https://github.com/turbot/steampipe/issues/3156))
* Fixes issue where variables list command was not including description in JSON output. ([#3114](https://github.com/turbot/steampipe/issues/3114))
* Ensure version display is consistent between startup and `--v` flag. ([#3031](https://github.com/turbot/steampipe/issues/3031))
* When a plugin fails to load, remove connections for that plugin from the connection state file. ([#3124](https://github.com/turbot/steampipe/issues/3124))
* Fix running a single dashboard from the command line failing if the dashboard needs inputs and the dashboard name is not fully qualified. ([#3168](https://github.com/turbot/steampipe/issues/3168),[#3154](https://github.com/turbot/steampipe/issues/3154))
* Fix workspace load crash for invalid mod definition. ([#3174](https://github.com/turbot/steampipe/issues/3174))
* Limit should not be pushed down if there are unconverted restrictions. ([#291](https://github.com/turbot/steampipe-postgres-fdw/issues/291))
* Dashboard text inputs are not correctly themed in Steampipe Cloud dashboard UI dark mode. ([#3181](https://github.com/turbot/steampipe/issues/3181))
* Fix nil reference panic in FDW when a scan fails to start - do not add an iterator to Hub.runningIterators until scan is started successfully. ([#298](https://github.com/turbot/steampipe-postgres-fdw/issues/298))
* Fix `tuple concurrently updated ` error when running multiple instances of steampipe dashboard concurrently. ([#3188](https://github.com/turbot/steampipe/issues/3188))
* Fix Postgres error "cached plan must not change result type" when dynamic plugin schema changes. ([#3185](https://github.com/turbot/steampipe/issues/3185))

## v0.18.6 [2023-02-15]
_Bug fixes_
* Fix issue where inspect would not work with table names with a '.' (dot). ([#2455](https://github.com/turbot/steampipe/issues/2455))
* Fix issue where autocomplete does not quote table names that need to be quoted. ([#3065](https://github.com/turbot/steampipe/issues/3065))
* Fix issue where check csv output was appending an extra line at the end. ([#3106](https://github.com/turbot/steampipe/issues/3106))
* Fixes issue where snapshot mode in query leads to duplicate rows in console/file output. ([#3112](https://github.com/turbot/steampipe/issues/3112))

## v0.18.5 [2023-02-07]
_Bug fixes_
* Fix double counting of control errors in benchmark summary. ([#3084](https://github.com/turbot/steampipe/issues/3084))

## v0.18.4 [2023-02-03]
_Bug fixes_
* Fix dashboard panel detail crash when viewing data tables with non-string values in text columns. ([#3071](https://github.com/turbot/steampipe/issues/3071))
* Fixes issue where steampipe notifies of available update even if plugin is updated. ([#2998](https://github.com/turbot/steampipe/issues/2998))
* Fix issue where snapshot creation was failing for command line queries in batch mode. ([#2943](https://github.com/turbot/steampipe/issues/2943))
* Add a helpful error message when snapshot sharing fails because of an invalid token. ([#2944](https://github.com/turbot/steampipe/issues/2944))
* Fix query batch mode returning zero exit code when rows return errors. ([#3044](https://github.com/turbot/steampipe/issues/3044))
* Fixes issue where options from `default.spc` were taking precedence over environment variable settings. ([#3060](https://github.com/turbot/steampipe/issues/3060))

## v0.18.3 [2023-02-01]
_Bug fixes_
* Fix issue where `search_path` is not getting set from connection-config watching in service mode. ([#3047](https://github.com/turbot/steampipe/issues/3047))
* Fix issue where extra newline was added to interactive prompt before messages were printed. ([#3027](https://github.com/turbot/steampipe/issues/3027))
* Fix issue where when running a dashboard from a dependent mod, default variable vals are not being included in the snapshot. ([#2730](https://github.com/turbot/steampipe/issues/2730))
* Update `--version` output to match the startup message. ([#3028](https://github.com/turbot/steampipe/issues/3028))

## v0.18.2 [2023-01-27]
_Bug fixes_
* Fix dashboard property blocks not taking effect in node/edge property tooltips. ([#3026](https://github.com/turbot/steampipe/issues/3026))

## v0.18.1 [2023-01-18]
_Bug fixes_
* Fix workspace file watching events sometime causing dashboard to stall and stop responding to events. ([#3007](https://github.com/turbot/steampipe/issues/3007))
* Fix cancelling dashboards (e.g. by pressing 'back' on the browser) sometimes leaving the dashboard server in a state where it will not respond to socket events. ([#3008](https://github.com/turbot/steampipe/issues/3008))
* Increase database connection timeout and improve the error message if connection failure occurs. ([#2377](https://github.com/turbot/steampipe/issues/2377))
* Validate that input references are of the form `self.input.<input-name>`. ([#2990](https://github.com/turbot/steampipe/issues/2990))
* Fix `check --where` and `check --tag`. ([#3001](https://github.com/turbot/steampipe/issues/3001))
* Ensure correct exit code is returned when a mod plugin requirements are not met. ([#2986](https://github.com/turbot/steampipe/issues/2986))
* Fix dashboard leaf_node_updated events for v0.17.4 CLI being ignored by v0.18.0 UI clients. ([#2994](https://github.com/turbot/steampipe/issues/2994))
* Fix dashboard table interpolated template rendering not working in line view. ([#3014](https://github.com/turbot/steampipe/issues/3014))
* Fix HCL validation to allow benchmark and control blocks in dashboard. ([#3015](https://github.com/turbot/steampipe/issues/3015))

## v0.18.0 [2023-01-12]
_What's new?_
* Add support for visualisations of your data with graphs, with easily composable data structures using nodes and edges. ([#2249](https://github.com/turbot/steampipe/issues/2249))
* Improved dashboard UI panel controls for quicker access to common tasks such as downloading panel data. ([#2663](https://github.com/turbot/steampipe/issues/2663))
* Add support for `with` blocks. ([#2772](https://github.com/turbot/steampipe/issues/2772))
* Add support for `param` runtime dependencies. ([#2910](https://github.com/turbot/steampipe/issues/2910))
* Add dashboard panel log to panel detail to get an understanding of the execution history of a panel. ([#2895](https://github.com/turbot/steampipe/issues/2895))
* Remove usage of prepared statements - instead execute sql directly.([#2789](https://github.com/turbot/steampipe/issues/2789))
* Modify the update checker to run asynchronously. ([#2770](https://github.com/turbot/steampipe/issues/2770))
* Update steampipe_reference introspection table to include references from `with` blocks. ([#2934](https://github.com/turbot/steampipe/issues/2934))
* Update arg validation to ignore extra named args but fail on extra positional args (currently fails if too many named args passed) ([#2783](https://github.com/turbot/steampipe/issues/2783))
* Update dashboard states to `initialized`, `blocked`, `running`, `complete`, `error`, `canceled`. ([#2939](https://github.com/turbot/steampipe/issues/2939))
* Update dashboard UI version mismatch logic to redirect to a version-enabled URL to get past localhost cached index.html. ([#2940](https://github.com/turbot/steampipe/issues/2940))
* Upgrades 'pgx' to v5. ([#2776](https://github.com/turbot/steampipe/issues/2776))
* Add a `--max-parallel` flag to `dashboard` command and set default to 10. ([#2754](https://github.com/turbot/steampipe/issues/2754))
* When parsing query args, ensure jsonb args are passed to query as string not map.([#2802](https://github.com/turbot/steampipe/issues/2802)) 
* Update Makefile to allow overriding build output directory path 

_Bug fixes_
* Fixes issue where interactive prompt was not showing timing data for 'json', 'csv' and 'line' outputs. ([#2699](https://github.com/turbot/steampipe/issues/2699))
* Fixes issue where value from '--separator' was not being used in CSV rendering. ([#544](https://github.com/turbot/steampipe/issues/544))
* Fixes issue where implicit services are not shutting down when the last instance of steampipe exits. ([#2833](https://github.com/turbot/steampipe/issues/2833))
* When editing dashboard files, after adding/fixing errors in the HCL the dashboard server will sometimes stall. ([#2952](https://github.com/turbot/steampipe/issues/2952))
* Dashboard select/combo inputs using integer `value` do not render options. ([#2972](https://github.com/turbot/steampipe/issues/2972))

_Deprecations_
* Hcl validation is now stricter. ([#2923](https://github.com/turbot/steampipe/issues/2923))
* Add deprecation warnings for deprecated hcl properties. ([#2973](https://github.com/turbot/steampipe/issues/2973))
* Remove `search_path` and `search_path_prefix` from `control` and `query` resources. ([#2963](https://github.com/turbot/steampipe/issues/2963))
* Exit codes have been updated. ([#2329](https://github.com/turbot/steampipe/issues/2395))
```
const (
	ExitCodeSuccessful                 = 0
	ExitCodeControlsAlarm              = 1   // check - no runtime errors, 1 or more control alarms, no control errors
	ExitCodeControlsError              = 2   // check - no runtime errors, 1 or more control errors
	ExitCodePluginLoadingError         = 11  // plugin - loading error
	ExitCodePluginListFailure          = 12  // plugin - listing failed
	ExitCodePluginNotFound             = 13  // plugin - not found
	ExitCodeSnapshotCreationFailed     = 21  // snapshot - creation failed
	ExitCodeSnapshotUploadFailed       = 22  // snapshot - upload failed
	ExitCodeServiceSetupFailure        = 31  // service - setup failed
	ExitCodeServiceStartupFailure      = 32  // service - start failed
	ExitCodeServiceStopFailure         = 33  // service - stop failed
	ExitCodeQueryExecutionFailed       = 41  // query - 1 or more queries failed - change in behavior(previously the exitCode used to be the number of queries that failed)
	ExitCodeLoginCloudConnectionFailed = 51  // login - connecting to cloud failed
	ExitCodeInitializationFailed       = 250 // common - initialization failed
	ExitCodeBindPortUnavailable        = 251 // common (service/dashboard) - port binding failed
	ExitCodeNoModFile                  = 252 // common - no mod file
	ExitCodeFileSystemAccessFailure    = 253 // common - file system access failed
	ExitCodeInsufficientOrWrongInputs  = 254 // common - runtime error (insufficient or wrong input)
	ExitCodeUnknownErrorPanic          = 255 // common - runtime error (unknown panic)
)
```
## v0.17.4 [2022-12-02]
_Bug fixes_
* Fixes issue where the `--separator` flag was not being used in the `csv` output/export for `steampipe check`. ([#544](https://github.com/turbot/steampipe/issues/544))

## v0.17.3 [2022-11-24]
_Bug fixes_
* Fix shared memory errors for high connection count - update postgres config to reverts `max_locks_per_transaction` to the pre v0.17.0 value of 2048. ([#2756](https://github.com/turbot/steampipe/issues/2756))

## v0.17.2 [2022-11-18]
_Bug fixes_
* Fix dashboard interpolated string expressions with adjacent expressions not separated by spaces not rendering the second expression ([#2752](https://github.com/turbot/steampipe/issues/2752))
* Ensure workspace and panel errors are shown in dashboard panels ([#2742](https://github.com/turbot/steampipe/issues/2742))
* Fix issue where control execution errors were not shown in CSV rendering. ([#2674](https://github.com/turbot/steampipe/issues/2674))
* Escape query arguments when resolving prepared statement execution SQL. ([#2676](https://github.com/turbot/steampipe/issues/2676))
* Fixes issue where a '--where' or '--tag' flag were not creating the introspection tables. ([#2670](https://github.com/turbot/steampipe/issues/2670))

## v0.17.1 [2022-11-10]
_Bug fixes_
* Fix query command `--export` flag raising an error that it cannot be used in interactive mode, even when not in interactive mode. ([#2707](https://github.com/turbot/steampipe/issues/2707))
* Fix RefreshConnections sometimes storing an unset plugin ModTime property in the connection state file. This leads to failure to refresh connections when plugin has been rebuilt or updated. ([#2721](https://github.com/turbot/steampipe/issues/2721))
* Fix dashboard text inputs being editable in snapshot mode. ([#2717](https://github.com/turbot/steampipe/issues/2717))
* Fix dashboard JSONB columns in CSV data downloads not serialising correctly. ([#2733](https://github.com/turbot/steampipe/issues/2733))
* Add dashboard error modal when users are running a different UI and CLI version ([#2728](https://github.com/turbot/steampipe/issues/2728))
* Fixes control dashboards not displaying progress. ([#2735](https://github.com/turbot/steampipe/issues/2735))

## v0.17.0 [2022-11-08]
_What's new?_
* Add support for `workspace profiles`, defined using HCL config and selected using `--workspace` arg. ([#2510](https://github.com/turbot/steampipe/issues/2510), [#2574](https://github.com/turbot/steampipe/issues/2574))
* Update CLI to upload snapshots to Steampipe cloud using `--share` and `--snapshot` options. ([#2367](https://github.com/turbot/steampipe/issues/2367))
* Add `steampipe login` command. ([#2583](https://github.com/turbot/steampipe/issues/2583))
* Update `dashboard` command to support passing a dashboard name as an argument. ([#2365](https://github.com/turbot/steampipe/issues/2365))
* Adds `list` sub command for `query`, `check` and `dashboard`. ([#2653](https://github.com/turbot/steampipe/issues/2653))
* Add `snapshot`/`sps` output and export format. ([#2473](https://github.com/turbot/steampipe/issues/2473))
* Add `--snapshot-title arg`. Ensure snapshots and exports are named consistently.([#2666](https://github.com/turbot/steampipe/issues/2666))
* Add `autocomplete` meta command and terminal option. ([#2560](https://github.com/turbot/steampipe/issues/2560), [#1692](https://github.com/turbot/steampipe/issues/1692))
* Add ability to save and open snapshots from the dashboard UI. ([#2577](https://github.com/turbot/steampipe/issues/2577))
* Add support for viewing control snapshots in the dashboard UI. ([#2688](https://github.com/turbot/steampipe/issues/2688))
* Add a configurable query timeout. ([#666](https://github.com/turbot/steampipe/issues/666), [#2593](https://github.com/turbot/steampipe/issues/2593)) 
* Update database code to use `pgx` interface so we can leverage the connection pool hook functions to pre-warm connections. ([#2422](https://github.com/turbot/steampipe/issues/2422))
* Rationalise and simplify postgres configuration. ([#2471](https://github.com/turbot/steampipe/issues/2471))
* Support executing any query-provider resources using the steampipe query command. ([#2558](https://github.com/turbot/steampipe/issues/2558))
* Improve help messages when a plugin is installed but the connection is not configured. ([#2319](https://github.com/turbot/steampipe/issues/2319))
* Add better help messages for mod plugin requirements not satisfied error. ([#2361](https://github.com/turbot/steampipe/issues/2361))
* Reduce the max frequency of connection config changed events to every 4 second. ([#2535](https://github.com/turbot/steampipe/issues/2535))
* Add `Variables` and `Inputs` to dashboard `ExecutionStarted` event. ([#2606](https://github.com/turbot/steampipe/issues/2606))
* Validate check output and export formats _before_ execution. ([#2619](https://github.com/turbot/steampipe/issues/2619)) 
* When starting a plugin process, pass a SecureConfig, to silence the `nil SecureConfig` error. ([#2567](https://github.com/turbot/steampipe/issues/2567))
* Optimise autocomplete by only loading completions on startup or when connection config changes, rather than every time a query is entered . ([#2561](https://github.com/turbot/steampipe/issues/2561))
* Remove explicit setting of open-file limit, now that Go 1.19 does it automatically. ([#2630](https://github.com/turbot/steampipe/issues/2630))

_Bug fixes_
* Update `GetPathKeys` to treat key columns with `AnyOf` require property with the same precedence as `Required`. ([#254](https://github.com/turbot/steampipe-postgres-fdw/issues/254))
* Remove blank lines in CSV and JSON query results ([#2333](https://github.com/turbot/steampipe/issues/2333), [#2340](https://github.com/turbot/steampipe/issues/2340))
* Fix UpdateConnectionConfigs call to pass the new connection for changed connections (currently the old connection is passed). ([#2349](https://github.com/turbot/steampipe/issues/2349))
* When passing empty array as variable, cast to correct type if possible. ([#2094](https://github.com/turbot/steampipe/issues/2094))
* Fixes issue where progress bars are not sorted for plugin update. ([#2501](https://github.com/turbot/steampipe/issues/2501))
* Fix intermittent dashboard shutdown stall. ([#2328](https://github.com/turbot/steampipe/issues/2328))
* Fix connection watching only adding first changed connection config to the payload of the UpdateConnectionConfigs call. ([#2395](https://github.com/turbot/steampipe/issues/2395))
* Fix the alignment of plugin update/install outputs. ([#2417](https://github.com/turbot/steampipe/issues/2417))
* Fix timeout running `service start --dashboard` with many mods installed - increase dashboard service startup timeout to 30s. ([#2434](https://github.com/turbot/steampipe/issues/2434))
* Ensure `dashboard` and `control` return exit status zero after successful run ([#2449](https://github.com/turbot/steampipe/issues/2449), [#2447](https://github.com/turbot/steampipe/issues/2447))
* Fixes issue where steampipe requests for firewall exceptions during installation. ([#2478](https://github.com/turbot/steampipe/issues/2478))
* Fix retrieval of default user workspace. ([#2499](https://github.com/turbot/steampipe/issues/2499))
* Fix plugin-manager panic when plugin startup times out. ([#2546](https://github.com/turbot/steampipe/issues/2546))
* Fix prompt failing to show when service installation runs in interactive mode. ([#2529](https://github.com/turbot/steampipe/issues/2529))
* Validate inputs when running single dashboard. Do not upload snapshot if dashboard was cancelled. ([#2551](https://github.com/turbot/steampipe/issues/2551))
* Fixes issue where the CLI would fail to connect to local service if there are credential files in `~/.postgresql`. ([#1417](https://github.com/turbot/steampipe/issues/1417))
* Fixes issue where 'Alt` keyboard combinations would error in WSL. ([#2549](https://github.com/turbot/steampipe/issues/2549))
* Fix unintuitive errors from steampipe plugin commands when a plugin (version) is missing. ([#2361](https://github.com/turbot/steampipe/issues/2361))
* Clean up error messaging when a bad template is put in the templates dir. ([#2670](https://github.com/turbot/steampipe/issues/2670))
* Fix crash when plugin list fails to connect to database.

_Deprecations_
* Deprecate `workspace-chdir`, replace with `mod-location`. ([#2511](https://github.com/turbot/steampipe/issues/2511))


## v0.16.4 [2022-09-26]
_Bug fixes_
* Fix `Plugin.GetSchema failed - no connection name passed and multiple connections loaded` error - update FDW to fix packaging issue affecting Arm Linux. ([#2464](https://github.com/turbot/steampipe/issues/2464))

## v0.16.3 [2022-09-17]
_Bug fixes_
* Fix dashboard UI benchmark controls rendering a control node per control result, rather than a control node with multiple results within it. ([#2440](https://github.com/turbot/steampipe/issues/2440))
* Fix `double` qual values not being passed to plugin. ([#243](https://github.com/turbot/steampipe-postgres-fdw/issues/243))

## v0.16.2 [2022-09-15]
_Bug fixes_
* Update FDW to not start scan until the first time IterateForeignScan is called. ([#237](https://github.com/turbot/steampipe-postgres-fdw/issues/237))
* Fix database initialisation failures due to invalid locale. ([#2368](https://github.com/turbot/steampipe/issues/2368))
* Use ellipsis char instead of 3 dots in plugin update/install when cutting off the plugin name. ([#2355](https://github.com/turbot/steampipe/issues/2355))
* Add help message for WSL1 installation failures. ([#2379](https://github.com/turbot/steampipe/issues/2379))
* Show query timing information even if query returns an error.([#2331](https://github.com/turbot/steampipe/issues/2331))
* Fix dashboard UI benchmarks with both child controls and benchmarks not rendering their controls. ([#2440](https://github.com/turbot/steampipe/issues/2440))

## v0.16.1 [2022-08-31]
_Bug fixes_
* Limit connection lifetime in the database connection pool. ([#2375](https://github.com/turbot/steampipe/issues/2375))
* Fix connection watching when multiple connection configs are changed - ensure _all_ configs are updated. ([#2395](https://github.com/turbot/steampipe/issues/2395))
* Reduce startup time when multiple mods are loaded - only create introspection tables if `STEAMPIPE_INTROSPECTION` environment variable is set. ([#2396](https://github.com/turbot/steampipe/issues/2396))

## v0.16.0 [2022-08-24]
_What's new?_
* Add support for plugin processes to handle multiple connections (rather than a process per connection), improving startup time and reducing memory usage.  ([#2262](https://github.com/turbot/steampipe/issues/2262))
* Limit the maximum memory used by the plugin query cache can using the environment variable STEAMPIPE_CACHE_MAX_SIZE_MB ([#2363](https://github.com/turbot/steampipe/issues/2363))
* Update base image for the steampipe docker container. ([#2233](https://github.com/turbot/steampipe/issues/2233))
* Improve help messages when a plugin is installed but the connection is not configured. ([#2319](https://github.com/turbot/steampipe/issues/2319))
* Only add a blank line between query results, not after the final result. ([#2333](https://github.com/turbot/steampipe/issues/2333), [#2340](https://github.com/turbot/steampipe/issues/2340))
* Timing terminal output now uses appropriate fidelity (secs, ms) for easier readability. ([#2246](https://github.com/turbot/steampipe/issues/2246))
* Disable FDW update message during plugin update. ([#2312](https://github.com/turbot/steampipe/issues/2312))
* Update dashboard `ExecutionComplete` event to include only variables referenced by the dashboard/benchmark being run. ([#2283](https://github.com/turbot/steampipe/issues/2283))
* Add support for single and multi-select combo inputs in dashboards, allowing for a combination of static/query-driven and custom options.
* Improve display of connection validation errors.
* Improve handling of dashboards with multiple inputs.
* Improve layout of dashboard error modal.

_Bug fixes_
* Fix interactive multi-line mode. ([#2260](https://github.com/turbot/steampipe/issues/2260))
* Fix intermittent failure for dashboard server shutting down when pressing ctrl+c. ([#2328](https://github.com/turbot/steampipe/issues/2328))
* Fix Steampipe terminating if query (or empty line) is entered before initialisation completes. ([#2300](https://github.com/turbot/steampipe/issues/2300))
* Fix pasting a query during CLI initialization causing it to be duplicated on the screen. ([#1980](https://github.com/turbot/steampipe/issues/1980))
* Fix connecting to remote database using `--workspace-database`. ([#2324](https://github.com/turbot/steampipe/issues/2324))

## v0.15.4 [2022-07-14]

_Bug fixes_
* Fix dashboard UI not rendering for chart/flow/hierarchy/input when type is set to table. ([#2250](https://github.com/turbot/steampipe/issues/2250))
* Fix flow/hierarchy dashboard UI bug where id/to_id and id/from_id/to_id rows would not render the expected results. ([#2254](https://github.com/turbot/steampipe/issues/2254))
* Fix FDW build issue which causes load failure on Arm Docker images.  ([#219](https://github.com/turbot/steampipe-postgres-fdw/issues/219))

## v0.15.3 [2022-07-14]
_Bug fixes_
* Fix crash when inspecting tables in interactive mode. ([#2243](https://github.com/turbot/steampipe/issues/2243))

## v0.15.2 [2022-07-13]
_Bug fixes_
* Fix intermittent hang in interactive mode if timing is enabled.  ([#2237](https://github.com/turbot/steampipe/issues/2237))

## v0.15.1 [2022-07-07]
_Bug fixes_
* Fixes various EOF query errors. ([#192](https://github.com/turbot/steampipe-postgres-fdw/issues/192), [#201](https://github.com/turbot/steampipe-postgres-fdw/issues/201), [#207](https://github.com/turbot/steampipe-postgres-fdw/issues/207))
* Ensure DashboardChanged events are generated when child elements have a changed index within a container. ([#2228](https://github.com/turbot/steampipe/issues/2228))
* Fix incorrectly identified changed inputs in DashboardChanged events. ([#2221](https://github.com/turbot/steampipe/issues/2221))
* Fix dashboard UI crashing when socket connection reconnects. ([#2224](https://github.com/turbot/steampipe/issues/2224))
* Fix intermittent "concurrent map access" error when timing is enabled. ([#2231](https://github.com/turbot/steampipe/issues/2231))

## v0.15.0 [2022-06-23]
_What's new?_
* Add support for Open Telemetry. ([#1193](https://github.com/turbot/steampipe/issues/1193))
* Update `.timing` output to return additional query metadata such as the number of hydrate functions called andd the cache status. ([#2192](https://github.com/turbot/steampipe/issues/2192))
* Add `steampipe_command.scan_metadata` table to support returning additional data from `.timing` command.  ([#203](https://github.com/turbot/steampipe-postgres-fdw/issues/203))
* Update postgres config to enable auto-vacuum. ([#2083](https://github.com/turbot/steampipe/issues/2083))
* Add `--show-password` CLI arg to reveal the db user password. Disables password visibility by default. ([#2033](https://github.com/turbot/steampipe/issues/2033)) 
* Update dashboard snapshot format, making control/benchmark output consistent with dashboards. ([#2154](https://github.com/turbot/steampipe/issues/2154)) 
* Support optional names for dashboard child blocks. ([#2161](https://github.com/turbot/steampipe/issues/2161))
* Improve the response to `steampipe plugin update all` to make it more helpful. ([#2125](https://github.com/turbot/steampipe/issues/2125))
* Add better help message when invalid locale settings caused db init failure. ([#1673](https://github.com/turbot/steampipe/issues/1673))
* Update json control output template to use Go templating, rather than just serialising the results. ([#2163](https://github.com/turbot/steampipe/issues/2163))

_Bug fixes_
* Add control severity in the check run CSV output. ([#2083](https://github.com/turbot/steampipe/issues/2083))
* Ensure prompt is shown after installing updated FDW. ([#2101](https://github.com/turbot/steampipe/issues/2101))
* Fix nil pointer error when empty array passed as variable value. ([#2094](https://github.com/turbot/steampipe/issues/2094))
* Fix interactive query failing with EOF error if the history.json is empty. ([#2151](https://github.com/turbot/steampipe/issues/2151))
* Update autocomplete description for `.output` to include `line` as an option. ([#2142](https://github.com/turbot/steampipe/issues/2142))
* Fix issue where check/templates were not getting updated even when the template file has been updated. ([#2180](https://github.com/turbot/steampipe/issues/2180))
* Fix `check all` so it does not runs controls/benchmarks from dependency mods. ([#2182](https://github.com/turbot/steampipe/issues/2182))

## v0.14.6 [2022-05-25]
_Bug fixes_
* Fix update check failing for large numbers of plugins, with little or no feedback on the error. ([#2118](https://github.com/turbot/steampipe/issues/2118))
* Fix database startup failure with `EOF` error on Mac M1 after updating FDW. ([#2116](https://github.com/turbot/steampipe/issues/2116))
* Fix intermittent `Unrecognized remote plugin message` error on Mac M1 after updating a plugin which has been locally built. Closes ([#2123](https://github.com/turbot/steampipe/issues/2123))

## v0.14.5 [2022-05-23]
_Bug fixes_
* Add support for setting dependent mod variable values using an spvars file or by setting the `Args` property in the mod `Require` block. ([#2076](https://github.com/turbot/steampipe/issues/2076), [#2077](https://github.com/turbot/steampipe/issues/2077))
* Add support for JSONB quals. ([#185](https://github.com/turbot/steampipe-postgres-fdw/issues/185))
* Fix pasting a query during CLI initialization causing it to be duplicated on the screen. ([#1980](https://github.com/turbot/steampipe/issues/1980))
* Remove limit of 2 decodes - execute as many passes as needed (as long as the number of unresolved dependencies decreases). Fixes intermittent dependency error when loading steampipe-mod-ibm-insights. ([#2062](https://github.com/turbot/steampipe/issues/2062))
* Fix workspace lock file not being correctly migrated. ([#2069](https://github.com/turbot/steampipe/issues/2069))
* Fix intermittent panic error on plugin install. ([#2069](https://github.com/turbot/steampipe/issues/2069))
* Fix nil pointer error when an empty array passed as variable value. ([#2094](https://github.com/turbot/steampipe/issues/2094))
* When running `steampipe service start --dashboard`, ensure `--workspace-chdir` arg is respected. ([#2103](https://github.com/turbot/steampipe/issues/2103))


## v0.14.4 [2022-05-12]
_Bug fixes_
* Fix ctrl+c during dashboard execution causing a `panic: send on closed channel`. ([#2048](https://github.com/turbot/steampipe/issues/2048))
* Fix backward compatibility issues in config file migration which could cause the plugin `versions.json` to become corrupted. ([#2042](https://github.com/turbot/steampipe/issues/2042))
* Fix `backups` folder is being created even if no database backup is taken. ([#2049](https://github.com/turbot/steampipe/issues/2049))
* If updated db package with same Postgres version is detected, install binaries without doing a full db install. ([#2038](https://github.com/turbot/steampipe/issues/2038))
* Fix dashboard UI benchmark nodes collapsing during running. ([#2045](https://github.com/turbot/steampipe/issues/2045))

## v0.14.3 [2022-05-10]
_Bug fixes_
* Fix a regression in v0.14.2 that would prevent migration of public schema data during migration from v0.14.x versions.  ([#2034](https://github.com/turbot/steampipe/issues/2034))

## v0.14.2 [2022-05-10]
_Bug fixes_
* When initialising the database, check whether the ImageRef of the currently installed database is correct and if not, reinstall. This provides a mechanism to force a db package update even if the Postgres version has not changed. ([#2026](https://github.com/turbot/steampipe/issues/2026))
* Ensure `Digest` payload field is not empty when calling VersionCheck endpoint. This is to handle a potential config migration bug which can result in empty `image_digest` fields in the plugin versions state file. ([#2030](https://github.com/turbot/steampipe/issues/2030))
* Fix prepared statement creation failure when installing a fresh db from a mod folder. ([#2028](https://github.com/turbot/steampipe/issues/2028))
* Limit the number of database backups as part of the daily cleanup. ([#2012](https://github.com/turbot/steampipe/issues/2012))

## v0.14.1 [2022-05-09]
_Bug fixes_
* Check if a previous version of Steampipe has a service running, and fail gracefully if so.
  If we fail to detect as service, but find a postgres process running in the install dir, kill it before migrating data. ([#2022](https://github.com/turbot/steampipe/issues/2022))

## v0.14.0 [2022-05-09]
_What's new?_
* Support real-time running and viewing of benchmarks in the dashboard UI with drill-down through benchmarks and controls to individual resource results. ([#1760](https://github.com/turbot/steampipe/issues/1760))
* Update database version to Postgresql 14. ([#43](https://github.com/turbot/steampipe/issues/43))
* Add native support for Arm architecture machines. ([#253](https://github.com/turbot/steampipe/issues/253))
* Update Go to 1.18. ([#1783](https://github.com/turbot/steampipe/issues/1783))
* Migrate all json config files to use snake case property names. ([#1730](https://github.com/turbot/steampipe/issues/1730))
* Add `input` flag to disable interactive prompting for variables. ([#1839](https://github.com/turbot/steampipe/issues/1839))
* Add `variable list` command. ([#1868](https://github.com/turbot/steampipe/issues/1868))
* Allow dependent mods to have the same variable name as the parent mod. ([#1922](https://github.com/turbot/steampipe/issues/1922))
* Update Dockerfile for postgres 14, and to disable telemetry. ([#1941](https://github.com/turbot/steampipe/issues/1941))
* Update the output and performance of plugin operations. ([#1780](https://github.com/turbot/steampipe/issues/1780), [#1778](https://github.com/turbot/steampipe/issues/1778), [#1777](https://github.com/turbot/steampipe/issues/1777), [#1776](https://github.com/turbot/steampipe/issues/1776)) 
* Rename folder .steampipe/report/assets to .steampipe/dashboard/assets. ([#1751](https://github.com/turbot/steampipe/issues/1751))
* Add `Alias` property to the dependencies listed in .mod.cache.json. ([#1731](https://github.com/turbot/steampipe/issues/1731))

_Bug fixes_
* Fix issue preventing dashboard UI from displaying in Safari ([#1984](https://github.com/turbot/steampipe/issues/1984))
* Fix intermittent "relation not found errors", when running dashboards. ([#1919](https://github.com/turbot/steampipe/issues/1919))
* Update 'check' and 'dashboard' command to NOT fail if any connection fails to load. ([#1885](https://github.com/turbot/steampipe/issues/1885))
* Update mod parsing to pass variable values to dependent mods. ([#1694](https://github.com/turbot/steampipe/issues/1694))
* Update control running to retry acquireSession in case of error, and report error in case of failure. ([#1951](https://github.com/turbot/steampipe/issues/1951))
* Fix required Steampipe version in mod.sp not being respected when running query command. ([#1734](https://github.com/turbot/steampipe/issues/1734))
* Fix dashboard cancellation is stalling when the dashboard has no children. ([#1837](https://github.com/turbot/steampipe/issues/1837))
* Fix interactive query Initialisation hang when no plugins are installed. ([#1860](https://github.com/turbot/steampipe/issues/1860))
* Escape quotes in all postgres object names. ([#1893](https://github.com/turbot/steampipe/issues/1893))
* Fixes issue where plugin install crashes for non-existent plugins. ([#1896](https://github.com/turbot/steampipe/issues/1896))
* Fix execution of dashboards causing a hang after a change or recovering from workspace error. ([#1907](https://github.com/turbot/steampipe/issues/1907))
* Fix JSON data with \u0000 errors in Postgres with "unsupported Unicode escape sequence". ([#118](https://github.com/turbot/steampipe-postgres-fdw/issues/118))
* Update dashboards to handle ExecutionError events. ([#1997](https://github.com/turbot/steampipe/issues/1997))
* Fixes issue where `service stop` command outputs "service stopped" even if no services were actually running. ([#1456](https://github.com/turbot/steampipe/issues/1456))

## v0.13.6 [2022-04-14]
_Bug fixes_
* Update dashboard UI to use wss when the location protocol is https. ([#1717](https://github.com/turbot/steampipe/issues/1717))
* Fix interactive query initialisation hang when no plugins are installed. ([#1860](https://github.com/turbot/steampipe/issues/1860))
* Fixes issue where `steampipe query` was always using a default port. ([#1753](https://github.com/turbot/steampipe/issues/1753))

## v0.13.5 [2022-04-01]
_Bug fixes_
* Ensure the search path is escaped. ([#1770](https://github.com/turbot/steampipe/issues/1770))

## v0.13.4 [2022-03-31]
_What's new?_
* Add `ShortName` property to the dependencies listed in .mod.cache.json. ([#1731](https://github.com/turbot/steampipe/issues/1731))

_Bug fixes_
* Fix setting search path after connection config changed event. ([#1700](https://github.com/turbot/steampipe/issues/1700))
* Fixes issue where tags and dimensions are not sorted in output of `check` command. ([#1715](https://github.com/turbot/steampipe/issues/1715))
* Fix required Steampipe version in mod.sp not being validated when running `query` command. ([#1734](https://github.com/turbot/steampipe/issues/1734))

## v0.13.3 [2022-03-21]
_Bug fixes_
* Fix issue where dashboard starts up even if there are initialization errors (for example unmet dependencies). ([#1711](https://github.com/turbot/steampipe/issues/1711))

## v0.13.2 [2022-03-18]
_Bug fixes_
* Fix dashboard shutdown sometimes stalling. ([#1708](https://github.com/turbot/steampipe/issues/1708))

## v0.13.1 [2022-03-17]
_What's new?_
* Improve recording of browser history in dashboard UI. ([#1633](https://github.com/turbot/steampipe/issues/1633))
* Improve template rendering performance in dashboard UI. ([#1646](https://github.com/turbot/steampipe/issues/1646))
* Add linking support to cards in dashboard UI.  ([#1651](https://github.com/turbot/steampipe/issues/1651))
* Add support for `--search-path`, `--search-path-prefix`, `--var` and `--var-file` flags to `dashboard` command. ([#1674](https://github.com/turbot/steampipe/issues/1674))
* Add ability to define static card label and value in HCL. ([#1695](https://github.com/turbot/steampipe/issues/1695))
* Add feedback during workspace load in `dashboard` command. ([#1567](https://github.com/turbot/steampipe/issues/1567))

_Bug fixes_
* Fix excessive memory usage intialising a high number of connections. ([#1656](https://github.com/turbot/steampipe/issues/1656))
* Fix issue where service was not shut down if command is cancelled during initialisation. ([#1288](https://github.com/turbot/steampipe/issues/1288))
* Fix issue where installing a plugin from any `stream` other than `latest` did not install the default `config` file. ([#1660](https://github.com/turbot/steampipe/issues/1660))
* Fix query argument resolution not working correctly when some args are provided by HCL and some from runtime args. ([#1661](https://github.com/turbot/steampipe/issues/1661))
* Fix issue where legacy `requires` property was not evaluating in mods. ([#1686](https://github.com/turbot/steampipe/issues/1686))

## v0.13.0 [2022-03-10]
_What's new?_
* Add `steampipe dashboard` command ([#1364](https://github.com/turbot/steampipe/issues/1364))
* Add `--dashboard` option to `steampipe service` command.  ([#1472](https://github.com/turbot/steampipe/issues/1472))
* Add support for `ltree` columns. ([#157](https://github.com/turbot/steampipe-postgres-fdw/issues/157))
* Add support for `inet` columns. ([#156](https://github.com/turbot/steampipe-postgres-fdw/issues/156))
* Add support for finding the mod definition by searching up the working directory tree. ([#1533](https://github.com/turbot/steampipe/issues/1533))
* Update OCI download to use a tmp folder underneath the destination folder. ([#1545](https://github.com/turbot/steampipe/issues/1545))
* Disable update checks running for plugin update command. ([#1470](https://github.com/turbot/steampipe/issues/1470))

_Bug fixes_
* Fix connection file watching. ([#1469](https://github.com/turbot/steampipe/issues/1469))
* Fix `.inspect` command for steampipe cloud connections. ([#1497](https://github.com/turbot/steampipe/issues/1497))
* Fix plugin validation error sometimes causing Steampipe to crash. ([#1387](https://github.com/turbot/steampipe/issues/1387), [#146](https://github.com/turbot/steampipe-postgres-fdw/issues/146))
* Fix plugin validation errors not being displayed as warnings on startup. ([#1413](https://github.com/turbot/steampipe/issues/1413))
* Fix workspace event handler causing freeze during initialisation. ([#1428](https://github.com/turbot/steampipe/issues/1428))
* Fix duplicate resources not being reported during mod load. ([#1477](https://github.com/turbot/steampipe/issues/1477))
* Fix interactive query cancellation only working once.([#1625](https://github.com/turbot/steampipe/issues/1625))
* Fix failure to detect duplicate pseudo resources. ([#1478](https://github.com/turbot/steampipe/issues/1478))
* Fix refreshing an aggregate connection causing a plugin crash. ([#1537](https://github.com/turbot/steampipe/issues/1537))
* Ensure SetConnectionConfig is only called once. ([#1368](https://github.com/turbot/steampipe/issues/1368))
* Fix 'is nil' qual causing a plugin crash. ([#154](https://github.com/turbot/steampipe-postgres-fdw/issues/154))
* Update plugin manager to remove plugin from map if startup fails. Prevents timeout when retrying to start a failed plugin. ([#1631](https://github.com/turbot/steampipe/issues/1631))
* Fix issue where plugin-manager becomes unstable if plugins crash. ([#1453](https://github.com/turbot/steampipe/issues/1453))

## v0.12.2 [2022-01-27]
_Bug fixes_
* Fix occasional `Unrecognized remote plugin message` errors on startup when running update checks. ([#1354](https://github.com/turbot/steampipe/issues/1354))

## v0.12.1 [2022-01-22]
_Bug fixes_
* When running queries with `csv` output, "loading results..." remains on screen after displaying results. ([#1340](https://github.com/turbot/steampipe/issues/1340))

## v0.12.0 [2022-01-20]
_What's new?_
* Update `check` to support template based export and output formats. ([#1289](https://github.com/turbot/steampipe/issues/1289))
* Add new check output format: `asff` (AWS Security Finding Format). ([#1305](https://github.com/turbot/steampipe/issues/1305))
* Add new check output format: `nunit3`. ([#1196](https://github.com/turbot/steampipe/issues/1196))

_Bug fixes_
* Fixes issue where plugins, FDW and Postgres were logging using a different timestamp formats. Now all timestamps use `UTC` ([#927](https://github.com/turbot/steampipe/issues/927))

## v0.11.2 [2022-01-10]
_Bug fixes_
* Fix issue where `steampipe check` table output only displays the summary. ([#1300](https://github.com/turbot/steampipe/issues/1300))

## v0.11.1 [2022-01-06]
_Bug fixes_
* Plugin instantiation failures should be reported as warnings not errors. ([#1283](https://github.com/turbot/steampipe/issues/1283))
* Fix issue where database name is not printed in output of `steampipe service start`. ([#1270](https://github.com/turbot/steampipe/issues/1270))
* Fix issue where service is not shutdown if interrupted while interactive prompt is initialising. ([#1004](https://github.com/turbot/steampipe/issues/1004))
* Add support for installer to detect running service when upgrading. ([#1269](https://github.com/turbot/steampipe/issues/1269))

## v0.11.0 [2021-12-21]
_What's new?_
* Add support for mod management commands: `mod install`, `mod update`, `mod uninstall`, `mod list`, `mod init`. ([#442](https://github.com/turbot/steampipe/issues/442), [#443](https://github.com/turbot/steampipe/issues/443))
* Startup optimizations.   
  * When retrieving plugin schema, identify the minimum set of schemas we need to fetch - to allow for multiple connections with the same schema. ([#1183](https://github.com/turbot/steampipe/issues/1183))
  * Avoid retrieving schema from database for check and non-interactive query execution. 
  * Update plugin manager to instantiate plugins in parallel.
  * Only create prepared statements if the query has parameters.  ([#1231](https://github.com/turbot/steampipe/issues/1231))
  * Update Postgres driver to `pgx`. (This removes the need to query the database for the db connection Pid every time we execute a query.)  ([#1179](https://github.com/turbot/steampipe/issues/1179))
  * Update connection management to use file modified time instead of filehash to detect connection changes. ([#1186](https://github.com/turbot/steampipe/issues/1186))
* Show query timing at the end of the query results. ([#1177](https://github.com/turbot/steampipe/issues/1177))
* Update workspace-database argument to handle connection strings starting with both `postgres` and `postgresql`. ([#1199](https://github.com/turbot/steampipe/issues/1199))
* Enables the `tablefunc` extension for the Steampipe database. ([#1154](https://github.com/turbot/steampipe/issues/1154))
* Improve plugin uninstall output when connections remain.  ([#1158](https://github.com/turbot/steampipe/issues/1158))
* Disable progress when running in a non-tty environment. ([#1210](https://github.com/turbot/steampipe/issues/1210))
* Bump Go to 1.17
* Add support for protoc-gen-go-grpc 1.1.0_2

_Changed Behaviour_
* Only load pseudo-resources if there is a modfile in the workspace folder. (Note - a modfile can be created by running `steampipe mod init`). ([#1238](https://github.com/turbot/steampipe/issues/1238))

_Bug fixes_
* Update database planning code give required key columns a lower cost than than optional key columns. Fixes some complex queries with `in` clauses. ([#116](https://github.com/turbot/steampipe-postgres-fdw/issues/116), [#117](https://github.com/turbot/steampipe-postgres-fdw/issues/117), [#124](https://github.com/turbot/steampipe-postgres-fdw/issues/124))
* Fix issue where `local` plugins are not evaluated as `local` as given in docs. ([#1176](https://github.com/turbot/steampipe/issues/1176))
* Fix nil reference exception during refresh connections when using dynamic plugins. ([#1223](https://github.com/turbot/steampipe/issues/1223))
* Fix issue where running service had to be stopped to install in a new install-dir. ([#1216](https://github.com/turbot/steampipe/issues/1216))
* Fix warning not being shown when running 'steampipe check'. ([#1229](https://github.com/turbot/steampipe/issues/1229))

## v0.10.0 [2021-11-24]
_What's new?_
* Add support for parallel control execution. ([#1001](https://github.com/turbot/steampipe/issues/1001))
  * Only spawn a single plugin per steampipe connection, no matter how many db connections use it. 
  * Share a single query result cache between multiple database connections. 
* Add support for connecting to a remote database, including a Steampipe Cloud workspace database.  ([#1175](https://github.com/turbot/steampipe/issues/1175))
* When cli displays error messages from plugins, they are now be prefixed with plugin name. ([#1071](https://github.com/turbot/steampipe/issues/1071))
* Do not show plugin error messages in JSON/CSV output. ([#1110](https://github.com/turbot/steampipe/issues/1110))
* Provider more responsive feedback for control runs. ([#1101](https://github.com/turbot/steampipe/issues/1101))
* Create prepared statements one by one to allow accurate error reporting and reduce memory burden. ([#1148](https://github.com/turbot/steampipe/issues/1148))
* Improve display of asyncronous error in interactive prompt. ([#1085](https://github.com/turbot/steampipe/issues/1085))
* Deprecate `workspace` argument, replace with `workspace-chdir`

_Bug fixes_
* Table names with special characters are now escaped correctly in auto-complete and `.inspect`. ([#1109](https://github.com/turbot/steampipe/issues/1109))
* Fix reflection error when loading a workspace from a hidden folder. ([#1157](https://github.com/turbot/steampipe/issues/1157))
* Fix intermittent crash when using boolean quals on jsonb columns. ([#122](https://github.com/turbot/steampipe-postgres-fdw/issues/122))

## v0.9.1 [2021-11-11]
_Bug fixes_
* Escape schema names when dropping connection schema. ([#1074](https://github.com/turbot/steampipe/issues/1074))
* Add support for quoted arguments with whitespace in query meta-commands (e.g. `.inspect`). ([#1067](https://github.com/turbot/steampipe/issues/1067))
* Fix issue where Postgres usernames weren't getting escaped properly when setting search path. ([#1094](https://github.com/turbot/steampipe/issues/1094)).
* Add support to fall back to `more` (if available) where `less` is not available in the environment. ([#1072](https://github.com/turbot/steampipe/issues/1072))
* Non-turbot plugin installs now show link to documentation. ([#1075](https://github.com/turbot/steampipe/issues/1075))
* Constrain check table-output rendering to a minimum width to avoid rendering crashes. ([#1062](https://github.com/turbot/steampipe/issues/1062))
* `steampipe check --dry-run` should not display control summary. ([#1053](https://github.com/turbot/steampipe/issues/1053))

## v0.9.0 [2021-10-24]
_What's new?_
* Update `check` command to support `markdown` and `HTML` output. ([#480](https://github.com/turbot/steampipe/issues/480), [#1011](https://github.com/turbot/steampipe/issues/1011))
* Add support for plugins with dynamic schema - reload plugin schema on startup. ([#1012](https://github.com/turbot/steampipe/issues/1012))
* Add `steampipe_reference` introspection table. ([#972](https://github.com/turbot/steampipe/issues/972))
* Add `steampipe_variable` reflection table. ([#859](https://github.com/turbot/steampipe/issues/859))
* Add `check` summary in `table` output. ([#710](https://github.com/turbot/steampipe/issues/710))
* Update DateTime and Timestamp columns to use "timestamp with time zone", not "timestamp". ([#94](https://github.com/turbot/steampipe-postgres-fdw/issues/94))
* Add support for setting a custom database name when installing. ([#936](https://github.com/turbot/steampipe/issues/936))
* Support JSON and YAML connection config. ([#969](https://github.com/turbot/steampipe/issues/969))
* Allow plugin uninstall even if there are active connections. ([#852](https://github.com/turbot/steampipe/issues/852))
* Control results are now ordered by status.  ([465](https://github.com/turbot/steampipe/issues/465))
* Add support for SSL certificate validation and rotation. ([#1020](https://github.com/turbot/steampipe/issues/1020))
* Remove deprecated flags `--db-listen` and `--db-port` from service start. ([#582](https://github.com/turbot/steampipe/issues/582))

_Bug fixes_
* Plugin commands now exit with a non-zero code on error. ([#980](https://github.com/turbot/steampipe/issues/980))
* Fix for incorrect message from service status when service is not running. ([#975](https://github.com/turbot/steampipe/issues/975))
* Update introspection tables to ensure naming consistency - fix mods and pseudo resources to remove type prefix. ([#959](https://github.com/turbot/steampipe/issues/959))
* Fix for plugin list failing with 'invalid memory address'. ([#984](https://github.com/turbot/steampipe/issues/984))


## v0.8.5 [2021-10-07]
_Bug fixes_
* Fix handling of null unicode chars in JSON fields. ([#102](https://github.com/turbot/steampipe-postgres-fdw/issues/102))
* Fix issue where queries with a`limit` clause not always listing all results. Only pass the limit to the plugin if all quals are supported by plugin `key columns`. [#103](https://github.com/turbot/steampipe-postgres-fdw/issues/103))

## v0.8.4 [2021-09-29]
_Bug fixes_
* Update client error handling to only refresh session data for a 'context deadline exceeded' error. This avoids recursion in the error handling. ([#970](https://github.com/turbot/steampipe/issues/970))

## v0.8.3 [2021-09-28]

_What's new?_
* Update `service start` command to support `database-password` arg and `STEAMPIPE_DATABASE_PASSWORD` environment variable, to allow a custom password to be used when running in service mode. ([#725](https://github.com/turbot/steampipe/issues/725))
* Small updates to output of `steampipe service` commands. ([#812](https://github.com/turbot/steampipe/issues/812))
* Add support for piping `stdout` and `stderr` from `service start` to the `TRACE log`.  ([#810](https://github.com/turbot/steampipe/issues/810))

_Bug fixes_
* Update Docker image to remove password file. ([#957](https://github.com/turbot/steampipe/issues/957))
* Fix filewatching to ensure prepared statements are correctly created and updated to reflect SQL file changes. ([#901](https://github.com/turbot/steampipe/issues/901))
* Ensure session data is restored after a SQL client error. Reset SQL client after a failure to create a transaction. ([#939](https://github.com/turbot/steampipe/issues/939))
* Fix service lifecycle management issues when state file is deleted while service is running. ([#872](https://github.com/turbot/steampipe/issues/872))
* Fix issue where `service stop` shuts down service even if non-Steampipe clients are connected. ([#887](https://github.com/turbot/steampipe/issues/887))
* Fix connection config not being passed when instantiating plugins to retrieve their schema. This resulted in descriptions not being shown for dynamic tables dynamic tables. ([#932](https://github.com/turbot/steampipe/issues/932))
* Fix issue where `install.sh` fails for IPv6 enabled system. ([#861](https://github.com/turbot/steampipe/issues/861))

## v0.8.2 [2021-09-14]
_Bug fixes_
* Fix nil pointer error when running a fully qualified query (i.e. including mod name). ([#902](https://github.com/turbot/steampipe/issues/902))

## v0.8.1 [2021-09-12]
_Bug fixes_
* Disable database log polling, which was causing high CPU usage. 
* Fix null reference exception for certain `is null` queries. ([#97](https://github.com/turbot/steampipe-postgres-fdw/issues/97)) 
* Add support for CIDROID type when converting Postgres datums to qual values. ([#54](https://github.com/turbot/steampipe-postgres-fdw/issues/54))
* Fix autocomplete casing for .cache metacommands. ([#875](https://github.com/turbot/steampipe/issues/875))

## v0.8.0 [2021-09-09]
_What's new?_
* Add HCL support for variables. ([#754](https://github.com/turbot/steampipe/issues/754))
* Add HCL support for passing parameters to queries. ([#802](https://github.com/turbot/steampipe/issues/802))
* Add `completion` command providing completion support for bash, zshell and fish. ([#481](https://github.com/turbot/steampipe/issues/481))
* Add `.cache` metacommand to control the FDW cache from the interactive prompt. ([#688](https://github.com/turbot/steampipe/issues/688))
* Remove hardcoded Postgres runtime flags by adding defaults to postgresql.conf ([#767](https://github.com/turbot/steampipe/issues/767))
* Add support for syntax highlighting in interactive prompt. ([#64](https://github.com/turbot/steampipe/issues/64))
* Update interactive prompt to use adaptive suggestion window instead of giving `console window is too small` error. ([#712](https://github.com/turbot/steampipe/issues/712))
* Log Postgres output if database initialisation fails. ([#800](https://github.com/turbot/steampipe/issues/800))
* Various minor UI tweaks. ([#786](https://github.com/turbot/steampipe/issues/786))

_Bug fixes_
* Fix issue where the `>` prompt disappears when messages are shown from file watcher or asyncronous initialisation. ([#713](https://github.com/turbot/steampipe/issues/713))
* Fix errors during async interactive startup leaving the prompt in a bad state. ([#728](https://github.com/turbot/steampipe/issues/728))
* Fix for delay in `loading results` spinner showing, caused by asyncronous initialisation. ([#671](https://github.com/turbot/steampipe/issues/671))
* Fix for missing `control_description`, `control_title` in `csv` output of `check` command. ([#739](https://github.com/turbot/steampipe/issues/739))
* Fix for `0` exit code even if `service start` fails. ([#762](https://github.com/turbot/steampipe/issues/762))
* Fix issue where configs referring to unavailable plugin will display incorrect error message. ([#796](https://github.com/turbot/steampipe/issues/796))
* Mod parsing now raises an error if duplicate locals are found. ([#846](https://github.com/turbot/steampipe/issues/846))
* Fix JSON data with '\u0000' resulting in Postgres error "unsupported Unicode escape sequence". ([#93](https://github.com/turbot/steampipe-postgres-fdw/issues/93))

## v0.7.3 [2021-08-18]
_Bug fixes_
* Retry a control run if the plugin crashes. ([#757](https://github.com/turbot/steampipe/issues/757))
* Restart a plugin if it exits unexpectedly. ([#89](https://github.com/turbot/steampipe-postgres-fdw/issues/89))

## v0.7.2 [2021-08-06]
_Bug fixes_
* Fix issue where interactive prompt hangs with a `;` input. ([#700](https://github.com/turbot/steampipe/issues/700))
* Fix cancellation not working when database client becomes unresponsive. ([#733](https://github.com/turbot/steampipe/issues/733))
* Prevent update checks from getting triggered for `service stop`. ([#745](https://github.com/turbot/steampipe/issues/745))
* Add `initializing` spinner while waiting for asynchronous initialization to finish. ([#671](https://github.com/turbot/steampipe/issues/671))
* Prevent `interactive prompt` from disappearing after asynchronous messages are shown. ([#713](https://github.com/turbot/steampipe/issues/713))

## v0.7.1 [2021-07-29]
_What's new?_
* Add `open_graph` property to `steampipe_mod` reflection table. ([#692](https://github.com/turbot/steampipe/issues/692))
  
_Bug fixes_
* When an aggregator connection is evaluating a wildcard, only include connections with compatible plugin type. ([#687](https://github.com/turbot/steampipe/issues/687))
* Fix search path not being honored by `steampipe check`. ([#708](https://github.com/turbot/steampipe/issues/708))
* Fix interactive console becoming unresponsive after ";" query. ([#700](https://github.com/turbot/steampipe/issues/700))
* Fix `nil pointer exception` in `steampipe plugin`. ([#678](https://github.com/turbot/steampipe/issues/678))

## v0.7.0 [2021-07-22]
_What's new?_
* Add support for aggregator connections. ([#610](https://github.com/turbot/steampipe/issues/610)) 
* Service management improvements: 
  * Remove locking from service code to allow multiple `query` and `check` sessions in parallel without requiring a service start.([#579](https://github.com/turbot/steampipe/issues/579))
  * Update service start to 'claim' a service started by query or check session, instead of failing. ([#580](https://github.com/turbot/steampipe/issues/580))
  * Update `service status` - add `--all` flag to list status for all running services.([#580](https://github.com/turbot/steampipe/issues/580))
  * Update `service start` to add `--foreground` flag. ([#535](https://github.com/turbot/steampipe/issues/535))
* Improvements for Docker:
  * Run `initdb` if database is installed but `data directory` is empty. ([#575](https://github.com/turbot/steampipe/issues/575))
  * Split `versions.json` into 2 files, one in the plugins dir, one in the database dir. ([#576](https://github.com/turbot/steampipe/issues/576))
  * Update plugin install to put temp files underneath the plugin directory. ([#600](https://github.com/turbot/steampipe/issues/600))
  * Steampipe service startup now validates that the `data-dir` is writable. ([#659](https://github.com/turbot/steampipe/issues/659))
* Optimise interactive startup by initializing asynchronously. ([#627](https://github.com/turbot/steampipe/issues/627))
* Optimise query caching - construct key based on the columns returned by the plugin, not the columns requested.([#82](https://github.com/turbot/steampipe-postgres-fdw/issues/82))
* Update Steampipe service to support SSL. ([#602](https://github.com/turbot/steampipe/issues/602)) 
* Show timer result before query output, so it is visible even if results require paging. ([#655](https://github.com/turbot/steampipe/issues/655))
* Increase length of history file to 500 entries. ([#664](https://github.com/turbot/steampipe/issues/664))

_Bug fixes_
* Do not disable pager when errors are displayed in interactive mode. ([#606](https://github.com/turbot/steampipe/issues/606))
* Fixes issue where `STEAMPIPE_INSTALL_DIR` was not being respected. ([#613](https://github.com/turbot/steampipe/issues/613))
* Fix multiple ctrl+C presses causing a crash on control runs. ([#630](https://github.com/turbot/steampipe/issues/630))
* Ensure multiline control errors are rendered in full ([#672](https://github.com/turbot/steampipe/issues/672))
* Fix crash when benchmark has duplicate children. Instead, raise a validaiton failure. ([#667](https://github.com/turbot/steampipe/issues/667))
* Fixes issue where `service stop` does not work on `Linux` systems. ([#653](https://github.com/turbot/steampipe/issues/653))
* Plugin schema validation errors should be displayed as warning, and not cause Steampipe to exit. ([#644](https://github.com/turbot/steampipe/issues/644))

## v0.6.2 [2021-07-08]
_Bug fixes_
* Revert prototype code inadvertently included in 0.6.1 

## v0.6.1 [2021-07-08]
_What's new?_
* Support executing control queries using the query command. ([#470](https://github.com/turbot/steampipe/issues/470))
* Update steampipe-plugin-sdk reference version to support ProtocolVersion `20210701`

_Bug fixes_
* Fix issue where `dimension` values were not rendered in generated CSV for `check`. ([#587](https://github.com/turbot/steampipe/issues/587))
* Fix Linux Installer script showing verification error for Amazon Linux. ([#479](https://github.com/turbot/steampipe/issues/438))
* Fix issue where using `--timing` with `check` was not showing duration. ([#571](https://github.com/turbot/steampipe/issues/571))
* Fix problem where milliseconds of timestamps were not being displayed ([#76](https://github.com/turbot/steampipe-postgres-fdw/issues/76))
* Fix  freezing issues with 'limit' and cancellation. ([#74](https://github.com/turbot/steampipe-postgres-fdw/issues/74))
* Fix incorrect caching of 'get' query results for plugins build with sdk >= 0.3.0. ([#60](https://github.com/turbot/steampipe-postgres-fdw/issues/60))
  
## v0.6.0 [2021-06-17]
_What's new?_
* Add `csv` output format to `check` command. ([#479](https://github.com/turbot/steampipe/issues/479))
* Add `--export` flag to `check` command. ([#511](https://github.com/turbot/steampipe/issues/511))
* Add `--dry-run` flag to `check` command to show which controls would be run. ([#468](https://github.com/turbot/steampipe/issues/468))
* Add `--tag` and `--where` arguments to `check` command to provide filtering of the controls which are run. ([#539](https://github.com/turbot/steampipe/issues/539))
* Update `service status` to make messaging more helpful when the service is running for a query session. ([#531](https://github.com/turbot/steampipe/issues/531))
* Update `query` to add support for reading from `STDIN`. ([#499](https://github.com/turbot/steampipe/issues/499))
* Validate that plugin versions required by the workspace mod are installed. ([#557](https://github.com/turbot/steampipe/issues/557))

_Bug fixes_
* Update `check` exit code to be the number of alerts. ([#498](https://github.com/turbot/steampipe/issues/498))
* Update check output formatting is now consistent when there is both a plugin and steampipe update.  ([#423](https://github.com/turbot/steampipe/issues/423))
* Fix failure to load SQL files from workspace folder if they include `$$` escape characters. ([#554](https://github.com/turbot/steampipe/issues/554))

## v0.5.3 [2021-06-14]
_Bug fixes_
* Fixes Steampipe failing to run when too many benchmarks use the same controls. ([#528](https://github.com/turbot/steampipe/issues/528))

## v0.5.2 [2021-06-10]
_Bug fixes_
* Ensure consistent ordering of query result cache key when more than one qual is used. ([#53](https://github.com/turbot/steampipe-postgres-fdw/issues/53))
* Fixes `check` command `json` output. ([#525](https://github.com/turbot/steampipe/issues/525))

## v0.5.1 [2021-05-27]
_What's new?_
* Update the `check` output to show the tree structure of the benchmarks and controls. ([#500](https://github.com/turbot/steampipe/issues/500))

_Bug fixes_
* Fix issue where interactive prompt sometimes hangs on cancellation. ([#507](https://github.com/turbot/steampipe/issues/507))
* Fix stack overflow error when allocating colors for large number of dimension property values. ([#509](https://github.com/turbot/steampipe/issues/509))
* Fix query result cache key being built incorrectly when more than one qual is used. ([#453](https://github.com/turbot/steampipe-postgres-fdw/issues/53))

## v0.5.0 [2021-05-20]
_What's new?_
* New `check` command, to run controls and benchmarks. ([#410](https://github.com/turbot/steampipe/issues/410), [#413](https://github.com/turbot/steampipe/issues/413))
* Add resource reflection tables `steampipe_mod`, `steampipe_query`, `steampipe_control` and `steampipe_benchmark`.  ([#406](https://github.com/turbot/steampipe/issues/406))
* Parsing of variable references, functions and locals. ([#405](https://github.com/turbot/steampipe/issues/405))
* Support for cancellation of queries and control runs.  ([#475](https://github.com/turbot/steampipe/issues/475))
  
## v0.4.3 [2021-05-13]

_Bug fixes_
* Fix cache check code incorrectly identifying a cache hit after a count(*) query.  ([#44](https://github.com/turbot/steampipe-postgres-fdw/issues/44))
* Fix spinner displaying multiple newlines if spinner text is wider than the terminal. ([#450](https://github.com/turbot/steampipe/issues/450))

## v0.4.2 [2021-05-06]

_Bug fixes_
* Make `.inspect` column headers lowercase. ([#439](https://github.com/turbot/steampipe/issues/439))
* Fix edge case where update notification may be displayed once when running in query `batch` mode, instead if being suppressed. This occurred the very first time an update check was performed. ([#428](https://github.com/turbot/steampipe/issues/428))
* When checking for SDK compatibility of loaded plugins, use the protocol version, not the SDK version. ([#453](https://github.com/turbot/steampipe/issues/453))

## v0.4.1 [2021-04-22]

_Bug fixes_
* Ensure we report an error and do not start database service if `port` is already in use. ([#399](https://github.com/turbot/steampipe/issues/399))
* Update check should not run when executing `query` command non-interactively. ([#301](https://github.com/turbot/steampipe/issues/301))

## v0.4.0 [2021-04-15]
_What's new?_
* Named query support - all SQL file in current folder (or the folder specified by the `workspace` argument) will be loaded and available to run as `named queries`. ([#369](https://github.com/turbot/steampipe/issues/369)) 
* When running in interactive mode, a file watcher is enabled for the current workspace (can be disabled using the `watch` argument or `terminal` config property). When enabled, any new or updated SQL files in the workspace will be reflected in the available named queries. ([#380](https://github.com/turbot/steampipe/issues/380)) 
* The `query` command now accepts multiple unnamed arguments, each of which may be either a filepath to a SQL file, a named query or the raw SQL of the query. ([#388](https://github.com/turbot/steampipe/issues/388)) 
* The search path for the steampipe database service may be specified using the `database` config. ([#353](https://github.com/turbot/steampipe/issues/353))
* The search path and search path prefix terminal sessions may be specified using `terminal` config, command line argument or meta-commands. ([#353](https://github.com/turbot/steampipe/issues/353),  [#357](https://github.com/turbot/steampipe/issues/358), [#358](https://github.com/turbot/steampipe/issues/358)) 

## v0.3.6 [2021-04-08]
_Bug fixes_
* Fix log trimming, which was broken by the change of log location. ([#344](https://github.com/turbot/steampipe/issues/344))
* Plugin updates should be  listed alphabetically. ([#339](https://github.com/turbot/steampipe/issues/339))

## v0.3.5 [2021-04-02]
_Bug fixes_
* Fix `.inspect` not working with unqualified table names. ([#346](https://github.com/turbot/steampipe/issues/346))

## v0.3.4 [2021-04-01]
_Bug fixes_
* Ensure that after adding a connection, search path changes are reflected in the current query session. ([#340](https://github.com/turbot/steampipe/issues/340))
* Fix extra trailing white-space issue in `line` output. ([#332](https://github.com/turbot/steampipe/issues/332))
* Remove HTML escaping from JSON output. ([#336](https://github.com/turbot/steampipe/issues/336))
* Fix issue where service is always listening on network listener. ([#330](https://github.com/turbot/steampipe/issues/330))
* Fix incorrect error message when trying to update a non-installed plugin ([#343](https://github.com/turbot/steampipe/issues/343))
* Fix the search path not being updated when removing the last connection. ([#345](https://github.com/turbot/steampipe/issues/345))

## v0.3.3 [2021-03-22]
_Bug fixes_
* Verify the `steampipe` foreign server exists when starting the database service and if it does not, re-initialise the FDW and create the server. ([#324](https://github.com/turbot/steampipe/issues/324))

## v0.3.2 [2021-03-20]
_Bug fixes_
* Remove Postgres synchronous_commit=off setting, which could cause FDW setup in Postgres to not be committed during setup (on Linux). ([#319](https://github.com/turbot/steampipe/issues/319))
* `.header` terminal setting should also affect table output. ([#312](https://github.com/turbot/steampipe/issues/312))

## v0.3.1 [2021-03-19]
_Bug fixes_
* Fix crash when doing "is (not) null" checks on JSON fields. ([#38](https://github.com/turbot/steampipe-postgres-fdw/issues/38))

## v0.3.0 [2021-03-18]
_What's new?_
* Support setting Steampipe options using a config file. ([#230](https://github.com/turbot/steampipe/issues/230))
* Add `install-dir` argument to specify location of the installation folder. ([#241](https://github.com/turbot/steampipe/issues/241))
* Improve the handling of database quals. Query restrictions are now passed the plugin for a much wider ranger of queries including joins and nested queries. ([#3](https://github.com/turbot/steampipe-postgres-fdw/issues/3))  
* Improve handling and reporting of config parsing failures. ([#307](https://github.com/turbot/steampipe/issues/307))
* Move the log location to `~/.steampipe/logs` ([#278](https://github.com/turbot/steampipe/issues/278))
* Change postgres log prefix to `database-` ([#310](https://github.com/turbot/steampipe/issues/310))
* Deprecate `db-port` and `listener` arguments, replace with `database-port` and `database-listener`. ([#302](https://github.com/turbot/steampipe/issues/302)) 

## v0.2.5 [2021-03-15]
_Bug fixes_
* Fix crash when installing a plugin after a fresh install. ([#283](https://github.com/turbot/steampipe/issues/283))
* Fix `.inspect` meta-command failure if no arguments are provided. ([#282](https://github.com/turbot/steampipe/issues/282))

## v0.2.4 [2021-03-11]
_What's new?_
* Autocomplete now includes public schema.  ([#123](https://github.com/turbot/steampipe/issues/123))
* Add bug report and feature request issue templates.  ([#266](https://github.com/turbot/steampipe/issues/266))
* Add `SECURITY.md`. ([#266](https://github.com/turbot/steampipe/issues/266))
* Update spacing for plugin update and install messages. ([#264](https://github.com/turbot/steampipe/issues/264))

_Bug fixes_
* Remove invalid update notifications for plugins which cannot be found in the registry.  ([#265](https://github.com/turbot/steampipe/issues/265))
* Fix typo in install.sh. 

## v0.2.3 [2021-03-03]
_What's new?_
* Increase timeout for plugin update HTTP call. ([#216](https://github.com/turbot/steampipe/issues/216))
* `plugin update` now checks installed version of a plugin is out of date before updating. ([#234](https://github.com/turbot/steampipe/issues/234))
* Improve the error messages for sql errors. ([#118](https://github.com/turbot/steampipe/issues/118))
* Wrap `plugin list` output to window width. ([#235](https://github.com/turbot/steampipe/issues/235))

_Bug fixes_
* Fix timestamp quals not being passed to plugin. ([#247](https://github.com/turbot/steampipe/issues/247))
* Fix `steampipe server not found` error after failed connection validation. ([#220](https://github.com/turbot/steampipe/issues/220))
* Ensure all panics are recovered. ([#246](https://github.com/turbot/steampipe/issues/246))

## v0.2.2 [2021-02-25]
_What's new?_
* Set Inspect column width to no larger than required to display data. ([#155](https://github.com/turbot/steampipe/issues/155))
* Plugin SDK version check should ignore patch and prerelease version. ([#217](https://github.com/turbot/steampipe/issues/217))
* Enforce reserved connection name ('public', 'internal'). ([#168](https://github.com/turbot/steampipe/issues/168))
* Do not allow Steampipe to run from Root. ([#167](https://github.com/turbot/steampipe/issues/167))
* `plugin update`, `plugin install` and `plugin uninstall` commands display error if no plugins specified in args. ([#199](https://github.com/turbot/steampipe/issues/199))
* Remove global `--config` flag. ([#215](https://github.com/turbot/steampipe/issues/215))

_Bug fixes_
* Fix cache retrieving incorrect data for multi-connection queries.([#223](https://github.com/turbot/steampipe/issues/223))
* Ensure search path is set for clients other than Steampipe. ([#218](https://github.com/turbot/steampipe/issues/218))
* Spinner should not be displayed in non-interactive query mode. ([#227](https://github.com/turbot/steampipe/issues/227))

## v0.2.1 [2021-02-20]
_Bug fixes_
* Ensure all hydrate errors are reported. ([#206](https://github.com/turbot/steampipe/issues/206))
* Change plugin update URL to hub.steampipe.io. ([#201](https://github.com/turbot/steampipe/issues/201))
* Steampipe version string should include 'prerelease' suffix if it is set. ([#200](https://github.com/turbot/steampipe/issues/200))
* Column headers in table output should respect casing of the column name. ([#181](https://github.com/turbot/steampipe/issues/181))

## v0.2.0 [2021-02-18]
_What's new?_
* Add support for multiregion queries. ([#197](https://github.com/turbot/steampipe/issues/197))
* Add support for connection config. ([#173](https://github.com/turbot/steampipe/issues/173))
* Add `plugin update` command. ([#176](https://github.com/turbot/steampipe/issues/176))
* Add automatic checking of plugin versions. ([#164](https://github.com/turbot/steampipe/issues/164))
* Add caching of query results. This is disabled by default but may be enabled by setting `STEAMPIPE_CACHE=true`
  NOTE: It is expected this will be updated to default to true in the next patch release. ([#11](https://github.com/turbot/steampipe-postgres-fdw/issues/11)) 
* Log whether Steampipe is running in Windows subsystem for Linux. ([#171](https://github.com/turbot/steampipe/issues/171))
* All env vars should have STEAMPIPE_ prefix. ([#172](https://github.com/turbot/steampipe/issues/172))
* Display null column values as <null> instead of an empty string. ([#186](https://github.com/turbot/steampipe/issues/186))
* Validate that plugins do not have an sdk version greater than the version steampipe is built against. ([#183](https://github.com/turbot/steampipe/issues/183))

_Bug fixes_
* Fix hitting a space after a meta-command causing runtime error. ([#182](https://github.com/turbot/steampipe/issues/182))

## v0.1.3 [2021-02-11]

_What's new?_
* Add 'line' output format. ([#114](https://github.com/turbot/steampipe/issues/114))
* Log files older than 7 days are deleted. ([#121](https://github.com/turbot/steampipe/issues/121))

_Bug fixes_
* Fix multi line editing issues. ([#103](https://github.com/turbot/steampipe/issues/103))
* Fix command-Right breaking for unicode chars ([#9](https://github.com/turbot/steampipe/issues/9))
* Fix 'no unpinned buffers available' error.  ([#122](https://github.com/turbot/steampipe/issues/122))
* Fix database installation failure for certain Linux configurations. ([#133](https://github.com/turbot/steampipe/issues/133))

## v0.1.2 [2021-02-04]

_What's new?_
* The `.inspect` command no longer requires the fully qualified name for tables. ([#21](https://github.com/turbot/steampipe/issues/21))
* The helper function `glob` has been added. ([#134](https://github.com/turbot/steampipe/issues/134))
* The output of the `plugin install` command now shows the installed version.  ([#93](https://github.com/turbot/steampipe/issues/93))
* The `.help` command now displays a link to the inline help docs.  ([#92](https://github.com/turbot/steampipe/issues/92))
* The wait spinner is now only shown in interactive mode. ([#106](https://github.com/turbot/steampipe/issues/106))

_Bug fixes_
* Fix JSON and bool columns displaying as strings. ([#95](https://github.com/turbot/steampipe/issues/95))
* Fix column headings displaying in upper case.  ([#94](https://github.com/turbot/steampipe/issues/94))

## v0.1.1 [2021-01-28]

_What's new?_
* A new meta-command `.help` has been added.  ([#54](https://github.com/turbot/steampipe/issues/54))
* After `steampipe plugin install`, a link to the plugin docs is displayed.
* A spinner is now displayed for slow queries. ([#77](https://github.com/turbot/steampipe/issues/77))
* A maximum column width of 1024 is now enforced - content longer than this will wrap. ([#12](https://github.com/turbot/steampipe/issues/12))
* The `description` column of the `.inspect` command now fills the available horizontal screen space. ([#11](https://github.com/turbot/steampipe/issues/11))
* The Linux installation package now uses tar instead of zip. ([#63](https://github.com/turbot/steampipe/issues/63))

_Bug fixes_
* Fix results paging failure for very long rows (> 64k chars). ([#75](https://github.com/turbot/steampipe/issues/75))
* Fix invalid query resulting in the database session remaining open. ([#60](https://github.com/turbot/steampipe/issues/60))
* Fix data formatting in json output. ([#14](https://github.com/turbot/steampipe/issues/14))
* Fix incorrect plugin hub link.
* Fix `steampipe query` panic when exiting after `service stopped --force` has been run. ([#38](https://github.com/turbot/steampipe/issues/38))
* Fix `runtime error: slice bounds out of range [1:0]`.  ([#40](https://github.com/turbot/steampipe/issues/40))
* Fix boolean meta-command showing wrong status when no parameter is passed. ([#48](https://github.com/turbot/steampipe/issues/48))
