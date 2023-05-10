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
- _After updating every N connections, set their state to  [ready / error] as appropriate_ ???
- After deletions, delete removed connections from table


**Update execution**
- build search path connections list
- execute these first (in parallel)
- Notify(?)
- then execute remaining updates (in parallel)

**Connection Error**

If there is a connection error for the first pluginm connection in the search path, 
**remove all other connections for plugin** and set their state to "error - first connection in search path ('xxxx') failed to load"  

# Command execution (Query/Control/Dashboard)

When executing query, if receive "relation not found" error:
- if schema is specified
  - if connection does not exist in state map, bubble error
  - if connection is in error, bubble error
  - if connection is ready ( and has been for > backoff interval) assume an actual missing table - bubble error
  - if connection is loading, wait/retry
- if schema is NOT specified
  - if all connections are ready, bubble error
  - otherwise wait for search path connections (if first plugin connection in search path is in error, bubble error)
    
Before staring query/control/dashboard execution:
    - if custom search path, wait "search path schemas" are loaded loaded


NO:
  - receive error notification: 
    - if static schema and first plugin connection in search path, bubble error
    - if dynamic schema and failed connection in active search path, fail




**QUESTIONS**
- what if a connections change midways through control/dashboard run? (client detects and warns?)
- what do we do if there is a file watch event before previous refresh is complete - cancel previous

**ISSUES**
- inspect broken
- autocomplete update
- empty spinner for query
- observed multiple plugin startup timeouts when running benchmark, maybe
caused by the 10 execution threads all trying to start the plugin
- once got transaction deadlocks 