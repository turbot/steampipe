# Search Path

## Configuring the search path

Search path config precedence
- The session setting, as set by the most recent `.search_path` and/or `.search_path_prefix` meta-command (for interactive session).
- The `--search-path` or `--search-path-prefix` command line arguments.
- The `search_path` or `search_path_prefix` set in the terminal options for the workspace, in the workspace.spc file.
- The `search_path` or `search_path_prefix` set in the terminal global option, typically set in ~/.steampipe/config/default.spc
- The `search_path` or `search_path_prefix` set in the database global option, typically set in ~/.steampipe/config/default.spc
- The compiled default (public, then alphabetical by connection name)

## Implementation
### Overview

During initialisation, the `requiredSessionSearchPath` for a `DbClient` is set by `RefreshConnectionsAndSearchPaths`. 

When a DB session is created, the session search path is set to the `requiredSessionSearchPath` in `DbClient.ensureSessionSearchPath()`

Finally, call `LoadForeignSchemaNames` which updates the client `foreignSchemas` property with a list of foreign schema

### RefreshConnectionAndSearchPaths implementation
`LocalDbClient.RefreshConnectionAndSearchPaths` simplified, does this:
```
refreshConnections()
setUserSearchPath()
SetSessionSearchPath()
```
#### setUserSearchPath
This function sets the search path for all steampipe users of the db service.
We do this so that the search path is set even when connecting to the DB from a non Steampipe client.
(When using Steampipe to connect to the DB, it is the Session search path which is respected.)

It does this by finding all users assigned to the role `steampipe_users` and setting their search path.

To determine the search path to set, it checks whether the `search-path` config is set.
- If set, it uses the configured value (with "internal" at the end)
- If not, it calls `getDefaultSearchPath` which builds a search path from the connection schemas, bookended with `public` and `internal`.


#### SetRequiredSessionSearchPath
This function populates the `requiredSessionSearchPath` property on the client.
This will be used during session initialisation to actually set the search path

In order to construct the required search path, `ContructSearchPath` is called

#### ContructSearchPath
- If a custom search path has been provided, prefix this with the search path prefix (if any) and suffix with `internal`
- Otherwise use the default search path, prefixed with the search path prefix (if any)

If either a `search-path` or `search-path-prefix` is set in config, this sets the search path
(otherwise fall back to the user search path set in setUserSearchPath`)    


## Responding to runtime search path changes
The search path setting in the `database` or `terminal` options may be changed while the steampipe service is running. 

The result currently depends on what is running.

### Steampipe DB service
If the steampipe DB service is running and search path options are changed in the `database` or `terminal` options, 
the updated search path will be reflected in any _new_ Steampipe interactive sessions. 
(New sessions using other DB clients will reflect changes in the `database config only)   

### Interactive session
If an interactive session (or third paty client session) is running, changes to the search path options _will not_ be
reflected in the current session.
 
### Dasboard Service
If the dashboard service is running, changes to the search path options _will not_ be
reflected until the dashboard service is restarted

### Implementation of runtime search path updates
At initialisation time, the connection config options are parsed and these are used to determine
the DbClient `requiredSessionSearchPath`.

Whenever a steampipe service is running (either db service or dashboard service), a plugin manager process runs.
This is a GRPC service which has connections to the plugins, and the FDW. 
It is started by the steampipe service startup code.

In the plugin manager process, a connection config file-watcher runs. If the connection config or options have changed,
`RefreshConnectionsAndSearchPaths` is called. As discussed above this has the affect of:
- setting the user search path on the DB (this search path will be used for any subsequent connecitons from external clients)
- setting the `requiredSessionSearchPath` on the (local) DbClient. HOWEVER - this just sets the required search path on the DbClient in the plugin manager process, NOT any DbClient used by Steampipe Query or Dashboard processes.

####Dashboard service search path implementation
When the dashboard server is started, it creates a DbClient, whose `requiredSessionSearchPath` is populated _at init time_, based on the current options amnd config values.
If the options are changed while the service is running, the `requiredSessionSearchPath` for the Dashboard server DbClient _is not updated_