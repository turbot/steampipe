## Description

The root user account is the most privileged user in an AWS account. AWS Access Keys provide programmatic access to a given AWS account. It is recommended that all access keys associated with the root user account be removed.

By default IAM *root user* account for us-gov cloud regions is not enabled. However, on request AWS support can enable *root user* access keys only through CLI or API methods.

Removing access keys associated with the *root user* account limits vectors by which the account can be compromised. Additionally, removing the root access keys encourages the creation and use of role based accounts that are least privileged.

## Remediation

### From Console

Perform the following action to delete or disable active root user access keys:

1. Sign into the AWS console as a root user, and navigate to [Your Security Credentials](https://console.aws.amazon.com/iam/home#/security_credentials).
2. Click on Access Keys (access key ID and secret access key) section.
3. Under the Status column if there are any Keys which are Active
    - Click on **Make Inactive** to make it inactive.
    - Click **Delete** to delete it permanently (deleted keys cannot be recovered).