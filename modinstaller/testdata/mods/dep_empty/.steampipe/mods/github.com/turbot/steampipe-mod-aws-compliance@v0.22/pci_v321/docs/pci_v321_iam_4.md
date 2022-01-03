## Description

This control checks whether your AWS account is enabled to use multi-factor authentication (MFA) hardware device to sign in with root user credentials.

It does not check whether you are using virtual MFA.

To address PCI DSS requirement 8.3.1, you can choose between hardware MFA (this control) or virtual MFA `PCI.IAM.5`(Virtual MFA should be enabled for the root user).

## Remediation

To enable hardware-based MFA for the root account

1. Log in to your account using the root user credentials.
2. Choose the account name at the top right of the page and then choose **My Security Credentials**.
3. In the warning, choose **Continue to Security Credentials**.
4. Choose **Multi-factor authentication (MFA)**.
5. Choose **Activate MFA**.
6. Choose a hardware-based (not virtual) device to use for MFA and then choose **Continue**.
7. Complete the steps to configure the device type appropriate to your selection.
