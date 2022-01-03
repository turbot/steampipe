## Description

CloudTrail is a web service that records AWS API calls for an account and makes those logs available to users and resources in accordance with IAM policies. AWS Key Management Service (AWS KMS) is a managed service that helps create and control the encryption keys used to encrypt account data, and uses hardware security modules (HSMs) to protect the security of encryption keys.

You can configure CloudTrail logs to leverage server-side encryption (SSE) and AWS KMS customer-created master keys (CMKs) to further protect CloudTrail logs.

Configuring CloudTrail to use SSE-KMS provides additional confidentiality controls on log data because a given user must have S3 read permission on the corresponding log bucket and must be granted decrypt permission by the CMK policy.

## Remediation

Perform the following to configure CloudTrail to use SSE-KMS:

### From Console

1. Open the CloudTrail console at [CloudTrail](https://console.aws.amazon.com/cloudtrail)
2. Choose Trails, select the trail to update, by clicking **Edit** button in `General details`.
3. In the `Storage location`,
    - For `Log file SSE-KMS encryption`, choose `Enabled`.
4. In `Customer managed AWS KMS key`, select an existing CMK from the KMS key Id drop-down menu
    - Note: Ensure the CMK is located in the same region as the S3 bucket
    - Note: You will need to apply a KMS Key policy on the selected CMK in order for CloudTrail as a service to encrypt and decrypt log files using the CMK provided.
    - To create a key, enter an alias for the key in the `KMS alias` field. The key is created in the same Region as the bucket.
5. Click Save

### From Command Line

```bash
aws cloudtrail update-trail --name <trail_name> --kms-id <cloudtrail_kms_key> aws kms put-key-policy --key-id <cloudtrail_kms_key> --policy <cloudtrail_kms_key_policy>
```
