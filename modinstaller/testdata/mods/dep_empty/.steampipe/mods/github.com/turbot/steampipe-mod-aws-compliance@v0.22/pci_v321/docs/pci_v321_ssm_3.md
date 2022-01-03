## Description

This control checks whether the EC2 instances in your account are managed by Systems Manager.

AWS Systems Manager is an AWS service that you can use to view and control your AWS infrastructure. To help you to maintain security and compliance, Systems Manager scans your managed instances. A managed instance is a machine that is configured for use with Systems Manager. Systems Manager then reports or takes corrective action on any policy violations that it detects. Systems Manager also helps you to configure and maintain your managed instances. Additional configuration is needed in Systems Manager for patch deployment to managed EC2 instances.

## Remediation

You can use the Systems Manager quick setup to set up Systems Manager to manage your EC2 instances.

To determine whether your instances can support Systems Manager associations, see [Systems Manager prerequisites](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-prereqs.html) in the AWS Systems Manager User Guide.

1. Open the [AWS Systems Manager console](https://console.aws.amazon.com/systems-manager/).
2. In the navigation pane, choose **Quick setup**.
3. On the configuration screen, keep the default options.
4. Choose **Set up Systems Manager**.
