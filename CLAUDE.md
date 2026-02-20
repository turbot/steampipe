# Steampipe

Steampipe is a zero-ETL tool that lets you query cloud APIs using SQL. It embeds PostgreSQL and uses a Foreign Data Wrapper (FDW) to translate SQL queries into API calls via a plugin system.

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│  User: steampipe query "SELECT * FROM aws_s3_bucket WHERE region='us-east-1'"
└──────────────┬───────────────────────────────────────────────────────┘
               │
       ┌───────▼────────┐
       │  Steampipe CLI  │  ← This repo (turbot/steampipe)
       │  (Cobra + Go)   │
       └───────┬─────────┘
               │ Starts/manages
       ┌───────▼──────────────┐
       │  Embedded PostgreSQL  │  (v14, port 9193)
       │  + FDW Extension      │  ← turbot/steampipe-postgres-fdw
       └───────┬──────────────┘
               │ gRPC
       ┌───────▼──────────────┐
       │  Plugin Process       │  Built with turbot/steampipe-plugin-sdk
       │  (e.g. steampipe-    │
       │   plugin-aws)        │
       └───────┬──────────────┘
               │ API calls
       ┌───────▼──────────────┐
       │  Cloud API / Service  │
       └──────────────────────┘
```

### Query Flow

1. User executes SQL (interactive REPL or batch mode)
2. Steampipe CLI ensures PostgreSQL + FDW + plugins are running
3. SQL goes to PostgreSQL, which routes foreign table access to the FDW
4. FDW translates the query (columns, WHERE quals, LIMIT, ORDER BY) into a gRPC `ExecuteRequest`
5. Plugin receives the request, calls the appropriate API, streams rows back via gRPC
6. FDW converts rows to PostgreSQL tuples, returns to the query engine
7. PostgreSQL applies any remaining filters/joins/aggregations and returns results

### Key Design Decisions

- **Process-per-plugin**: Each plugin is a separate OS process, communicating via gRPC (using HashiCorp go-plugin)
- **Qual pushdown**: WHERE clauses are pushed to plugins so they can filter at the API level (e.g. `region = 'us-east-1'` becomes an API parameter)
- **Limit pushdown**: LIMIT is pushed to plugins when sort order can also be pushed
- **Streaming**: Rows are streamed progressively, not buffered
- **Caching**: Two-level caching (query cache in plugin manager, connection cache per-plugin)

## Repository Map

### This Repo: `turbot/steampipe` (CLI)

The Steampipe CLI manages the database lifecycle, plugin installation, and provides the query interface.

```
steampipe/
├── main.go                          # Entry point: system checks, then cmd.Execute()
├── cmd/                             # Cobra commands
│   ├── root.go                      # Root command, global flags
│   ├── query.go                     # `steampipe query` - interactive/batch SQL
│   ├── service.go                   # `steampipe service` - start/stop/status of DB service
│   ├── plugin.go                    # `steampipe plugin` - install/update/list/uninstall
│   ├── plugin_manager.go           # Plugin manager daemon process
│   ├── login.go                     # `steampipe login` - Turbot Pipes auth
│   └── completion.go               # Shell completion
├── pkg/
│   ├── db/
│   │   ├── db_local/               # PostgreSQL process management (start, stop, install, backup)
│   │   ├── db_client/              # Database client (pgx connection pool, query execution, sessions)
│   │   └── db_common/              # Shared DB interfaces and types
│   ├── steampipeconfig/            # HCL config loading (connections, options, connection state)
│   ├── connection/                  # Connection refresh, state tracking, config file watcher
│   ├── pluginmanager_service/      # gRPC plugin manager (starts plugins, manages lifecycle)
│   ├── pluginmanager/              # Plugin manager state persistence
│   ├── interactive/                # Interactive REPL (go-prompt, autocomplete, metaqueries)
│   ├── query/                      # Query execution (init, batch/interactive, history, results)
│   ├── ociinstaller/               # OCI image installer for DB binaries and FDW
│   ├── introspection/              # Internal metadata tables (steampipe_connection, steampipe_plugin, etc.)
│   ├── constants/                  # App constants (ports, schemas, env vars, exit codes)
│   ├── options/                    # Config option types (database, general, plugin)
│   ├── initialisation/             # Startup initialization (DB client, services, cloud metadata)
│   ├── export/                     # Query result export (snapshots)
│   ├── display/                    # Output formatting
│   ├── cmdconfig/                  # CLI flag configuration via viper
│   └── ...                         # error_helpers, statushooks, utils, etc.
├── tests/
│   ├── acceptance/                 # Acceptance test suite
│   ├── dockertesting/             # Docker-based tests
│   └── manual_testing/            # Manual test scripts
└── .ai/                            # AI development guides (see below)
```

#### Key Internal Flows

**Service startup** (`steampipe service start` or implicit on `steampipe query`):
1. `db_local.StartServices()` ensures PostgreSQL is installed (via OCI images)
2. Starts PostgreSQL process with the FDW extension loaded
3. Starts plugin manager, loads plugin processes
4. Refreshes all connections (creates/updates foreign table schemas)
5. Creates internal metadata tables (`steampipe_internal` schema)

**Database client** (`pkg/db/db_client/`):
- Uses `jackc/pgx/v5` connection pool
- Manages per-session search paths (so each query sees the right schemas)
- Executes queries and streams results back

**Interactive mode** (`pkg/interactive/`):
- Uses a fork of `c-bata/go-prompt` for the REPL
- Provides autocomplete for table names, columns, SQL keywords
- Supports metaqueries (`.inspect`, `.tables`, `.help`, etc.)

**Plugin management** (`steampipe plugin install aws`):
- Downloads OCI image from registry → extracts to `~/.steampipe/plugins/`
- On next query, plugin manager starts the plugin process
- FDW imports foreign schema (creates foreign tables for each plugin table)

### Related Repo: `turbot/steampipe-postgres-fdw` (FDW)

The Foreign Data Wrapper is a PostgreSQL extension written in C + Go. It bridges PostgreSQL and plugins.

```
steampipe-postgres-fdw/
├── fdw/                    # C code: PostgreSQL extension callbacks
│   ├── fdw.c              # FDW init, handler registration (FdwRoutine)
│   ├── query.c            # Query planning: column extraction, sort/limit pushdown
│   └── common.h           # Core C structs (ConversionInfo, FdwPlanState, FdwExecState)
├── hub/                    # Go code: query engine that talks to plugins
│   ├── hub_base.go        # Planning (GetRelSize, GetPathKeys) and scan management
│   ├── hub_remote.go      # Remote hub: connection pooling, iterator creation
│   ├── scan_iterator.go   # Row streaming from plugin via gRPC
│   └── connection_factory.go # Plugin connection caching
├── fdw.go                  # Go↔C bridge: exported functions (goFdwBeginForeignScan, etc.)
├── quals.go                # PostgreSQL restrictions → protobuf Quals conversion
├── schema.go               # Plugin schema → CREATE FOREIGN TABLE SQL
├── helpers.go              # C↔Go type conversion (Go values ↔ PostgreSQL Datums)
└── types/                  # Go type definitions (Relation, Options, PathKeys)
```

#### FDW Lifecycle (per query)

| Phase | C Callback | Go Function | What Happens |
|-------|-----------|-------------|--------------|
| Planning | `fdwGetForeignRelSize` | `Hub.GetRelSize()` | Estimate row count and width |
| Planning | `fdwGetForeignPaths` | `Hub.GetPathKeys()` | Generate access paths (for join optimization) |
| Planning | `fdwGetForeignPlan` | - | Choose plan, serialize state |
| Execution | `fdwBeginForeignScan` | `Hub.GetIterator()` | Convert quals, create scan iterator |
| Execution | `fdwIterateForeignScan` | `iterator.Next()` | Fetch rows, convert to Datums |
| Cleanup | `fdwEndForeignScan` | `iterator.Close()` | Cleanup, collect scan metadata |

#### Qual Pushdown

WHERE clauses are converted from PostgreSQL's internal representation to protobuf `Qual` messages:
- `column = value` → `Qual{FieldName, "=", value}`
- `column IN (a, b)` → `Qual{FieldName, "=", ListValue}`
- `column IS NULL` → `NullTest` qual
- `column LIKE '%pattern%'` → `Qual{FieldName, "~~", value}`
- Boolean expressions (AND/OR) are handled recursively
- Volatile functions and self-references are excluded (left for PostgreSQL to filter)

### Related Repo: `turbot/steampipe-plugin-sdk` (Plugin SDK)

The SDK provides the framework for building plugins. Plugin authors only write API-specific code.

```
steampipe-plugin-sdk/
├── plugin/                 # Core plugin framework
│   ├── plugin.go          # Plugin struct, initialization, execution orchestration
│   ├── table.go           # Table definition (columns, List/Get config, hydrate config)
│   ├── column.go          # Column definition (name, type, transform, hydrate func)
│   ├── table_fetch.go     # Fetch orchestration: Get vs List decision, row building
│   ├── query_data.go      # QueryData: quals, key columns, streaming, pagination
│   ├── row_data.go        # Row building: parallel hydrate execution, transform application
│   ├── key_column.go      # Key column definitions (required/optional/any_of, operators)
│   ├── hydrate_config.go  # Hydrate config: dependencies, retry, ignore, concurrency
│   ├── hydrate_error.go   # Error wrapping: retry with backoff, error ignoring
│   └── serve.go           # Plugin startup: gRPC server registration
├── grpc/                   # gRPC server implementation (PluginServer)
│   ├── pluginServer.go    # RPC methods: Execute, GetSchema, SetConnectionConfig, etc.
│   └── proto/             # Protobuf definitions (plugin.proto)
├── query_cache/            # Query result caching
├── rate_limiter/           # Token bucket rate limiting with scoped instances
├── connection/             # Per-connection in-memory caching (Ristretto)
├── transform/              # Data transformation functions (FromField, FromGo, NullIfZero, etc.)
└── row_stream/             # Row streaming channel management
```

#### Plugin Execution Model

When a query hits a plugin table:

1. **Get vs List decision**: If all required key columns have `=` quals → Get call. Otherwise → List call.
2. **List hydrate** runs first, streaming items via `QueryData.StreamListItem()`
3. **Row building** (per item, in parallel):
   - Start all hydrate functions (respecting dependency graph)
   - Hydrates without dependencies run concurrently
   - Each hydrate is wrapped with retry + ignore error logic
   - Rate limiters throttle API calls per scope (connection, region, service)
4. **Transform chain** applied per column: `FromField("Name").Transform(toLower).NullIfZero()`
5. **Row streamed** back to FDW via gRPC

#### Key Types

```
Plugin              → Top-level struct, holds TableMap, config, caches
Table               → Name, Columns, List/Get config, HydrateConfig
Column              → Name, Type, Transform, optional Hydrate function
KeyColumn           → Column name, operators, required/optional/any_of
HydrateFunc         → func(ctx, *QueryData, *HydrateData) (interface{}, error)
QueryData           → Quals, key columns, streaming, connection config
TransformCall       → Chain of FromXXX → Transform → NullIfZero
```

### Related Repo: `turbot/pipe-fittings` (Shared Library)

Shared infrastructure library used by Steampipe, Flowpipe, and Powerpipe.

```
pipe-fittings/
├── modconfig/              # Mod resources: Mod, HclResource, ModTreeItem interfaces
├── connection/             # Connection types (48+ implementations: AWS, Azure, GCP, GitHub, etc.)
│   └── PipelingConnection  # Core interface: Resolve(), Validate(), GetEnv(), CtyValue()
├── parse/                  # HCL parsing engine (decoder, body processing, custom types)
├── constants/              # Shared constants across Turbot products
├── utils/                  # Plugin utilities, string helpers, file ops
├── credential/             # Credential management
├── schema/                 # Resource schema definitions
├── versionmap/             # Dependency version management
├── modinstaller/           # Mod dependency installation
├── ociinstaller/           # OCI image installation
└── backend/                # PostgreSQL connector
```

Steampipe imports pipe-fittings as `github.com/turbot/pipe-fittings/v2`. Key usage:
- `modconfig.SteampipeConnection` for connection configuration types
- `constants` for shared database and cloud constants
- `utils` for common helper functions
- `connection` types for Turbot Pipes integration

## Development Guide

### Building

```bash
go build -o steampipe
```

### Testing

```bash
# Unit tests
go test ./...

# Acceptance tests (local) - sets up a temp install dir, installs chaos plugins, runs all tests
tests/acceptance/run-local.sh

# Run a single acceptance test file
tests/acceptance/run-local.sh 001.query.bats
```

`run-local.sh` creates a temporary `STEAMPIPE_INSTALL_DIR`, runs `steampipe plugin install chaos chaosdynamic`, then delegates to `run.sh`. This isolates tests from your real `~/.steampipe` installation. The `steampipe` binary must already be on your `PATH` (build it first with `go build -o steampipe` and add it or use `go install`).

### Local Development with Related Repos

#### Dependency Chain

```
pipe-fittings          (shared library, no Turbot dependencies)
       ↑
steampipe-plugin-sdk   (depends on nothing Turbot-specific)
       ↑
steampipe-postgres-fdw (depends on pipe-fittings + steampipe-plugin-sdk)
       ↑
steampipe              (depends on pipe-fittings + steampipe-plugin-sdk)
```

Changes flow upward: a change in `pipe-fittings` can affect all three consumers. A change in `steampipe-plugin-sdk` affects `steampipe` and `steampipe-postgres-fdw`. The FDW and CLI are independent of each other.

#### Using `go.mod` Replace Directives

Steampipe's `go.mod` has **commented-out replace directives** that point to sibling directories:

```go
replace (
    github.com/c-bata/go-prompt => github.com/turbot/go-prompt v0.2.6-steampipe.0.0.20221028122246-eb118ec58d50
// github.com/turbot/pipe-fittings/v2 => ../pipe-fittings
//  github.com/turbot/steampipe-plugin-sdk/v5 => ../steampipe-plugin-sdk
)
```

**To develop against a local `pipe-fittings` or `steampipe-plugin-sdk`**, uncomment the relevant line(s). This tells Go to use your local checkout instead of the published module version. This is essential when:

- You need to change `pipe-fittings` or `steampipe-plugin-sdk` alongside `steampipe`
- You're debugging an issue that spans repos (e.g. a config parsing bug in pipe-fittings that manifests in steampipe)
- You want to test unreleased SDK or pipe-fittings changes with the CLI

**Important**: The `go.mod` expects sibling directories (`../pipe-fittings`, `../steampipe-plugin-sdk`). The local workspace should look like:

```
turbot/
├── steampipe/                  # this repo
├── steampipe-postgres-fdw/     # FDW
├── steampipe-plugin-sdk/       # plugin SDK
└── pipe-fittings/              # shared library
```

**Remember to re-comment the replace directives before committing** — they should never be checked in uncommented, as CI and other developers won't have the same local paths. The `go-prompt` replace is permanent (it points to Turbot's fork, not a local path).

The `steampipe-postgres-fdw` repo does **not** have pre-configured replace directives for local development. If you need to develop the FDW against local copies, add them manually:

```go
// in steampipe-postgres-fdw/go.mod
replace (
    github.com/turbot/pipe-fittings/v2 => ../pipe-fittings
    github.com/turbot/steampipe-plugin-sdk/v5 => ../steampipe-plugin-sdk
)
```

#### Cross-Repo Change Workflow

When a change spans multiple repos (e.g. adding a new config field):

1. Make the change in the lowest dependency first (e.g. `pipe-fittings`)
2. Uncomment the replace directive in the consumer repo (`steampipe`)
3. Build and test locally with the replace active
4. Once working, publish the dependency (merge + tag a release)
5. Update `go.mod` in the consumer to reference the new version: `go get github.com/turbot/pipe-fittings/v2@v2.x.x`
6. Re-comment the replace directive
7. Commit and PR the consumer repo

### Key Directories for Common Tasks

| Task | Where to Look |
|------|--------------|
| Fix a CLI command | `cmd/` (command definition) + relevant `pkg/` package |
| Fix query execution | `pkg/query/`, `pkg/db/db_client/` |
| Fix interactive mode | `pkg/interactive/` |
| Fix plugin install/management | `pkg/ociinstaller/`, `pkg/pluginmanager_service/` |
| Fix connection handling | `pkg/steampipeconfig/`, `pkg/connection/` |
| Fix DB startup/shutdown | `pkg/db/db_local/` |
| Fix autocomplete | `pkg/interactive/interactive_client_autocomplete.go` |
| Fix service management | `cmd/service.go`, `pkg/db/db_local/` |
| Change internal tables | `pkg/introspection/` |
| Change config parsing | `pkg/steampipeconfig/load_config.go`, pipe-fittings |
| Fix FDW query planning | `steampipe-postgres-fdw/fdw/` (C) + `hub/` (Go) |
| Fix qual pushdown | `steampipe-postgres-fdw/quals.go` |
| Fix type conversion | `steampipe-postgres-fdw/helpers.go` |
| Fix plugin SDK behavior | `steampipe-plugin-sdk/plugin/` |
| Fix hydrate execution | `steampipe-plugin-sdk/plugin/table_fetch.go`, `row_data.go` |
| Fix caching | `steampipe-plugin-sdk/query_cache/` |
| Fix rate limiting | `steampipe-plugin-sdk/rate_limiter/` |

### Important Constants

- **Default DB port**: 9193 (`pkg/constants/db.go`)
- **PostgreSQL version**: 14.19.0
- **FDW version**: 2.1.4
- **Internal schema**: `steampipe_internal`
- **Install directory**: `~/.steampipe/`
- **Plugin directory**: `~/.steampipe/plugins/`
- **Config directory**: `~/.steampipe/config/`
- **Log directory**: `~/.steampipe/logs/`

### Branching and Workflow

- **Base branch**: `develop` for all work
- **Main branch**: `main` (releases merge here)
- **Release branch**: `v2.3.x` (or similar version branch)
- **Bug fixes**: Use the 2-commit pattern (see `.ai/docs/bug-fix-prs.md`)
- **PR titles**: End with `closes #XXXX` for bug fixes
- **Merge-to-develop PRs**: When merging a release or feature branch into `develop`, the PR title must be `Merge branch '<branchname>' into develop` (e.g. `Merge branch 'v2.3.x' into develop`)
- **Small PRs**: One logical change per PR

### AI Development Guides

The `.ai/` directory contains detailed guides for AI-assisted development:
- `.ai/docs/bug-fix-prs.md` - Two-commit bug fix pattern (demonstrate bug, then fix)
- `.ai/docs/bug-workflow.md` - Creating GitHub bug issues
- `.ai/docs/test-generation-guide.md` - Writing effective Go tests
- `.ai/docs/parallel-coordination.md` - Coordinating parallel AI agents
- `.ai/templates/` - PR description templates

## Release Process

Follow these steps in order to perform a release:

### 1. Changelog
- Draft a changelog entry in `CHANGELOG.md` matching the style of existing entries.
- Use today's date and the next patch version.

### 2. Commit
- Commit message for release changelog changes should be the version number, e.g. `v2.3.5`.

### 3. Release Issue
- Use the `.github/ISSUE_TEMPLATE/release_issue.md` template.
- Title: `Steampipe v<version>`, label: `release`.

### 4. PRs
1. **Against `develop`**: Title should be `Merge branch '<branchname>' into develop`.
2. **Against `main`**: Title should be `Release Steampipe v<version>`.
   - Body format:
     ```
     ## Release Issue
     [Steampipe v<version>](link-to-release-issue)

     ## Checklist
     - [ ] Confirmed that version has been correctly upgraded.
     ```
   - Tag the release issue to the PR (add `release` label).

### 5. steampipe.io Changelog
- Create a changelog PR in the `turbot/steampipe.io` repo.
- Branch off `main`, branch name: `sp-<version without dots>` (e.g. `sp-235`).
- Add a file at `content/changelog/<year>/<YYYYMMDD>-steampipe-cli-v<version-with-dashes>.md`.
- Frontmatter format:
  ```
  ---
  title: Steampipe CLI v<version> - <short summary>
  publishedAt: "<YYYY-MM-DD>T10:00:00"
  permalink: steampipe-cli-v<version-with-dashes>
  tags: cli
  ---
  ```
- Body should match the changelog content from `CHANGELOG.md`.
- PR title: `Steampipe CLI v<version>`, base: `main`.

### 6. Deploy steampipe.io
- After the steampipe.io changelog PR is merged, trigger the `Deploy steampipe.io` workflow in `turbot/steampipe.io` from `main`.

### 7. Close Release Issue
- Check off all items in the release issue checklist as steps are completed.
- Close the release issue once all steps are done.
