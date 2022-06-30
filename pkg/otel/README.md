# OpenTelemetry Collector 

This collector is provided for local testing purposes. It uses `docker-compose` and by default runs against the 
`otel/opentelemetry-collector-contrib-dev:latest` image. 

To run the collector, switch to the `otel` folder and run:

```shell
docker-compose up -d
```

The demo exposes the following backends:

- Jaeger at http://0.0.0.0:16686
- Prometheus at http://0.0.0.0:9090 

Notes:

- It may take some time for the application metrics to appear on the Prometheus
 dashboard;

To clean up any docker container from the demo run `docker-compose down` from 
the `examples/demo` folder.



