## Description

This control checks whether Classic Load Balancers have connection draining enabled.

Enabling connection draining on Classic Load Balancers ensures that the load balancer stops sending requests to instances that are de-registering or unhealthy. It keeps the existing connections open. This is particularly useful for instances in Auto Scaling groups, to ensure that connections aren’t severed abruptly.

## Remediation
To enable connection draining on Classic Load Balancers, following the steps in [Configure connection draining for your Classic Load Balancer](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-conn-drain.html) in User Guide for Classic Load Balancers.