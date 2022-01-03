## Description

VPC flow logs is a feature that enables you to capture information about the IP traffic going to and from network interfaces in your VPC. After you have created a flow log, you can view and retrieve its data in CloudWatch Logs.

VPC Flow Logs provide visibility into network traffic that traverses the VPC and can be used to detect anomalous traffic or insight during security workflows.

While setting up the VPC flow log, setting filter to `Reject` will dramatically reduce the logging data accumulation for this recommendation and provide sufficient information for the purposes of breach detection, research and remediation. However, during periods of least privilege security
group engineering, setting this the filter to "All" can be very helpful in discovering existing traffic flows required for proper operation of an already running environment.

## Remediation

Perform the following to determine if VPC Flow logs is enabled:

### From Console

1. Open the Amazon VPC console at [VPC](https://console.aws.amazon.com/vpc/)
2. Select required VPC to update from **Your VPCs**
3. Choose the Flow Logs tab in the bottom section of the page.
4. If no Flow Log exists, choose **Create flow log**
5. For Filter, choose `Reject`.
6. For `Destination log group`, select the log group to use.
7. For IAM role, select the IAM role to use.
8. Choose **Create flow log**.
