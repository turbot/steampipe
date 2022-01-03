## Description

Checks whether the inline policies that are embedded in your IAM identities (role, user, or group) allow the AWS KMS decryption actions on all KMS keys. This control uses Zelkova, an automated reasoning engine, to validate and warn you about policies that may grant broad access to your secrets across AWS accounts.

This control fails if `kms:Decrypt` or `kms:ReEncryptFrom` actions are allowed on all KMS keys in an inline policy.

With AWS KMS, you control who can use your customer master keys (CMKs) and gain access to your encrypted data. IAM policies define which actions an identity (user, group, or role) can perform on which resources. Following security best practices, AWS recommends that you allow least privilege. In other words, you should grant to identities only the permissions they need and only for keys that are required to perform a task. Otherwise, the user might use keys that are not appropriate for your data.

Instead of granting permission for all keys, determine the minimum set of keys that users need to access encrypted data. Then design policies that allow the users to use only those keys. For example, do not allow `kms:Decrypt` permission on all KMS keys. Instead, allow them only on keys in a particular Region for your account. By adopting the principle of least privilege, you can reduce the risk of unintended disclosure of your data.


## Remediation

To remediate this issue, you modify the inline policy to restrict access to the keys.

**To modify an IAM inline policy**

1. Open the [IAM console](https://console.aws.amazon.com/iam/).
2. In the IAM navigation pane, choose `Users`, `Groups`, or `Roles`.
3. Choose the name of the user, group or role for which to modify IAM inline policies.
4. Choose the arrow next to the policy to modify.
5. Choose `Edit policy`.
6. Choose the `JSON` tab.
7. Change the “Resource” value to the specific keys that you want to allow.
8. After you modify the policy, choose `Review policy`.
9. Choose `Save changes`.