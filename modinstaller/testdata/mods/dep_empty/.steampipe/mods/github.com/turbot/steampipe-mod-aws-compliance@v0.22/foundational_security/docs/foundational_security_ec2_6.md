## Description

This control checks whether VPC flow logs are found and enabled for VPCs. The traffic type is set to REJECT.

With VPC Flow Logs, you can capture information about the IP address traffic to and from network interfaces in your VPC. After you create a flow log, you can use CloudWatch Logs to view and retrieve the log data.

Security Hub recommends that you enable flow logging for packet rejects for VPCs. Flow logs provide visibility into network traffic that traverses the VPC. They can detect anomalous traffic and provide insight into security workflows.

By default, the record includes values for the different components of the IP address flow, including the source, destination, and protocol. For more information and descriptions of the log fields, see [VPC Flow Logs](https://docs.aws.amazon.com/vpc/latest/userguide/flow-logs.html)in the Amazon VPC User Guide.

## Remediation

To enable VPC flow logging

1. Open the [Amazon VPC console](https://console.aws.amazon.com/vpc/).
2. In the navigation pane, under Virtual Private Cloud, choose Your VPCs.
3. Select a `VPC` to update.
4. At the bottom of the page, choose **Flow Logs**.
5. Choose **Create** flow log.
6. For Filter, choose **Reject**.
7. For Destination log group, choose the `log group` to use.
8. If you chose `CloudWatch Logs` for your destination log group, for IAM role, choose the IAM role to use.
9. Choose **Create**.