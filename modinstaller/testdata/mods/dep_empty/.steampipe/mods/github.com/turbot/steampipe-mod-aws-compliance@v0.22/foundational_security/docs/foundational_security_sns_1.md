## Description

This control checks whether an SNS topic is encrypted at rest using AWS KMS.

Encrypting data at rest reduces the risk of data stored on disk being accessed by a user not authenticated to AWS. It also adds another set of access controls to limit the ability of unauthorized users to access the data. For example, API permissions are required to decrypt the data before it can be read. SNS topics should be encrypted at-rest for an added layer of security. For more information, see [Encryption at rest](https://docs.aws.amazon.com/sns/latest/dg/sns-server-side-encryption.html) in the `Amazon Simple Notification Service Developer Guide`.

## Remediation

To remediate this issue, update your SNS topic to enable encryption.

**To encrypt an unencrypted SNS topic**

1. Open the [Amazon SNS console](https://console.aws.amazon.com/sns/v3/home).
2. In the navigation pane, choose `Topics`.
3. Choose the name of the topic to encrypt.
4. Choose `Edit`.
5. Under `Encryption`, choose `Enable Encryption`.
5. Choose the `KMS key` to use to encrypt the topic.
6. Choose `Save` changes.