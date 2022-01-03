## Description

IAM password policies can prevent the reuse of a given password by the same user. It is recommended that the password policy prevent the reuse of passwords.

Preventing password reuse increases account resiliency against unethical password hackers.

## Remediation

Perform the following to set the password policy as prescribed:

### From Console:

1. Sign into the AWS console and navigate to the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. Choose **Account settings**.
3. Click **Change** or **Change password policy** (if no password policy set earlier).
4. Ensure `Prevent password reuse` is checked.
5. Ensure `Remember password(s)` is set to 24 and then choose **Save changes**.

### From Command Line:

```bash
aws iam update-account-password-policy --password-reuse-prevention 24
```

**Note**: All commands starting with "aws iam update-account-password-policy" can be combined into a single command.
