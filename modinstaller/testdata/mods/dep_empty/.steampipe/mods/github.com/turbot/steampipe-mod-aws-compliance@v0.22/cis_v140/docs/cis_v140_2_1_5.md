## Description

Amazon S3 provides *Block public access (bucket settings)* and *Block public access (account settings)* to help you manage public access to Amazon S3 resources. By default, S3 buckets and objects are created with public access disabled. However with an IAM principle with sufficient S3 permissions can enable public access at the bucket and/or object level.

While enabled, Block public access (bucket settings) prevents an individual bucket and its objects, from becoming publicly accessible.
Similarly, Block public access (account settings) prevents all buckets and it's objects in an account, from becoming publicly accessible.

Amazon S3 *Block public access (bucket settings)* prevents the accidental or malicious public exposure of data contained within the respective bucket(s).

Amazon S3 *Block public access (account settings)* prevents the accidental or malicious public exposure of data contained within all buckets of the respective AWS account.

Whether blocking public access to all or some buckets is an organizational decision that should be based on data sensitivity, least privilege, and use case.

When you apply Block Public Access settings to an account, the settings apply to all AWS Regions globally. The settings might not take effect in all Regions immediately or simultaneously, but they eventually propagate to all Regions.

## Remediation

### From Console

By using Block Public Access (bucket settings):

1. Login to AWS Management Console and open the [Amazon S3 console](https://console.aws.amazon.com/s3/).
2. Click on the bucket name.
3. Go to **Permissions** tab.
4. Click **Edit** for `Block all public access (bucket setting)`.
5. Ensure that block public access settings are set appropriately for this bucket.
6. Repeat for all the buckets in your AWS account that contain sensitive data.

By using Block Public Access (account settings):

1. Login to AWS Management Console and open the [Amazon S3 console](https://console.aws.amazon.com/s3/).
2. In the left navigation pane, choose **Block Public Access settings for this account**
3. Ensure that block public access settings are set appropriately for your AWS account.

### From Command Line

To set Block Public access settings for the buckets, run the following commands:

1. List all of the S3 Buckets

```bash
aws s3 ls
```

2. Set the public access to true on that bucket

```bash
aws s3api put-public-access-block --bucket <name-of-bucket> --public-access- block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPu blicBuckets=true"
```

To set Block Public access settings for the account, run the following command:

```bash
aws s3control put-public-access-block --public-access-block-configuration BlockPublicAcls=true, IgnorePublicAcls=true, BlockPublicPolicy=true, RestrictPublicBuckets=true --account-id <value>
```
