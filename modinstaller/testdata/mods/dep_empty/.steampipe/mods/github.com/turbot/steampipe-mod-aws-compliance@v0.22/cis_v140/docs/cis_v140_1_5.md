## Description

The root user account is the most privileged user in an AWS account. Multi-factor Authentication (MFA) adds an extra layer of protection on top of a username and password. With MFA enabled, when a user signs in to an AWS console, they will be prompted for their username and password as well as for an authentication code from their MFA device.

It is recommended that the device which is used for virtual MFA is NOT a personal device, but rather a dedicated device (phone or tablet). That can be managed to be kept charged and secured. It reduces the risks of losing access to the MFA code.

IAM *root user* account for us-gov cloud regions does not have console access. This control is not applicable for us-gov cloud regions.

Enabling virtual MFA provides increased security for console access as it requires the authenticating principal to possess a device that creates a time-sensitive key and have knowledge of a credential.

## Remediation

### From Console

Perform the following action to enabled virtual MFA for the root user account:

1. Sign into the AWS console as a root user, and navigate to [Your Security Credentials](https://console.aws.amazon.com/iam/home#/security_credentials).
2. Click on Multi-factor authentication (MFA) section and click **Activate MFA**.
3. In the Manage MFA device wizard, choose **virtual MFA device** and click on **continue**.
4. IAM generates and displays configuration information for the virtual MFA device, including a QR code graphic.
5. Open your virtual MFA application. (For a list of apps that you can use for hosting virtual MFA devices, see [Virtual MFA Applications](https://aws.amazon.com/iam/features/mfa/?audit=2019q1#Virtual_MFA_Applications). If the virtual MFA application supports multiple accounts (multiple virtual MFA devices), choose the option to create a new account (a new virtual MFA device).
6. Determine whether the MFA app supports QR codes, and then do one of the following:
    - Use the app to scan the QR code. For example, you might choose the camera icon or choose an option similar to Scan code, and then use the device's camera to scan the code.
    - In the Manage MFA Device wizard, choose Show secret key for manual configuration, and then type the secret configuration key into your MFA application.
7. Once you configure, the virtual MFA device starts generating MFA codes.
8. Type two consecutive MFA codes, MFA code 1 and MFA code 2 fields. Then click **Assign MFA**. Now the virtual MFA is enabled for the AWS account.
