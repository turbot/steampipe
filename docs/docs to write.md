# CLI DOCS



## Command Initialisation
### Query
#### Batch
#### Interactive

1) Create InitData
```
type InitData struct {
	Loaded    chan struct{}         // a channel which is closed when the initialisation is complete
	Queries   []string              // a list of queries specifed in the command args 
	Workspace *workspace.Workspace  // the workspace
	Client    db_common.Client      // the database client
	Result    *db_common.InitResult // the initialisation result 
	cancel    context.CancelFunc    // a cancel function used to cancel during initialisation 
}

initData := query.NewInitData(ctx, w, args)
```

During initialisation we 
### Check
### Dashboard

## Refresh connections

...

Finally, call `LoadForeignSchemaNames` which updates the client `foreignSchemas` property with a list of foreign schema 

###Setting Search Path
`LocalDbClient.RefreshConnectionAndSearchPaths` simplified, does this:
```
refreshConnections()
setUserSearchPath()
SetSessionSearchPath()
```
####setUserSearchPath
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
(otherwise fall back to the user search path set in  setUserSearchPath`)    


### Plugin Manager

### Connection watching
The plugin manager starts a file watch which detects changes in the connection config

Whenever a connection config change is detected:
- The first change event is ignored - we always receive stray event on startup
- The connection config is loaded
- The updated connection config is sent to the plugin manager
- *RefreshConnections* is called - this will ensure the database schema and search paths are updated to match the connection config

NOTE: if a connection is added while a query session is running:
- The new schema will be available for use
- The search path will NOT be updated, as this is set per database session.
For the new search path to apply, a NEW session must be started 



## Session data
### Introspection tables
### Prepared statements

## Control Hooks

When executing controls, a struct implementing `ControlHooks` interface is injected into the context.

```
type ControlHooks interface {
	OnStart(context.Context, *ControlProgress)
	OnControlStart(context.Context, ControlRunStatusProvider, *ControlProgress)
	OnControlComplete(context.Context, ControlRunStatusProvider, *ControlProgress)
	OnControlError(context.Context, ControlRunStatusProvider, *ControlProgress)
	OnComplete(context.Context, *ControlProgress)
}
```

### Check Implementation
When executing `steampipe check`, an instance of `StatusControlHooks` is used for the `ControlHooks`. 
This implementation displays the status of the current control run.

### Dashboard Implementation
When executing `steampipe dashboard`, for each `CheckRun` an instance of `DashboardEventControlHooks` is used for the `ControlHooks`.

This implementation raises dashboard events when the check run completes or experiences an error


NOTE: this is injected into the context in CheckRun.Execute, i.e. each dashboard in a check run will have iuts own implementation

### Snapshot Implementation

## Service management


## Option naming standards


## Plugin query result caching

IndexBuckets are stored keyed by table name and connection 

IndexBuckets contain an array of IndexItems: 
```
// IndexBucket contains index items for all cache results for a given table and connection
type IndexBucket struct {
	Items []*IndexItem
}
```

Each index item has the key of a cache result, and the columns, quals and insertion time of that item.

```
// IndexItem stores the columns and cached index for a single cached query result
// note - this index item it tied to a specific table and set of quals
type IndexItem struct {
	Columns       []string
	Key           string
	Limit         int64
	Quals         map[string]*proto.Quals
	InsertionTime time.Time
}
```

### Cache Get
- Build index bucket key from connection name and table
- Get the index bucket from cache 
- If the index bucket exists, determine whether it contains an index item which satisfies the quals, columns, limit and ttl or the request.
  (NOTE: only key column quals are used when checking cached data)
- If a matching index item is found, use the `Key` property to retrieve the result 

#### Identifying cache hits
- Columns
- Limit
- Qual subset
- Qual exact match
 




## Plugin Instantiation

```
hub.getConnectionPlugin(connection name)
    get plugin name from connection config for connection 
    <other stuff>
    
    hub.createConnectionPlugin(plugin name, connection name)
        CreateConnectionPlugins(connections []*modconfig.Connection)
            pluginManager.Get(&proto.GetRequest{})
            
            
also during refresh connections

populateConnectionPlugins        
    CreateConnectionPlugins(connectionsToCreate    
```


## Config Initialisation and Precedence

Connection config consists of: 
- Steampipe connections (including options overrides)
- Steampipe default options (from the config directory)
- Workspace specific options (from the mod location)



#### Load Workspace Profile

#### Set Install dir
If not set default to wd

### Load connection config

Uses mod location to load config files in mod directory




## Why does FDW Need connection config?

1. FDW needs to know if a connection is an aggregator and if so, it needs to resolve the child connection names
It does this to determine whether to push down limit and build the limit map
2. FDW needs the connection options to get cache parameters 

Possible solution is for plugin manager to have an endpoint which returns the necessary connection information


## Interface Usage
missing from QueryProviderBase
Name
GetTitle
GetUnqualifiedName

**QueryProvider**
```
HclResource
GetArgs() *QueryArgs
GetParams() []*ParamDef
GetSQL() *string
GetQuery() *Query
SetArgs(*QueryArgs)
SetParams([]*ParamDef)
GetMod() *Mod
GetDescription() string

GetPreparedStatementExecuteSQL(*QueryArgs) (*ResolvedQuery, error)
// implemented by QueryProviderBase
AddRuntimeDependencies([]*RuntimeDependency)
GetRuntimeDependencies() map[string]*RuntimeDependency
RequiresExecution(QueryProvider) bool
VerifyQuery(QueryProvider) error
MergeParentArgs(QueryProvider, QueryProvider) (diags hcl.Diagnostics)
```

- Control
- DashboardCard
- DashboardChart
- DashboardEdge
- DashboardFlow
- DashboardGraph
- DashboardHiearchy
- DashboardImage
- DashboardInput
- DashboardNode
- DashboardTable
- Query

**HclResource**

```
    Name() string
	GetTitle() string
	GetUnqualifiedName() string
	CtyValue() (cty.Value, error)
	OnDecoded(*hcl.Block, ResourceMapsProvider) hcl.Diagnostics
	GetDeclRange() *hcl.Range
	BlockType() string
	GetDescription() string
	GetTags() map[string]string
```
- DirectChildrenModDecorator
- Benchmark  x
- Control x
- Dashboard x
- DashboardCard x
- DashboardCategory x
- DashboardChart x
- DashboardEdge x
- DashboardFlow x
- DashboardGraph x
- DashboardHiearchy x
- DashboardImage x
- DashboardInput x
- DashboardNode x
- DashboardTable  x
- DashboardText
- Local x
- Mod     x
- Query   x
- variable x
- DashboardContainer

**ModTreeItem**
```
	AddParent(ModTreeItem) error
	GetParents() []ModTreeItem
	GetChildren() []ModTreeItem
	// TODO move to Hcl Resource
	GetDocumentation() string
	// GetPaths returns an array resource paths
	GetPaths() []NodePath
	SetPaths()
	GetMod() *Mod
	```

- Benchmark x  
- Control x
- Dashboard  x
- DashboardCard 
- DashboardCategory 
- DashboardChart 
- DashboardEdge 
- DashboardFlow
- DashboardGraph 
- DashboardHiearchy 
- DashboardImage 
- DashboardInput 
- DashboardNode 
- DashboardTable  
- DashboardText
- Local x
- Mod     x
- Query   x
- variable x

Meed to parse sepatarely into HclResourceBase and ModTreeItemBase

```

# SDK DOCS
## Connection source file watching


### Limit handling
#### FDW Limit Parsing
#### SDK Limit behaviour




