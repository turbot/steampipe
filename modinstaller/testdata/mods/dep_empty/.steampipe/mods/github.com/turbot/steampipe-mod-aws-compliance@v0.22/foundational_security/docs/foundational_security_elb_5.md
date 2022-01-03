## Description

This control checks whether the Application Load Balancer and the Classic Load Balancer have logging enabled. The control fails if access_logs.s3.enabled is false.

Elastic Load Balancing provides access logs that capture detailed information about requests sent to your load balancer. Each log contains information such as the time the request was received, the client's IP address, latencies, request paths, and server responses. You can use these access logs to analyze traffic patterns and to troubleshoot issues.

## Remediation

To remediate this issue, update your load balancers to enable logging.

**To enable access logs**

1. Open the [Amazon EC2 console](https://console.aws.amazon.com/ec2/).
2. In the navigation pane, choose `Load balancers`.
3. Choose an Application Load Balancer.
4. From `Actions`, choose `Edit attributes`.
5. Under `Access logs`, choose `Enable`.
6. Enter your S3 location. This location can exist or it can be created for you. If you do not specify a prefix, the access logs are stored in the root of the S3 bucket.
7. Choose `Save`.