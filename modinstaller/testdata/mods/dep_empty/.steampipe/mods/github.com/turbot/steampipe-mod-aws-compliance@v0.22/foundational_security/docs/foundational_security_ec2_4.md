## Description

This control checks whether any EC2 instances have been stopped for more than the allowed number of days. An EC2 instance fails this check if it is stopped for longer than the maximum allowed time period, which by default is 30 days.

A failed finding indicates that an EC2 instance has not run for a significant period of time. This creates a security risk because the EC2 instance is not being actively maintained (analyzed, patched, updated). If it is later launched, the lack of proper maintenance could result in unexpected issues in your AWS environment. To safely maintain an EC2 instance over time in a nonrunning state, start it periodically for maintenance and then stop it after maintenance. Ideally this is an automated process.

## Remediation

You can terminate an EC2 instance using either the console or the command line.

Before you terminate the EC2 instance, verify that you won't lose any data:
- Check that your Amazon EBS volumes will not be deleted on termination.
- Copy any data that you need from your EC2 instance store volumes to Amazon EBS or Amazon S3.

**To terminate an EC2 instance (console)**

1. Open the [Amazon EC2 console](https://console.aws.amazon.com/ec2/).
2. In the navigation pane, under `Instances`, choose `Instances`.
3. Select the instance, and then choose `Actions`, `Instance State`, `Terminate`.
4. When prompted for confirmation, choose `Yes`, `Terminate`.

**To terminate an EC2 instance (AWS CLI, Tools for Windows PowerShell)**

Use one of the following commands. For more information about the command line interface, see [Accessing Amazon EC2](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/concepts.html#access-ec2) in the Amazon EC2 User Guide for Linux Instances.

- From the AWS CLI, use [terminate-instances](https://docs.aws.amazon.com/cli/latest/reference/ec2/terminate-instances.html)
- From the Tools for Windows PowerShell, use [Stop-EC2Instance](https://docs.aws.amazon.com/powershell/latest/reference/items/Stop-EC2Instance.html).