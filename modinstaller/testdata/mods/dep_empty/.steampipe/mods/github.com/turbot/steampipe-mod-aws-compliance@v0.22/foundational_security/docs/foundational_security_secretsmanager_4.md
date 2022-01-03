## Description

This control checks whether your secrets have been rotated at least once within 90 days.

Rotating secrets can help you to reduce the risk of an unauthorized use of your secrets in your AWS account. Examples include database credentials, passwords, third-party API keys, and even arbitrary text. If you do not change your secrets for a long period of time, the secrets are more likely to be compromised.

As more users get access to a secret, it can become more likely that someone mishandled and leaked it to an unauthorized entity. Secrets can be leaked through logs and cache data. They can be shared for debugging purposes and not changed or revoked once the debugging completes. For all these reasons, secrets should be rotated frequently.

You can configure your secrets for automatic rotation in AWS Secrets Manager. With automatic rotation, you can replace long-term secrets with short-term ones, significantly reducing the risk of compromise.

Security Hub recommends that you enable rotation for your Secrets Manager secrets. To learn more about rotation, see [Rotating your AWS Secrets Manager secrets](https://docs.aws.amazon.com/secretsmanager/latest/userguide/rotating-secrets.html).

## Remediation

You can enable automatic secret rotation in the Secrets Manager console.

**To enable secret rotation**

1. Open the [Secrets Manager console](https://console.aws.amazon.com/secretsmanager/).
2. To locate the secret, enter the secret name in the search box.
3. Choose the secret to display.
4. Under `Rotation configuration`, choose `Edit rotation`.
5. From `Edit rotation configuration`, choose `Enable automatic rotation`.
6. From `Select Rotation Interval`, choose the rotation interval.
7. Choose a Lambda function to use for rotation.
8. Choose `Next`.
9. After you configure the secret for automatic rotation, under `Rotation Configuration`, choose `Rotate secret immediately`.