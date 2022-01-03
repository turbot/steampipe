## Description

Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. Security Groups are a stateful packet filter that controls ingress and egress traffic within a VPC. It is recommended that a metric filter and alarm be established for detecting changes to Security Groups.

Monitoring changes to security group will help ensure that resources and services are not unintentionally exposed.

## Remediation

The steps to remediate this issue include setting up an Amazon SNS topic, a metric filter, and an alarm for the metric filter.

### From Console

**To create SNS topic**

1. Open the Amazon SNS console at [SNS](https://console.aws.amazon.com/sns/v3/home).
2. Create an Amazon SNS topic that receives all CIS alarms. Create at least one subscriber to the topic. You can follow steps [here](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/US_SetupSNS.html)
3. Set up an active CloudTrail that applies to all Regions. To do so, follow the remediation steps in CIS 3.1 – Ensure CloudTrail is enabled in all Regions.
4. Make a note of the associated log group name.

**To create a metric filter and alarm**

1. Open the CloudWatch console at [CloudWatch](https://console.aws.amazon.com/cloudwatch/).
2. In the navigation pane, choose **Log groups**.
3. Select the check box for the log group that you made a note of in the previous step (4).
4. From **Actions**, choose **Create Metric Filter**.
5. Under Define pattern, do the following:
   - Copy the following pattern and then paste it into the Filter Pattern field.
      ```
      {($.eventName = AuthorizeSecurityGroupIngress) || ($.eventName =AuthorizeSecurityGroupEgress) || ($.eventName = RevokeSecurityGroupIngress) || ($.eventName = RevokeSecurityGroupEgress) || ($.eventName = CreateSecurityGroup) || ($.eventName = DeleteSecurityGroup) }
      ```
   - Choose **Next**.
6. Under **Assign metric**, do the following:
   - In Filter name, enter a name for your metric filter.
   - For Metric namespace, enter `CISBenchmark`. You can use the same namespace for all of your CIS log metric filters.
   - For Metric name, enter a name for the metric.
   - The name of the metric is required to create the alarm.
   - For Metric value, enter 1.
   - Choose **Next**.
7. Under **Review and create**, review the details and choose **Create metric filter**.
8. Choose the **Metric filters** tab, select the metric filter that you just created.
9. To choose the metric filter, select the `check box` at the upper right.
10. Choose **Create Alarm**.
11. Follow the steps provided in [Create an alarm](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudwatch-alarms-for-cloudtrail.html)
12. Under Configure actions, do the following:
      - Under **Alarm state trigger**, choose In alarm.
      - Under Select an SNS topic, choose Select an existing SNS topic.
      - For Send a notification to, enter the name of the SNS topic that you created in the previous procedure.
      - Choose **Next**.
13. Under Add name and description, enter a Name and Description for the alarm. For example, CIS4.10-[SmallDescription]. Then choose **Next**.
14. Under **Preview and create**, review the alarm configuration. Choose **Create alarm**.

### From Command Line

Perform the following to setup the metric filter, alarm, SNS topic, and subscription

1. Create a metric filter based on filter pattern provided and CloudWatch log group.

```bash
aws logs put-metric-filter --log-group-name "<cloudtrail_log_group_name>" --filter-name "<security_group_changes_metric>" --metric-transformations metricName= "<security_group_changes_metric>",metricNamespace="CISBenchmark",metricValue=1 --filter-pattern "{($.eventName = AuthorizeSecurityGroupIngress) || ($.eventName =AuthorizeSecurityGroupEgress) || ($.eventName = RevokeSecurityGroupIngress) || ($.eventName = RevokeSecurityGroupEgress) || ($.eventName = CreateSecurityGroup) || ($.eventName = DeleteSecurityGroup) }"
```

**Note**: You can choose your own metricName and metricNamespace strings. Using the same metricNamespace for all Foundations Benchmark metrics will group them together.

2. Create an SNS topic that the alarm will notify.

```bash
aws sns create-topic --name <sns_topic_name>
```
**Note**: you can execute this command once and then re-use the same topic for all monitoring alarms.

3. Create an SNS subscription to the topic created in step 2.

```bash
aws sns subscribe --topic-arn <sns_topic_arn> --protocol <protocol_for_sns> --notification-endpoint <sns_subscription_endpoints>
```
**Note**: you can execute this command once and then re-use the SNS subscription for all.

4. Create an alarm that is associated with the CloudWatch Logs Metric Filter created in step 1 and an SNS topic created in step 2.

```bash
aws cloudwatch put-metric-alarm --alarm-name "<security_group_changes_alarm>" --metric-name "<security_group_changes_metric>" --statistic Sum --period 300 --threshold 1 --comparison-operator GreaterThanOrEqualToThreshold --evaluation-periods 1 --namespace "CISBenchmark" --alarm-actions "<sns_topic_arn>"
```
