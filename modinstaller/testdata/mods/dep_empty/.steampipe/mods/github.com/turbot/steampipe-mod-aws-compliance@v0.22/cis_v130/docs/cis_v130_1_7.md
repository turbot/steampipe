## Description

With the creation of an AWS account, a root user account is created. This root user is the most privileged user in an AWS account and has unrestricted access to and control over all resources in the account. It is highly recommended that the use of this root user to be avoided for everyday tasks.

By default IAM *root user* account for us-gov cloud regions is not enabled. However, on request AWS support can enable *root user* access keys only through CLI or API methods.

As the root user has unrestricted access to all the resources, it is dangerous to use for daily task. To avoid this it better to deactivate or delete any access keys associated with it. Also to change the root user password as necessary. Use of it, is inconsistent with the principles of least privilege and separation of duties, and can lead to unnecessary harm due to mistakes.

## Remediation

When you find that the root user account is being used for daily activity that includes administrative tasks that do not require the root user, perform the following action:

1. Change the root user password.
2. Deactivate or delete any access keys associated with the root user.