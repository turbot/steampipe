## Description

This control checks whether CloudTrail trails are configured to send logs to CloudWatch Logs. The control fails if the CloudWatchLogsLogGroupArn property of the trail is empty.

CloudTrail records AWS API calls that are made in a given account. The recorded information includes the following:

- The identity of the API caller
- The time of the API call
- The source IP address of the API caller
- The request parameters
- The response elements returned by the AWS service

CloudTrail uses Amazon S3 for log file storage and delivery. You can capture CloudTrail logs in a specified S3 bucket for long-term analysis. To perform real-time analysis, you can configure CloudTrail to send logs to CloudWatch Logs.

For a trail that is enabled in all Regions in an account, CloudTrail sends log files from all of those Regions to a CloudWatch Logs log group.

Security Hub recommends that you send CloudTrail logs to CloudWatch Logs. Note that this recommendation is intended to ensure that account activity is captured, monitored, and has appropriately alarms. You can use CloudWatch Logs to set this up with your AWS services. This recommendation does not preclude the use of a different solution.

Sending CloudTrail logs to CloudWatch Logs facilitates real-time and historic activity logging based on user, API, resource, and IP address. You can use this approach to establish alarms and notifications for anomalous or sensitivity account activity.

## Remediation

To enable CloudTrail integration with CloudWatch Logs

1. Open the [CloudTrail console](https://console.aws.amazon.com/cloudtrail/).
2. Choose `Trails`.
3. Choose the trail that does not have a value for `CloudWatch Logs Log group`.
4. Under `CloudWatch Logs`, choose `Edit`.
5. Select `Enabled`.
6. For `Log group`, do one of the following:
   - To use the default log group, keep the name as is.
   - To use an existing log group, choose `Existing` and then enter the name of the log group to use.
   - To create a new log group, choose `New` and then enter a name for the log group to create.
7. For `IAM role`, do one of the following:
   - To use an existing role, choose `Existing` and then choose the role from the drop-down list.
   - To create a new role, choose `New` and then enter a name for the role to create. The new role is assigned a policy that grants the necessary permissions.
   - To view the permissions granted to the role, expand `Policy document`.
8. Choose `Save changes`.
