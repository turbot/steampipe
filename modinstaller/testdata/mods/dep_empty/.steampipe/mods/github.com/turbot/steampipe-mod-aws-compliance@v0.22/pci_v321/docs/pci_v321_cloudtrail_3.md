## Description

This control checks whether CloudTrail log file validation is enabled.

It does not check when configurations are altered.

To monitor and alert on log file changes, you can use Amazon EventBridge or CloudWatch metric filters.

## Remediation

To enable CloudTrail log file validation

1. Open the CloudTrail console at [CloudTrail](https://console.aws.amazon.com/cloudtrail/).
1. In the navigation pane, choose **Trails**.
1. In the Name column, choose the **Trail Name** to edit.
1. Under General details, choose **Edit**.
1. Under Additional settings, for Log file validation,, select **Enabled**.
1. Choose **Save**.