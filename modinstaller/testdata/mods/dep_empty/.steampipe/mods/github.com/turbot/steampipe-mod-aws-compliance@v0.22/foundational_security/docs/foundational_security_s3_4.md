## Description

This control checks that your Amazon S3 bucket either has Amazon S3 default encryption enabled or that the S3 bucket policy explicitly denies put-object requests without server-side encryption.

When you set default encryption on a bucket, all new objects stored in the bucket are encrypted when they are stored, including clear text PAN data.

Server-side encryption for all of the objects stored in a bucket can also be enforced using a bucket policy.

## Remediation

1. Open the [Amazon S3 console](https://console.aws.amazon.com/s3/).
2. Choose the bucket from the list.
3. Choose **Properties**.
4. Choose **Default encryption**.
5. For the encryption, choose either `AES-256` or `AWS-KMS`.
   1. To use keys that are managed by Amazon S3 for default encryption, choose AES-256. For more information about using Amazon S3 server-side encryption to encrypt your data,
   2. To use keys that are managed by AWS KMS for default encryption, choose AWS-KMS. Then choose a master key from the list of the AWS KMS master keys that you have created. Type the Amazon Resource Name (ARN) of the AWS KMS key to use. You can find the ARN for your AWS KMS key in the IAM console, under Encryption keys. Or, you can choose a key name from the drop-down list.
6. Choose **Save**.