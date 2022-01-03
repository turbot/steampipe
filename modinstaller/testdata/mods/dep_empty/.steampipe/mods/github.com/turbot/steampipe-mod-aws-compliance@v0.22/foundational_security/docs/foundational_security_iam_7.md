## Description

To access the AWS Management Console, IAM users need passwords. As a best practice, Security Hub highly recommends that instead of creating IAM users, you use federation. Federation allows users to use their existing corporate credentials to log into the AWS Management Console. Use AWS Single Sign-On (AWS SSO) to create or federate the user, and then assume an IAM role into an account.

To learn more about identity providers and federation, see Identity providers and federation in the IAM User Guide. To learn more about AWS SSO, see the AWS Single Sign-On User Guide.

If you need to use IAM users, Security Hub recommends that you enforce the creation of strong user passwords. You can set a password policy on your AWS account to specify complexity requirements and mandatory rotation periods for passwords. When you create or change a password policy, most of the password policy settings are enforced the next time users change their passwords. Some of the settings are enforced immediately. To learn more, see Setting an account password policy for IAM users in the IAM User Guide.

## Remediation

To remediate this issue, update your password policy to use the recommended configuration.

1. Sign into the AWS console, and navigate to [IAM Console](https://console.aws.amazon.com/iam/home#/).
3. Choose `Account settings`.
4. Select `Requires at least one uppercase letter`.
5. Select `Requires at least one lowercase letter`.
6. Select `Requires at least one non-alphanumeric character`.
7. Select `Requires at least one number`.
8. For `Minimum password length`, enter `8`.
9. Choose `Apply password policy`.