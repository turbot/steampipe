## Description

CloudTrail is a web service that records AWS API calls made in a given account. The recorded information includes the identity of the API caller, the time of the API call, the source IP address of the API caller, the request parameters, and the response elements returned by the AWS service.

CloudTrail uses Amazon S3 for log file storage and delivery, so log files are stored durably. In addition to capturing CloudTrail logs in a specified Amazon S3 bucket for long-term analysis, you can perform real-time analysis by configuring CloudTrail to send logs to CloudWatch Logs.

For a trail that is enabled in all Regions in an account, CloudTrail sends log files from all those Regions to a CloudWatch Logs log group.

Sending CloudTrail logs to CloudWatch Logs will facilitate real-time and historic activity logging based on user, API, resource, and IP address, and provides opportunity to establish alarms and notifications for anomalous or sensitivity account activity.

## Remediation

To ensure that CloudTrail trails are integrated with CloudWatch Logs, perform the following to establish the prescribed state:

### From Console

1. Open the CloudTrail console at [CloudTrail](https://console.aws.amazon.com/cloudtrail/).
2. Choose Trails.
3. Choose a trail that there is no value for in the CloudWatch Logs Log group column.
4. Scroll down to the `CloudWatch Logs` section and then choose Edit.
5. Select the `Enabled` check box.
6. For Log group field, do one of the following:
    - To use the default log group, keep the name as is.
    - To use an existing log group, choose Existing and then enter the name of the log group to use.
    - To create a new log group, choose New and then enter a name for the log group to create.
7. For IAM role, do one of the following:
    - To use an existing role, choose Existing and then choose the role from the drop-down list.
    - To create a new role, choose New and then enter a name for the role to create. The new role is assigned a policy that grants the necessary permissions.
    - To view the permissions granted to the role, expand the Policy document.
8. Choose Save changes.

### From Command Line

```bash
aws cloudtrail update-trail --name <trail_name> --cloudwatch-logs-log-grouparn <cloudtrail_log_group_arn> --cloudwatch-logs-role-arn <cloudtrail_cloudwatchLogs_role_arn>
```
