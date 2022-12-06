# Service Management and Lifecycle in Steampipe

* Two types of service invocation
  * `implicit`
  * `explicit`

## `explicit` service
* explicit service is started by `steampipe service start` and can only be shutdown by `steampipe service stop`
* `steampipe service status` reports connection parameters
* `steampipe service start` always binds to `network` by default
* all `steampipe` commands will bind to the service started by `service start`
* `service stop` by default does not shutdown the service if there are clients connected to it. We need to disconnect all clients or use `--force`

## `implicit` service
* all commands will bind to a running service if available - and if not available, will start the service and bind to it
* implicit services are shutdown if no other `steampipe` clients are connected to it. `steampipe` will shutdown the service even if other non-steampipe clients are connected. This differs from `explicit` shutdown behavior
* `steampipe service status` will not report connection parameters, but will report that `service is running implicitly`
* `implicit` service always binds to `localhost`

## Scenarios

### `Sc1`
1. No `service` is running
1. `steampipe check` is started - start `service` in `implicit` mode
1. `steampipe query` is started - binds to `service` started by `check`
1. `steampipe check` finishes
    1. tries to shutdown service
    1. sees another `steampipe` connected to `service`
    1. exits without shutting down `service`
1. `steampipe query` finished.
1. shuts down service before exiting

## `Sc2`
1. No `service` is running
1. `steampipe check` is started - start `service` in `implicit` mode
1. `pgcli` is started with the connection string: `postgres://steampipe@localhost:9193`
    > Since on `localhost`, `steampipe` doesn't require authentication and accepts the connection
1. `steampipe check` finishes
    1. tries to shutdown service
    1. does not see any another `steampipe` connected to `service`
    1. forcefully shuts down service even if another client is connected
    
## `Sc3`
1. `steampipe service start` is executed
1. `steampipe check` is started - `service` is already running - binds to that service
1. `steampipe query` is started - `service` is already running - binds to that service
1. `steampipe check` finished - `service` is in explicit mode - no shutdown
1. `steampipe query` finished - `service` is in explicit mode - no shutdown
1. `steampipe service stop` stops the `service`