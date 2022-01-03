## Description

This control checks for the CloudWatch metric filters using the following pattern:

```
{ $.userIdentity.type = "Root" && $.userIdentity.invokedBy NOT EXISTS && $.eventType != "AwsServiceEvent" }
```

It checks the following:

  - The log group name is configured for use with active multi-Region CloudTrail.
  - There is at least one Event Selector for a Trail with IncludeManagementEvents set to true and ReadWriteType set to All.
  - There is at least one active subscriber to an Amazon SNS topic associated with the alarm.

## Remediation

The steps to remediate this issue include setting up an Amazon SNS topic, a metric filter, and an alarm for the metric filter.

To create an Amazon SNS topic

1. Open the Amazon SNS console at https://console.aws.amazon.com/sns/v3/home.
2. Create an Amazon SNS topic that receives all CIS alarms.
3. Create at least one subscriber to the topic.
4. For more information about creating Amazon SNS topics, see the Amazon Simple Notification Service Developer Guide.
5. Set up an active CloudTrail trail that applies to all Regions.
6. To do this, follow the remediation steps in CIS v1.3.0 [3.1 Ensure CloudTrail is enabled in all Regions](https://hub.steampipe.io/mods/turbot/aws_compliance/controls/control.cis_v130_3_1).
7. Make a note of the associated log group name.

To create a metric filter and alarm

1. Open the [CloudWatch console](https://console.aws.amazon.com/cloudwatch/).
2. Choose Logs, then choose **Log groups**.
3. Choose the log group where CloudTrail is logging.
4. On the log group details page, choose **Metric filters**.
5. Choose **Create metric filter**.
6. Copy the following pattern and then paste it into Filter pattern.

   ```
    {$.userIdentity.type="Root" && $.userIdentity.invokedBy NOT EXISTS && $.eventType !="AwsServiceEvent"}
   ```
7. Enter the name of the new filter. For example, RootAccountUsage.
8. Confirm that the value for **Metric namespace** is `LogMetrics`.
9.  This ensures that all CIS Benchmark metrics are grouped together.
10. In **Metric name**, enter the name of the metric.
11. In Metric value, enter 1, and then choose **Next**.
12. Choose **Create metric filter**.
13. Next, set up the notification. Select the select the metric filter you just created, then choose **Create alarm**.
14. Enter the threshold for the alarm (for example, 1), then choose **Next**.
15. Under Select an SNS topic, for Send notification to, choose an email list, then choose Next.
16. Enter a **Name and Description** for the alarm, such as `RootAccountUsageAlarm`, then choose **Next**.
17. Choose **Create Alarm**.
