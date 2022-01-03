## Description

Password policies are, in part, used to enforce password complexity requirements. IAM password policies can be used to ensure passwords are at least a given length. It is recommended that the password policy require a minimum password length of 14.

Setting a complex password policy increases account resiliency against unethical password hackers.

## Remediation

Perform the following to set the password policy is configured as prescribed:

### From Console:

1. Sign into the AWS console and navigate to the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. Choose **Account settings**.
3. Click **Change** or **Change password policy** (if no password policy set earlier).
4. Ensure in the `Enforce minimum password length` field is set to 14, then choose **Save changes**.

### From Command Line:

```bash
aws iam update-account-password-policy --minimum-password-length 14
```

**Note**: All commands starting with "aws iam update-account-password-policy" can be combined into a single command.
