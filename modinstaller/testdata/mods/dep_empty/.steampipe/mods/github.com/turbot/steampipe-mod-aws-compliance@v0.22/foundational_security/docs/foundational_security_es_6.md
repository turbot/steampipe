## Description

This control checks whether Elasticsearch domains are configured with at least three data nodes and zoneAwarenessEnabled is true.

An Elasticsearch domain requires at least three data nodes for high availability and fault-tolerance. Deploying an Elasticsearch domain with at least three data nodes ensures cluster operations if a node fails.

## Remediation

To modify the number of data nodes in an Amazon ES domain

1. Open the [Amazon Elasticsearch console](https://console.aws.amazon.com/es/).
2. Under `My domains`, choose the name of the domain to edit.
3. Choose `Edit domain`.
4. Under `Data nodes`, set `Number of nodes` to a number greater than or equal to three.
   For three Availability Zone deployments, set to a multiple of three to ensure equal distribution across Availability Zones.
5. Under Dedicated master nodes, set Instance type to the desired instance type.
6. Choose `Submit`.