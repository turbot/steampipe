## Description

AWS KMS enables customers to rotate the backing key, which is key material stored in AWS KMS and is tied to the key ID of the CMK. It's the backing key that is used to perform cryptographic operations such as encryption and decryption. Automated key rotation currently retains all previous backing keys so that decryption of encrypted data can take place transparently.

Rotating encryption keys helps reduce the potential impact of a compromised key as data encrypted with a new key cannot be accessed with a previous key that may have been exposed.

## Remediation

Perform the following to configure key rotation:

### From Console

1. Open the AWS KMS console at [KMS](https://console.aws.amazon.com/kms).
2. In the left navigation pane, choose `Customer managed keys`.
3. Choose the alias of the key to `update` in the Alias column.
4. Under the **Key rotation** section, move down to Key Rotation .
5. Select `Automatically rotate this CMK every year` and then choose Save.

### From Command Line

1. Run the following command to get a list of all keys and their associated KeyIds

```bash
 aws kms list-keys
```

2. For each key, note the KeyId and run the following command

```bash
 aws kms get-key-rotation-status --key-id <kms_key_id>
 ```
 
3. Ensure KeyRotationEnabled is set to true
