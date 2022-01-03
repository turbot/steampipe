## Description

Amazon S3 provides multiple encryption options to protect data at rest. With default encryption, you can set the behavior for a S3 bucket so that all new objects are encrypted when they are stored in the bucket. The objects can be encrypted using server-side encryption with either Amazon S3-managed keys (SSE-S3) or customer master keys (CMKs) stored in AWS Key Management Service (AWS KMS) (SSE-KMS).

Encrypting data at rest reduces the likelihood that it is unintentionally exposed and can nullify the impact of disclosure if the encryption remains unbroken.

## Remediation

### From Console

1. Open AW S3 console [S3](https://console.aws.amazon.com/s3/).
2. In the buckets list, choose the **Name** of the bucket that you want.
3. Go to **Properties** tab and choose **Edit** under **Default encryption**.
4. Select **Enable** and either select `SSE-S3` or `SSE-KMS`.
5. Click **Save changes**.
6. Repeat for all the buckets in your AWS account lacking encryption.

### From Command Line

Run either
```bash
aws s3api put-bucket-encryption --bucket <bucket name> --server-side-encryption-configuration '{"Rules": [{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm": "AES256"}}]}'
```

or

```bash
aws s3api put-bucket-encryption --bucket <bucket name> --server-side-encryption-configuration '{"Rules": [{"ApplyServerSideEncryptionByDefault": {"SSEAlgorithm": "aws:kms","KMSMasterKeyID": "aws/s3"}}]}'
```
**Note**: The KMSMasterKeyID can be set to the master key of your choosing; aws/s3 is an AWS preconfigured default.
