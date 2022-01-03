## Description

This control checks whether AWS KMS customer managed keys (CMK) are scheduled for deletion. The control fails if a CMK is scheduled for deletion.

CMKs cannot be recovered once deleted. Data encrypted under a KMS CMK is also permanently unrecoverable if the CMK is deleted. If meaningful data has been encrypted under a CMK scheduled for deletion, consider decrypting the data or re-encrypting the data under a new CMK unless you are intentionally performing a cryptographic erasure.

When a CMK is scheduled for deletion, a mandatory waiting period is enforced to allow time to reverse the deletion, if it was scheduled in error. The default waiting period is 30 days, but it can be reduced to as short as 7 days when the KMS CMK is scheduled for deletion. During the waiting period, the scheduled deletion can be canceled and the KMS CMK will not be deleted.

## Remediation

For detailed remediation instructions to cancel a scheduled KMS CMK deletion, see To cancel key deletion under [Scheduling and canceling key deletion](https://docs.aws.amazon.com/kms/latest/developerguide/deleting-keys.html#deleting-keys-scheduling-key-deletion).