## Description

AWS console defaults to no check boxes selected when creating a new IAM user. When creating the IAM user access type you have to determine what type of access they require.

**Programmatic access**:The IAM user might need to make API calls, use the AWS CLI, or use the tools for windows powershell. In that case, create an access key (access key ID and a secret access key) for that user.
**AWS Management Console access**: If the user needs to access the AWS Management Console, create a password for the user.

After user profile is created, user can create access keys for programmatic access which will provide an indication that it is needed for their work. User can also put a support ticket to have access keys created for them.

## Remediation

### From Console:

Perform the following action to check if an access key is created during user creation:

1. Sign into the AWS console and navigate to the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, choose Users.
3. Click on the **User name** where column `Password age` and `Access key age` is not set to **None**.
4. Click on **Security credentials** tab.
5. Compare the user `Creation time` to the Access Key `Created` date and time.
6. For any that match, the key was created during initial user setup.

**Note**: Keys that were created at the same time as the user profile and do not have a last used date should be deleted.

Perform the following action to delete access keys:

1. Sign into the AWS console as an **Administrator** and navigate to the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, choose Users.
3. Click on the **User name** for which access key is to be deleted.
4. Click on **Security credentials** tab.
5. Click on the **Make inactive** to `deactivate` the keys that were created at the same time as the user profile but have not been used.
6. Now click X (delete) for the `Inactive` keys.

### From Command Line:

```bash
aws iam delete-access-key --access-key-id <access-key-id-listed> --user-name <users-name>
```
