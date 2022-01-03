## Description

The root user account is the most privileged user in an AWS account. Multi-factor Authentication (MFA) adds an extra layer of protection on top of a username and password. With MFA enabled, when a user signs in to an AWS console, they will be prompted for their username and password as well as for an authentication code from their MFA device.

For Level 2, it is recommended that the root user account can be protected with a hardware MFA.

It is recommended that the device which is used for virtual MFA is NOT a personal device, but rather a dedicated device (phone or tablet). That can be managed to be kept charged and secured. It reduces the risks of losing access to the MFA code.

IAM *root user* account for us-gov cloud regions does not have console access. This control is not applicable for us-gov cloud regions.

A hardware MFA has a smaller attack surface than a virtual MFA. For example, a hardware MFA does not suffer the attack surface introduced by the smartphone or tablet on which a virtual MFA resides.

Using hardware MFA for many AWS accounts can create a logistical device management issue. In such cases, consider only implementing this Level 2 recommendation selectively to the highest secured AWS accounts and the Level 1 recommendation applied to the remaining accounts.

## Remediation

### From Console

Perform the following action to enabled hardware MFA for the root user account:

1. Sign into the AWS console as a root user, and navigate to [Your Security Credentials](https://console.aws.amazon.com/iam/home#/security_credentials).
2. Click on Multi-factor authentication (MFA) section and click **Activate MFA**.
3. In the Manage MFA device wizard, choose **Other hardware MFA device** and click on **continue**.
4. In the **Serial number** field, enter the serial number that is found on the back of the MFA device.
5. Press the button on the front of the device and type the 6-digit number that appears in **MFA code 1** field.
6. Wait for 30 seconds and then press the button again. Type the second number in **MFA code 2** field.
7. Click **Assign MFA**. Now the MFA device is assigned to the AWS account.
