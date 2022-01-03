## Description

This control checks whether account-level encryption is enabled by default for Amazon Elastic Block Store(Amazon EBS). The control fails if the account level encryption is not enabled.

When encryption is enabled for your account, Amazon EBS volumes and snapshot copies are encrypted at rest. This adds an additional layer of protection for your data. For more information, see Encryption by default in the Amazon EC2 User Guide for Linux Instances.

Note that following instance types do not support encryption: R1, C1, and M1.

## Remediation

You can use the Amazon EC2 console to enable default encryption for Amazon EBS volumes.

**To configure the default encryption for Amazon EBS encryption for a Region**

1. Open the [Amazon EC2 console at](https://console.aws.amazon.com/ec2/).
2. From the navigation pane, select `EC2 Dashboard`.
3. In the upper-right corner of the page, choose `Account Attributes`, `EBS encryption`.
4. Choose `Manage`.
5. Select `Enable`. You can keep the AWS managed CMK with the alias alias/aws/ebs created on your behalf as the default encryption key, or choose a symmetric customer managed CMK.
6. Choose `Update EBS encryption`.