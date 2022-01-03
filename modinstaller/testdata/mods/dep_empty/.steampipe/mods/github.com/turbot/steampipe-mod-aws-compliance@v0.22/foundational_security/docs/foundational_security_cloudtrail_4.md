## Description

This control checks whether log file integrity validation is enabled on a CloudTrail trail.

CloudTrail log file validation creates a digitally signed digest file that contains a hash of each log that CloudTrail writes to Amazon S3. You can use these digest files to determine whether a log file was changed, deleted, or unchanged after CloudTrail delivered the log.

Security Hub recommends that you enable file validation on all trails. Log file validation provides additional integrity checks of CloudTrail logs.

## Remediation

To remediate this issue, update your CloudTrail trail to enable log file validation.

**To enable CloudTrail log file validation**

1. Open the [CloudTrail console](https://console.aws.amazon.com/cloudtrail/).
2. Choose `Trails`.
3. Under `Name`, choose the name of a trail to edit.
4. Under `General details`, choose `Edit`.
5. Under `Additional settings`, for Log file validation, choose `Enabled`.
6. Choose `Save changes`.