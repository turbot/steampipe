## Description

This control checks whether the S3 bucket policy prevents principals from other AWS accounts from performing denied actions on resources in the S3 bucket.

Implementing least privilege access is fundamental to reducing security risk and the impact of errors or malicious intent. If an S3 bucket policy allows access from external accounts, it could result in data exfiltration by an insider threat or an attacker.

The `blacklistedactionpatterns` parameter allows for successful evaluation of the rule for S3 buckets. The parameter grants access to external accounts for action patterns that are not included in the `blacklistedactionpatterns` list.

## Remediation

To remediate this issue, edit the S3 bucket policy to remove the permissions.

**To edit an S3 bucket policy**

1. Open the [Amazon S3 console](https://console.aws.amazon.com/s3/).
2. In the `Bucket name` list, choose the name of the S3 bucket for which you want to edit the policy.
3. Choose `Permissions`, and then choose `Bucket Policy`.
4. In the `Bucket policy editor` text box, do one of the following:
   - Remove the statements that grant access to denied actions to other AWS accounts
   - Remove the permitted denied actions from the statements
5. Choose `Save`.