# Introspection tables in the internal schema

## Overview
The internal schema contains the following introspection tables

- `steampipe_connection`
Lists all connections as defined in the connection config. 
- ``
Lists all plugin instances as defined in the connection config.
- `steampipe_plugin_limiter`
Lists all plugin Limiters as defined either in the plugin binary or the plugin connection block


## Lifecycle

### Startup
#### steampipe_connection
- Every time the server is started, the connections are loaded from the table into ConnectionState structs. 
- The table is then deleted and recreated - this is to handle any updates to the table structure
- The connection states are set to either `pending` (if currently `ready`) or `incomplete` (if not).
  (These states will be updated by RefreshConnections.)
- The connections are written back to the table
- RefreshConnections is triggered - this will apply any necessary connection updates and set the states of the connections 
to either `ready` or `error`

#### steampipe_plugin
- Every time the server is started, table is then deleted and recreated - this is to handle any updates to the table structure
- The configured plugin instances are written back to the table


(See `postServiceStart` in pkg/db/db_local/internal.go)

### Connection config file changed

The when a connection file is changed the ConnectionWatcher calls `pluginManager.OnConnectionConfigChanged`, and then calls 
`RefreshConnections` asyncronously

`OnConnectionConfigChanged`calls:
- `handleConnectionConfigChanges`
- `handlePluginInstanceChanges`
- `handleUserLimiterChanges`


`handleConnectionConfigChanges` determines which connections have been added, removed and deleted. It then builds a set of SetConnectionConfigRequest, one for each plugin instance with changed connections

`handlePluginInstanceChanges` determines which plugins have been added, removed and deleted. 
It updates the `steampipe_plugin` table.
###TODO if the plugin for an instance changes, all connections must be dropped and re-added  



`handleUserLimiterChanges` determines which plugin instances have changed limiter definitions. 
It updates the `steampipe_rate_limiter` table and makes a `SetRateLimiters` call to all plugin instances 
with updated rate limiters.  


### TODO: if a plugin instance has no more connections, we should stop it

`RefreshConnections` updates the plugin schemas to correspond with the updated connection config


## steampipe_plugin

### Lifecycle
#### Startup
- Every time the server is started, table is then deleted and recreated - this is to handle any updates to the table structure
- The configured plugin instances are written back to the table

### Plugin config file changed

The when a connection file is changed the ConnectionWatcher calls `pluginManager.OnConnectionConfigChanged`, and then calls
`RefreshConnections` asyncronously

`OnConnectionConfigChanged` determines which connections have been added, removed and deleted.
It then builds a set of SetConnectionConfigRequest, one for each plugin instance with changed connections


`steampipe_plugin` is  





## steampipe_connection

### Usage

`steampipe_connection` table is used to determine whether a connection has been loaded yet.
This is used to allow us to execute queries without wasiting for all connections to load. Instead, we execute the query,
and if it fails with a relation not found error, we poll the coneciton state table until the connection is ready.
Then we retry the query. 
