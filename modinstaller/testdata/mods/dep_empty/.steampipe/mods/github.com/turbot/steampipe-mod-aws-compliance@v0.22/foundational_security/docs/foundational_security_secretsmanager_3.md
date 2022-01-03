## Description

This control checks whether your secrets have been accessed within a specified number of days. The default value is 90 days. If a secret was accessed even once within the defined number of days, this control fails.

Deleting unused secrets is as important as rotating secrets. Unused secrets can be abused by their former users, who no longer need access to these secrets. Also, as more users get access to a secret, someone might have mishandled and leaked it to an unauthorized entity, which increases the risk of abuse. Deleting unused secrets helps revoke secret access from users who no longer need it. It also helps to reduce the cost of using Secrets Manager. Therefore, it is essential to routinely delete unused secrets.

## Remediation

You can delete inactive secrets from the Secrets Manager console.

**To delete inactive secrets**

1. Open the [Secrets Manager console](https://console.aws.amazon.com/secretsmanager/).
3. To locate the secret, enter the secret name in the search box.
3. Choose the secret to delete.
4. Under `Secret details`, from `Actions`, choose `Delete secret`.
5. Under `Schedule secret deletion`, enter the number of days to wait before the secret is deleted.
6. Choose `Schedule deletion`.