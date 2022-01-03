## Description

This control checks whether S3 buckets have bucket-level public access blocks applied. This control fails if any bucket level public access settings are set to false.

Block Public Access at the S3 bucket level provides controls to ensure that objects never have public access. Public access is granted to buckets and objects through access control lists (ACLs), bucket policies, or both.

## Remediation

For information on how to remove public access at a bucket level, see [Blocking public access to your Amazon S3 storage](https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-control-block-public-access.html).