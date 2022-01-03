## Description

This control checks whether AWS CloudTrail is configured to use the server-side encryption (SSE) AWS KMS customer master key (CMK) encryption.

If you are only using the default encryption option, you can choose to disable this check.

## Remediation

To enable encryption for CloudTrail logs

1. Open the CloudTrail console at [CloudTrail](https://console.aws.amazon.com/cloudtrail/).
1. Choose **Trails**.
1. Choose the trail to update.
1. Under General details, choose **Edit**.
1. For Log file SSE-KMS encryption, select **Enabled**.
1. Under AWS KMS customer managed CMK, do one of the following:
    - To create a key, choose **New**. Then in AWS KMS alias, enter an alias for the key. The key is created in the same Region as the S3 bucket.
    - To use an existing key, choose **Existing** and then from AWS KMS alias, select the key.
    - The AWS KMS key and S3 bucket must be in the same Region.
1. Choose **Save changes**.