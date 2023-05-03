# Connection State 

## Overview
Connection state is stored in multiple locations

- The connection config (spc files) - the golden source
- The connections.json state file - updated AFTER a successful refresh connections. (Would be nice to remove this if possible)
- the connection state table - updated by refresh connections
- the actual db foreign schemas
- the interactive client inspect data 

## Loading and saving state

**RefreshConnections**

Current behaviour:
- Load foreign Schema names
- Create ConnectionUpdates
  - Build requiredConnectionState (connection config)
  - Load current connection state (the connections.json file)
  - Execute update/deletion queries
  - On success, write back the connections.json file

Currently this loads the connections.json file
However instead it should load the connections table, joined with the foreign schema list, and identify 'ready' connections.
This is more up to date (the state file is onl written at the end)


## Connection state table


| Column                                  	 | Type                        	 | Description               	                    |
|-------------------------------------------|------------------------------|------------------------------------------------|
| connection_name                         	 | string                      	 | connection name           	                    |
| status 	                                  | string 	                     | pending / updating / deleting / ready / error  |
| destails   	                              | string                       | populated if state is `error`       	          |
| comments_set   	                          | bool             	         | have the comments been set for this connection |
| time_changed   	                          | timestamptz             	 | last change time	                              |


```sql
CREATE TABLE IF NOT EXISTS connection_state (
   connection_name text
   status text
   error text
   comments_set bool   
);
```

## Service startup

- create table if does not exist
- if it does exist, set all rows status to pending
```sql
UPDATE connection_state SET status = 'pending'
```

## Refresh Connections

- After building ConnectionUpdates, set status of connection to [updating / deleting / ready / error] as appropriate
- After updating every N connections, set their state to  [ready / error] as appropriate
- After deletions, delete removed connecitrons from table
- After comments are inserted

# Command execution (Query/Control/Dashboard)
## Interactive Query
- After client acquisistion:
  - start notification listener
  - read connection state table. If any connections are either `pending` or `updating`, wait for notifications to indicate the update is complete


# ConnectionData
```go
type ConnectionData struct {
	StructVersion int64 `json:"struct_version,omitempty"`
	// the fully qualified name of the plugin
	Plugin string `json:"plugin,omitempty"`
	// the underlying connection object
	Connection *modconfig.Connection `json:"connection,omitempty"`
	// schema mode - static or dynamic
	SchemaMode string `json:"schema_mode,omitempty"`
	// the hash of the connection schema
	SchemaHash string `json:"schema_hash,omitempty"`
	// the creation time of the plugin file (only used for local plugins)
	ModTime time.Time `json:"mod_time"`
	// loaded is false if the plugin failed to load
	Loaded bool `json:"loaded"`
	// error to be populated if we failed to start/load plugin
	Error string `json:"error,omitempty"`
}
```

Currently stored in connections.json, and updated in refreshConnections


## ISSUES
# connection error
  - save to connection state
  - send notification

Execution code
  - when executing query, if receive "relation not found" error
    - if schema is specified
      - if connection does not exist in state map, bubble error
      - if connection is in error, bubble error
      - if connection is ready ( and has been for > backoff interval) assume an actual missing table - bubble error
      - if connection is loading, wait/retry
    - if schema is NOT specified
      - if all connections are ready, bubble error
      - otherwise wait/retry
      - TODO if first plugin connection in search path is in error, bubble error (?????)
    
  - before staring query/control/dashboard execution
    - if custom search path, wait until first schema for each plugin is loaded (all schemas for dynamic)

  - receive error notification: 
    - if static schema and first plugin connection in search path, bubble error
    - if dynamic schema and failed connection in active search path, fail

- what if a connections change midways through control/dashboard run? (client detects and warns?)
- what do we do if there is a file watch event before previous refresh is complete