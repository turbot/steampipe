## Description

Checks whether the default version of IAM customer managed policies allow principals to use the AWS KMS decryption actions on all resources. This control uses [Zelkova](http://aws.amazon.com/blogs/security/protect-sensitive-data-in-the-cloud-with-automated-reasoning-zelkova/), an automated reasoning engine, to validate and warn you about policies that may grant broad access to your secrets across AWS accounts.

This control fails if the `kms:Decrypt` or `kms:ReEncryptFrom` actions are allowed on all KMS keys. The control evaluates both attached and unattached customer managed policies. It does not check inline policies or AWS managed policies.

With AWS KMS, you control who can use your customer master keys (CMKs) and gain access to your encrypted data. IAM policies define which actions an identity (user, group, or role) can perform on which resources. Following security best practices, AWS recommends that you allow least privilege. In other words, you should grant to identities only the `kms:Decrypt` or `kms:ReEncryptFrom` permissions and only for the keys that are required to perform a task. Otherwise, the user might use keys that are not appropriate for your data.

Instead of granting permissions for all keys, determine the minimum set of keys that users need to access encrypted data. Then design policies that allow users to use only those keys. For example, do not allow `kms:Decrypt` permission on all KMS keys. Instead, allow `kms:Decrypt` only on keys in a particular Region for your account. By adopting the principle of least privilege, you can reduce the risk of unintended disclosure of your data.

## Remediation

To remediate this issue, you modify the IAM customer managed policies to restrict access to the keys.

**To modify an IAM customer managed policy**

1. Open the [IAM console](https://console.aws.amazon.com/iam/).
2. In the IAM navigation pane, choose `Policies`.
3. Choose the arrow next to the policy you want to modify.
4. Choose `Edit policy`.
5. Choose the `JSON` tab.
6. Change the “Resource” value to the specific key or keys that you want to allow.
7. After you modify the policy, choose `Review policy`.
8. Choose `Save changes`.