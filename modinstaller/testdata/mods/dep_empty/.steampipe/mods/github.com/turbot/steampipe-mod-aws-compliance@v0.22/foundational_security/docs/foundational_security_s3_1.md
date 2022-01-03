## Description

The control passes if all of the public access block settings are set to true.

The control fails if any of the settings are set to false, or if any of the settings are not configured. When the settings do not have a value, the AWS Config rule cannot complete its evaluation.

Amazon S3 public access block is designed to provide controls across an entire AWS account or at the individual S3 bucket level to ensure that objects never have public access. Public access is granted to buckets and objects through access control lists (ACLs), bucket policies, or both.

Unless you intend to have your S3 buckets be publicly accessible, you should configure the account level Amazon S3 Block Public Access feature.

## Remediation

To remediate this issue, enable Amazon S3 Block Public Access.

**To enable Amazon S3 Block Public Access**

1. Open the [Amazon S3 console](https://console.aws.amazon.com/s3/).
2. Choose **Block public access** (account settings).
3. Choose `Edit`.
4. Select `Block all public access`.
5. Choose `Save` changes.