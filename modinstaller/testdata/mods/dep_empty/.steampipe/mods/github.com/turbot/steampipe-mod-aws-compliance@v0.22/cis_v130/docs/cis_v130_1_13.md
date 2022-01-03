## Description

Access keys are long-term credentials for an IAM user or the AWS account root user. You can use access keys to sign programmatic requests to the AWS CLI or AWS API (directly or using the AWS SDK).

One of the best ways to protect your account is to not allow users to have multiple access keys as this is being used for programmatic requests.

## Remediation

### From Console:

Perform the following action to deactivate access keys:

1. Sign into the AWS console as an **Administrator** and navigate to the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, choose Users.
3. Click on the **User name** for which more than one active access key exists.
4. Click on **Security credentials** tab.
5. Click on the **Make inactive** to `deactivate` the non-operational key.

**Note**: Test your application to make sure that the active access key is working.

### From Command Line:

Run the `update-access-key` command below using the IAM user name and the non-operational access key IDs to deactivate the unnecessary key.

```bash
aws iam update-access-key --access-key-id <access-key-id> --status Inactive - -user-name <user-name>
```