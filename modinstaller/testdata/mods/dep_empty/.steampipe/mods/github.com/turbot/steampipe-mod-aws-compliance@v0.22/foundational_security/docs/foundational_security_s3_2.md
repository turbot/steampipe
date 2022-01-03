## Description

This control checks whether your S3 buckets allow public read access by evaluating the Block Public Access settings, the bucket policy, and the bucket access control list (ACL).

Unless you explicitly require everyone on the internet to be able to write to your S3 bucket, you should ensure that your S3 bucket is not publicly writable.

It does not check for read access to the bucket by internal principals, such as IAM roles. You should ensure that access to the bucket is restricted to authorized principals only.

## Remediation

1. Open the [Amazon S3 console](https://console.aws.amazon.com/s3/).
2. Choose the name of the bucket identified in the finding.
3. Choose **Permissions** and then choose **Public access settings.**
4. Choose **Edit**, select all four options, and then choose **Save**.
5. If prompted, enter `confirm` and then choose **Confirm**.