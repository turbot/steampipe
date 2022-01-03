## Description

This control checks whether a secret stored in AWS Secrets Manager is configured with automatic rotation.

Secrets Manager helps you improve the security posture of your organization. Secrets include database credentials, passwords, and third-party API keys. You can use Secrets Manager to store secrets centrally, encrypt secrets automatically, control access to secrets, and rotate secrets safely and automatically.

Secrets Manager can rotate secrets. You can use rotation to replace long-term secrets with short-term ones. Rotating your secrets limits how long an unauthorized user can use a compromised secret. For this reason, you should rotate your secrets frequently. To learn more about rotation, see [Rotating your AWS Secrets Manager secrets](https://docs.aws.amazon.com/secretsmanager/latest/userguide/rotating-secrets.html).

## Remediation

To remediate this issue, you enable automatic rotation for your secrets.

**To enable automatic rotation for secrets**

1. Open the [Secrets Manager console](https://console.aws.amazon.com/secretsmanager/).
2. To find the secret that requires rotating, enter the secret name in the search field.
3. Choose the secret you want to rotate, which displays the secrets details page.
4. Under `Rotation configuration`, choose Edit `rotation`.
5. From `Edit rotation configuration`, choose `Enable automatic rotation`.
6. For `Select Rotation Interval`, choose a rotation interval.
7. Choose a Lambda function for rotation. For information about customizing your Lambda rotation function, see [Understanding and customizing your Lambda rotation function](https://docs.aws.amazon.com/secretsmanager/latest/userguide/rotating-secrets-lambda-function-customizing.html).
8. To configure the secret for rotation, choose `Next`.