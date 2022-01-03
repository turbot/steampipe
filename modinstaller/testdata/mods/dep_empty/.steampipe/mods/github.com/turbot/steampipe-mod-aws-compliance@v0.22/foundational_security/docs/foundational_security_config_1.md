## Description

AWS Config rule: None. To run this check, Security Hub runs through audit steps prescribed for it in Securing Amazon Web Services. No AWS Config managed rules are created in your AWS environment for this check.

This control checks whether AWS Config is enabled in the account for the local Region and is recording all resources.

It does not check for change detection for all critical system files and content files, as AWS Config supports only a subset of resource types.

The AWS Config service performs configuration management of supported AWS resources in your account and delivers log files to you. The recorded information includes the configuration item (AWS resource), relationships between configuration items, and any configuration changes between resources.

## Remediation

To configure AWS Config settings

1. Open the [AWS Config console](https://console.aws.amazon.com/config/).
2. Choose the Region to configure AWS Config in.
3. If you have not used AWS Config before, choose **Get started**.
4. On the Settings page, do the following:
   1. Under Resource types to record, choose Record all resources supported in this region and Include global resources (e.g., AWS IAM resources).
   2. Under Amazon S3 bucket, either specify the bucket to use or create a bucket and optionally include a prefix.
   3. Under Amazon SNS topic, either select an Amazon SNS topic from your account or create one. For more information about Amazon SNS, see the [Amazon Simple Notification Service Getting Started Guide](https://docs.aws.amazon.com/sns/latest/dg/sns-getting-started.html).
   4. Under AWS Config role, either choose `Create AWS Config service-linked role` or choose `Choose a role from your account` and then choose the role to use.
5. Choose Next.
6. On the **AWS Config** rules page, choose **Skip**.
7. Choose **Confirm**.

For more information about using AWS Config from the AWS CLI, see the [AWS Config Developer Guide](https://docs.aws.amazon.com/config/latest/developerguide/gs-cli-subscribe.html).

You can also use an AWS CloudFormation template to automate this process. For more information, see the [AWS CloudFormation User Guide](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/stacksets-sampletemplates.html).