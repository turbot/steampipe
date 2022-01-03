## Description

AWS IAM users can access AWS resources using different types of credentials, such as passwords or access keys. It is recommended that all credentials that have been unused in 45 or greater days to be deactivated or removed.

Disabling or removing unnecessary credentials will reduce the window of opportunity for credentials associated with a compromised or abandoned users to be used.

## Remediation

### From Console:

Perform the following action to disable user console password:

1. Sign into the AWS console and navigate to the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, choose Users.
3. Select the **User name** whose `Console last sign-in` is greater than 45 days.
4. Click on **Security credentials** tab.
5. In section `Sign-in credentials`, `Console password` click **Manage**.
6. Select `Disable`, click **Apply**

Perform the following action to deactivate access keys:

1. Sign into the AWS console as an **Administrator** and navigate to the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, choose Users.
3. Click on the **User name** for which access key is over 45 days old.
4. Click on **Security credentials** tab.
5. Click on the **Make inactive** to `deactivate` the key that is over 45 days old and that have not been used.
