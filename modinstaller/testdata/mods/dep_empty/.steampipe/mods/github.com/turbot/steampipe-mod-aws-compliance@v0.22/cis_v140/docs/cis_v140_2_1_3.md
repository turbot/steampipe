## Description

Once MFA Delete is enable on your sensitive and classified S3 bucket it requires the user to have two forms of authentication.

Adding MFA delete to an S3 bucket, requires additional authentication when you change the version state of your bucket or you delete and object version adding another layer of security in the event your security credentials are compromised or unauthorized access is granted.

## Remediation

### From Command Line

Perform the steps below to enable MFA delete on an S3 bucket.
**Note:** You cannot enable MFA Delete using the AWS Management Console. You must use the AWS CLI or API. You must use your root account to enable MFA Delete on S3 buckets

1. Run the s3ap put-bucket-versioning command

```bash
aws s3api put-bucket-versioning --profile my-root-profile \
--bucket Bucket_Name --versioning-configuration Status=Enabled,MFADelete=Enabled --mfa \
“arn:aws:iam::aws_account_id:mfa/root-account-mfa-device passcode”
```
