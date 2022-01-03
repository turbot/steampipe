## Description

This control checks that key rotation is enabled for each customer master key (CMK). It does not check CMKs that have imported key material.

You should ensure keys that have imported material and those that are not stored in AWS KMS are rotated. AWS managed customer master keys are rotated once every 3 years.

## Remediation

To enable CMK rotation

1. Open the [AWS KMS console](https://console.aws.amazon.com/kms).
2. To change the AWS Region, use the Region selector in the upper-right corner of the page.
3. Choose **Customer managed keys**.
4. In the Alias column, choose the alias of the key to update.
5. Choose **Key rotation**.
6. Select Automatically rotate this CMK every year and then choose **Save**.
