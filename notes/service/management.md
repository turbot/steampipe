# Service Management and Lifecycle in Steampipe

Steampipe uses `postgresql` under-the-hood for processing SQL queries. `PostgreSQL` parses the queries and converts it into `table reads` which it forwards to `Steampipe` for data. In `steampipe` nomenclature, we call this the `service`.

_Almost_ all `steampipe` commands requires that the service be running and it needs to be able to connect to it. If not, `steampipe` will invoke the service first, before continuing with it's task.

## Types of invocation

`steampipe` commands may start the `service` if it is not already running. We call this an **implicit** invocation. In this case, the service will always bind to `localhost`.

Alternatively, the `service` can be made to persist by using the `steampipe service start` command. This is called an **explicit invocation**. Here, the service will bind to all available IP interfaces along with `localhost`.

> Please note that if the `service` is started with `steampipe service start`, then it can only be shutdown with `steampipe service stop`.

## Service status

The state of the service can be queries with the `steampipe service status` command. Depending on the type of invocation, it will print out the status of the service.

### `status` for explicit invocation

If the `service` was invoked with a `steampipe service start` command, `status` returns the following:

```
user@hostname ~ % steampipe service status
Steampipe service is running:

Database:

  Host(s):            localhost, 127.0.0.1, 192.168.10.174
  Port:               9193
  Database:           steampipe
  User:               steampipe
  Password:           ********* [use --show-password to reveal]
  Connection string:  postgres://steampipe@localhost:9193/steampipe

Managing the Steampipe service:

  # Get status of the service
  steampipe service status

  # View database password for connecting from another machine
  steampipe service status --show-password

  # Restart the service
  steampipe service restart

  # Stop the service
  steampipe service stop

user@hostname ~ %
```

### `status` for implicit invocation

If the `service` was invoked automatically, `status` returns the following output:

```
user@hostname ~ % steampipe service status

Steampipe service was started for an active steampipe query session. The service will exit when all active sessions exit.

To keep the service running after the query session completes, use steampipe service start.

user@hostname ~ %

```

## Service discovery

All `steampipe` commands can discover a running service - even if it is an `implicit` invocation.

## Lifecycle

When the service is invoked with `steampipe service start`, the `service` will stay alive and available till `steampipe service stop` is executed or the system is restarted.

Other `steampipe` commands will detect this running service and connect to it.

However, a key difference on the lifecycle of the service when invoked as an `implicit` service.

When the `service` is started `implicitly`, it will stay available only till the last steampipe process is active. Multiple `steampipe` instances can discover a running service (`implicit` and `explicit`) and can connect to it.

However, when the last `steampipe` process exits, it will shutdown an `implicit` service. If other clients are connected to the service, these will be forcefully disconnected.

> When non-steampipe clients need to be connected to the `steampipe` service, the user MUST use `steampipe service start` to start the service. If not started using `steampipe service start`, the clients connected to the service may get forcefully disconnected.

## Scenarios

### Two Steampipe instances with `implicit` service

1. No `service` is running
1. `steampipe check` is started - starts `service` in `implicit` mode
1. `steampipe query` is started - discovers service started by `check` and connects to it
1. `steampipe check` finishes
   1. tries to shutdown service
   1. sees another `steampipe` connected to `service`
   1. exits without shutting down `service`
1. `steampipe query` finished.
1. shuts down service before exiting

## Steampipe and `pgcli` with `implicit` service 

1. No `service` is running
1. `steampipe check` is started - starts `service` in `implicit` mode
1. `pgcli` (or any non-steampipe client) is started with the connection string: `postgres://steampipe@localhost:9193`
   > Since the service is listening on `localhost`, `steampipe` doesn't require authentication and accepts the connection
1. `steampipe check` finishes
   1. tries to shutdown service
   1. does not see any another `steampipe` connected to `service`
   1. forcefully shuts down service even if another client is connected

## Steampipe and third party client with `explicit` service

1. `steampipe service start` is executed
1. `steampipe check` is started - `service` is already running - binds to that service
1. `steampipe query` is started - `service` is already running - binds to that service
1. `steampipe check` finished - `service` is in explicit mode - no shutdown
1. `steampipe query` finished - `service` is in explicit mode - no shutdown
1. `pgcli` (or any non-steampipe client) is started with the connection string: `postgres://steampipe@localhost:9193` - connection is accepted
1. `steampipe service stop` stops the `service`
