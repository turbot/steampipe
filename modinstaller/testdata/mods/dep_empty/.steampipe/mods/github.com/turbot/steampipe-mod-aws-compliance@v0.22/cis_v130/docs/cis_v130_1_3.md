## Description

Ensure *Security Challenge Questions* are set up in the AWS account settings page of your AWS account.

By adding security challenge questions improve the security of your account. It can be used to authenticate individuals calling AWS customer service for support. It is highly recommended that security questions be established.

When creating a new AWS account, a default super user is automatically created. This account is referred to as the "root user" account. It is recommended that the use of this account to be limited. During events in case the root password is no longer accessible or the MFA token associated with root user is lost or destroyed it is possible, through authentication using secret questions and associated answers, root user login access can be recovered.

## Remediation

There is no API available for setting security questions - you must log in to the AWS console to verify, and set your security questions and answers.

1. Sign into the AWS console as a root user, and navigate to the [Account Settings](https://console.aws.amazon.com/billing/home?#/account).
2. Verify that the information in the **Configure Security Challenge Questions** section is complete and you have the answers with you. If changes are required, click **Edit**, make your changes, and then click **Update**.
3. Keep the questions and answers in a secure location.