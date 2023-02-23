# Steampipe Caching

### Cache store
* Steampipe itself has no cache which can store cached data. 
* It depends on the SDK to cache data in plugin processes.
* Each plugin has an in-memory cache where it caches data. Right now, there's no mechanism to turn off plugin cache.

### Cache Setting (in order of precedence - highest first)
1. `<connection>.spc` -> `Connection.options.connection.Cache` -> `cache`
1. `STEAMPIPE_CACHE` environment variable
1. `default.spc` -> `options.Connection` -> `cache`

### Caching and FDW
All FDW instances and plugin processes inherit the environment of the process which starts the database `service`.

Therefore, when the database service is started with `steampipe service start`, FDW instances will lock in to the environment of the `service start` process and not subsequent commands which attach to the started service.

### Tweaking cache setting per FDW

The FDW exposes a `cache` table in the `steampipe_command` schema which is used to control the cache settings of an FDW instance.

`INSERT INTO steampipe_command.cache (operation) values ('on')` -> turns the override cache on
`INSERT INTO steampipe_command.cache (operation) values ('off')` -> turns the override cache off
`INSERT INTO steampipe_command.cache (operation) values ('clear')` -> sets the minimum acceptable cache entry time to now

The cache command handling can be referred in `fdw/hub/hub.go#HandleCacheCommand`

The cache is sent in `fdw/hub/hub.go#StartScan`

### Truth Table:

| `default value` | `STEAMPIPE_CACHE` | `options.Connection.Cache` | `<connection>.options.Connection.Cache` | expect cached | actual |
|------|-------|-------|-------|-------|-------|
| TRUE | -     | -     | -     | TRUE  | TRUE  |
| TRUE | -     | -     | TRUE  | TRUE  | TRUE  |
| TRUE | -     | -     | FALSE | FALSE | FALSE |
| TRUE | -     | TRUE  | -     | TRUE  | TRUE  |
| TRUE | -     | TRUE  | TRUE  | TRUE  | TRUE  |
| TRUE | -     | TRUE  | FALSE | FALSE | FALSE |
| TRUE | -     | FALSE | -     | FALSE | FALSE |
| TRUE | -     | FALSE | TRUE  | TRUE  | TRUE  |
| TRUE | -     | FALSE | FALSE | FALSE | FALSE |
| TRUE | TRUE  | -     | -     | TRUE  | TRUE  |
| TRUE | TRUE  | -     | TRUE  | TRUE  | TRUE  |
| TRUE | TRUE  | -     | FALSE | TRUE  | FALSE |
| TRUE | TRUE  | TRUE  | -     | TRUE  | TRUE  |
| TRUE | TRUE  | TRUE  | TRUE  | TRUE  | TRUE  |
| TRUE | TRUE  | TRUE  | FALSE | TRUE  | FALSE |
| TRUE | TRUE  | FALSE | -     | TRUE  | TRUE  |
| TRUE | TRUE  | FALSE | TRUE  | TRUE  | TRUE  |
| TRUE | TRUE  | FALSE | FALSE | TRUE  | FALSE |
| TRUE | FALSE | -     | -     | FALSE | FALSE |
| TRUE | FALSE | -     | TRUE  | FALSE | TRUE  |
| TRUE | FALSE | -     | FALSE | FALSE | FALSE |
| TRUE | FALSE | TRUE  | -     | FALSE | FALSE |
| TRUE | FALSE | TRUE  | TRUE  | FALSE | TRUE  |
| TRUE | FALSE | TRUE  | FALSE | FALSE | FALSE |
| TRUE | FALSE | FALSE | -     | FALSE | FALSE |
| TRUE | FALSE | FALSE | TRUE  | FALSE | TRUE  |
| TRUE | FALSE | FALSE | FALSE | FALSE | FALSE |

