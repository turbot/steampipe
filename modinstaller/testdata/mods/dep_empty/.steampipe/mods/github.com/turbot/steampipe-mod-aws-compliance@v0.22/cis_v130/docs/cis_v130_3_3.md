## Description

CloudTrail logs a record of every API call made in your AWS account. These logs file are
stored in an S3 bucket. It is recommended that the bucket policy or access control list (ACL)
applied to the S3 bucket that CloudTrail logs to prevent public access to the CloudTrail logs.

Allowing public access to CloudTrail log content might aid an adversary in identifying weaknesses in the affected account's use or configuration.

## Remediation

Perform the following to remove any public access that has been granted to the bucket via an ACL or S3 bucket policy:

### From Console

Using **Block public access** settings.

1. Go to Amazon S3 console at [S3](https://console.aws.amazon.com/s3/home)
2. Choose the name of the bucket where your CloudTrail are stored.
3. Choose **Permissions** and then choose **Block public access** settings.
4. Choose Edit, select all four options under `Block all public access` check box, and then choose save changes.
5. If prompted, enter `confirm` and then choose **Confirm**.

Using **ACL** and **Bucket Policy** settings.

1. Go to Amazon S3 console at [S3](https://console.aws.amazon.com/s3/home)
2. Choose the name of the bucket where your CloudTrail are stored.
3. Choose **Permissions** and navigate to `Access control list (ACL)`
4. The tab shows a list of grants, one row per grant, in the bucket ACL. Each row
identifies the grantee and the permissions granted.
5. Ensure no rows exists that have the Grantee set to Everyone or the Grantee set to
Any Authenticated User.
6. If the Edit bucket policy button is present, click it to review the bucket policy.
7. Ensure the policy does not contain a Statement having an Effect set to Allow and a
Principal set to `"*"` or `{"AWS" : "*"}`
