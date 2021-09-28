## v0.8.3 [2021-09-27]

_What's new?_
* Update `service start` command to support `database-password` arg and `STEAMPIPE_DATABASE_PASSWORD` environment variable, to allow a custom password to be used when running in service mode. ([#725](https://github.com/turbot/steampipe/issues/725))
* Small updates to output of `steampipe service` commands.  ([#812](https://github.com/turbot/steampipe/issues/812))
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
