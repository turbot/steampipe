## Description

This control evaluates AWS Application Load Balancers (ALB) to ensure they are configured to drop invalid HTTP headers. The control fails if the value of `routing.http.drop_invalid_header_fields.enabled` is set to `false`.

By default, ALBs are not configured to drop invalid HTTP header values. Removing these header values prevents HTTP desync attacks.

## Remediation

To remediate this issue, configure your load balancer to drop invalid header fields.

**To configure the load balancer to drop invalid header fields**

1. Open the [Amazon EC2 console](https://console.aws.amazon.com/ec2/).
2. In the navigation pane, choose `Load balancers`.
3. Choose an `Application Load Balancer`.
4. From `Actions`, choose `Edit attributes`.
5. Under `Drop Invalid Header Fields`, choose `Enable`.
6. Choose `Save`.