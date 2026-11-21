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

# Acceptance tests (local) - builds this checkout, then runs the files CI runs, in CI's order
tests/acceptance/run-local.sh

# Run only the named files
tests/acceptance/run-local.sh settings.bats
```

`run-local.sh` builds the checkout itself (Go required; it does not use a `steampipe` on your `PATH`) and gives each test file a fresh temporary install directory, warmed up as CI does (`steampipe plugin install chaos chaosdynamic`). This isolates tests from your real `~/.steampipe` installation: the runner itself only stops processes belonging to the installs it created. Some test files (`service`, `force_stop`, `ssl`) call `steampipe service stop --force`, which stops every steampipe service on the machine, so a run that includes them will stop a service of your own. Each file has a 15-minute limit (CI's); a file fails if a test fails, it times out, or it leaves a service running. It prints a per-file summary and exits non-zero if any file failed; per-file logs are kept in the temp directory named at the start of the output.

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
- **PostgreSQL version**: 14.24.0
- **FDW version**: 2.2.9
- **Internal schema**: `steampipe_internal`
- **Install directory**: `~/.steampipe/`
- **Plugin directory**: `~/.steampipe/plugins/`
- **Config directory**: `~/.steampipe/config/`
- **Log directory**: `~/.steampipe/logs/`

### Branching and Workflow

- **Base branch**: `develop` for all work
- **Main branch**: `main` takes only release PRs (see Release Process)
- **Changelog line**: Every PR into `develop` that ships something carries one `Changelog:` line in its body: one or two sentences a person using Steampipe can read. Its `shipyard/changelog-line` status fails until it has one. A PR whose files all match `notShipped` in `.github/release.json` (tests, `.github/`, docs) needs none. The PR template has the line.
- **Bug fixes**: Use the 2-commit pattern (see `.ai/docs/bug-fix-prs.md`)
- **PR titles**: End with `closes #XXXX` for bug fixes
- **Merge-to-develop PRs**: When merging a feature branch into `develop`, the PR title must be `Merge branch '<branchname>' into develop`
- **Small PRs**: One logical change per PR

### AI Development Guides

The `.ai/` directory contains detailed guides for AI-assisted development:
- `.ai/docs/bug-fix-prs.md` - Two-commit bug fix pattern (demonstrate bug, then fix)
- `.ai/docs/bug-workflow.md` - Creating GitHub bug issues
- `.ai/docs/test-generation-guide.md` - Writing effective Go tests
- `.ai/docs/parallel-coordination.md` - Coordinating parallel AI agents
- `.ai/templates/` - PR description templates

## Release Process

Releases go through Turbot's shared release workflows (turbot/shipyard-workflows, called at `v0`). Nobody types a version, pushes a tag, edits `CHANGELOG.md` or opens a Homebrew or steampipe.io PR by hand. The prep job that the start files names the skill that writes the release's prep file.

```
develop ──► start (weekly, or by hand) ──► prep file ──► cut: release/<version> + CHANGELOG section
                                                             │
                        release PR into main ◄───────────────┘  tested build, record, Slack ask in #steampipe
                             │ 1 approval (turbot/pipelings-release-approvers), merge commit
                             ▼
ship.yml: build ──► tag + GitHub Release (archives, .deb, .rpm) + merge-back of main into develop
                        │
                        ├──► Homebrew formula
                        ├──► steampipe.io changelog page
                        └──► Linux distro test
```

- `.github/release.json`: product, `notShipped`, the start, the ask channel and the changelog page.
- `.github/workflows/`: `release-start.yml`, `prep.yml`, `release.yml` (the cut), `release-pr.yml` (results and record), `ask.yml`, `ship.yml`, `announce.yml`, `shipyard-pr.yml` (the Changelog line check; Shipyard's fixed file, the same in every repo, never edited).
- A release that merged but did not ship (its ship job broke) is finished by a re-ship: `release.yml` with `reship` and `fixes`.
- A new FDW version is released in turbot/steampipe-postgres-fdw first, and `FdwVersion` in `pkg/constants/db.go` is bumped by a PR into `develop`.
- The `.claude/skills/release` skill has the steps.
