## Description

Access keys consist of an access key ID and secret access key, which are used to sign programmatic requests that you make to AWS. AWS users need their own access keys to make programmatic calls to AWS from the AWS Command Line Interface (AWS CLI), Tools for Windows PowerShell, the AWS SDKs, or direct HTTP calls using the APIs for individual AWS services. It is recommended that all access keys to be rotated within 90 days.

Rotating access keys will reduce the window of opportunity for an access key that is associated with a compromised or terminated account to be used. Access keys should be rotated to ensure that data cannot be accessed with an old key which might have been lost, cracked, or stolen.

## Remediation

### From Console:

Perform the following action to deactivate access keys:

1. Sign into the AWS console as an **Administrator** and navigate to the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, choose Users.
3. Click on the **User name** for which access key exists that have not been rotated in 90 days.
4. Click on **Security credentials** tab.
5. Click on the **Make inactive** to `deactivate` the key that have not been rotated in 90 days.
6. Click **Create access key** and update programmatic call with new key pair.

**Note**: Test your application to make sure that the new key pair is working.

### From Command Line:

While the first access key is still active, create a second access key, which is active by default. Run the following command:

```bash
aws iam create-access-key
```

At this point, the user has two active access keys.
  - Update all applications and tools to use the new access key pair.
  - Change the state of the first access key to `Inactive` using below command:
 ```bash
 aws iam update-access-key
 ```
