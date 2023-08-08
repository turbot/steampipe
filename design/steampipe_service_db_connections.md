## Queries that need to be executed over a client connection:

- Get scan metadata
  - read automatically if `--timing` is enabled
- Set search path
  - Can be automatically during session startup
  - Can be set by the user using meta commands
- Cache commands
  - Can be automatically during session startup
  - Can be set by the user using meta commands
- Introspection tables
  - Written automatically for each database connection
  - Read by system if `--tag` or `--where` are used for `check`

## Database Session

A thin wrapper around the raw database connection which caches the `search path` and the `scan metadata id`

### Acquire `session`

1. Get a database connection from the pool
1. if not found in `session cache map`
   1. create a `DatabaseSession` for the connection
   1. Persist `DatabaseSession` in `session cache map`
1. Set cache parameters (if required)
   1. If client `cache` is enabled, enable client `cache` on the connection
   1. If client `cache ttl` is set, set the `cache ttl` on the connection
1. Ensure `search path`
   1. Load the `search path` of the `steampipe` user - db query
   1. Get the resolved `search path` based on the `search_path` and `search_path_prefix` configs (`custom_search_path`)
   1. if the `loaded search path` and `resolved search path` differ, set the `resolved search path` on the connection
                                                                                                                                                                                                                                                                                                             