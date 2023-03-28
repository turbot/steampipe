# Connection State Table


| Column                                  	 | Type                        	 | Description               	                     |
|-------------------------------------------|------------------------------|-------------------------------------------------|
| connection_name                         	 | string                      	 | connection name           	                     |
| status 	                                 | string 	                     | pending / updating / deleting / ready / error   |
| error   	                                 | string                       | populated if state is `error`       	           |
| comments_set   	                         | bool             	           | have the comments been set for this connection	 |

```sql
CREATE TABLE IF NOT EXISTS connection_state (
   connection_name text
   status text
   error text
   comments_set bool   
);
```

## Service startup

- create if does not exist
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
