## Description

Multi-factor Authentication (MFA) adds an extra layer of protection on top of a username and password. With MFA enabled, when a user signs in to an AWS console, they will be prompted for their username and password as well as for an authentication code from their virtual or physical MFA device. It is recommended that MFA to be enabled for all users that have a console password.

Enabling MFA provides increased security for console access as it requires the authenticating principal to possess a device that creates a time-sensitive key and have knowledge of a credential.

## Remediation

### From Console

Perform the following action to enabled virtual MFA for the intended user:

1. Sign into the AWS console, and navigate to [IAM Console](https://console.aws.amazon.com/iam/home#/).
2. In the left navigation pane, choose Users.
3. In the user name list, choose the **name** of the intended user.
4. Choose the **Security credentials** tab, and then choose **Manage** for `Assigned MFA Device`.
5. In the Manage MFA device wizard, choose **virtual MFA device** and click on **continue**.
6. IAM generates and displays configuration information for the virtual MFA device, including a QR code graphic.
7. Open your virtual MFA application. (For a list of apps that you can use for hosting virtual MFA devices, see [Virtual MFA Applications](https://aws.amazon.com/iam/features/mfa/?audit=2019q1#Virtual_MFA_Applications). If the virtual MFA application supports multiple accounts (multiple virtual MFA devices), choose the option to create a new account (a new virtual MFA device).
8. Determine whether the MFA app supports QR codes, and then do one of the following:
    - Use the app to scan the QR code. For example, you might choose the camera icon or choose an option similar to Scan code, and then use the device's camera to scan the code.
    - In the Manage MFA Device wizard, choose Show secret key for manual configuration, and then type the secret configuration key into your MFA application.
9. Once you configure, the virtual MFA device starts generating MFA codes.
10. Type two consecutive MFA codes, MFA code 1 and MFA code 2 fields. Then click **Assign MFA**. Now the virtual MFA is enabled for the AWS account.