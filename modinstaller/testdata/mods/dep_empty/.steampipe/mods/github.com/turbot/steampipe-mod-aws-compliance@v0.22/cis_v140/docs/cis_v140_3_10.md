## Description

S3 object-level API operations such as GetObject, DeleteObject, and PutObject are called data events. By default, CloudTrail trails don't log data events and so it is recommended to enable Object-level logging for S3 buckets.

Enabling object-level logging will help you meet data compliance requirements within your organization, perform comprehensive security analysis, monitor specific patterns of user behavior in your AWS account or take immediate actions on any object-level API activity within your S3 Buckets using Amazon CloudWatch Events.

## Remediation

### From Console

1. Open the Amazon S3 console [S3](https://console.aws.amazon.com/s3/)
2. Choose the required bucket from the bucket list.
3. Choose **Properties** tab to see in detail bucket configuration.
4. Navigate to `AWS CloudTrail data events` section to select the CloudTrail name for the recording activity.
5. You can choose an existing Cloudtrail or create a new one by navigating to the Cloudtrail console link from S3.
6. Once the Cloudtrail console, navigate to `Data events : S3`  section.
7. If the current status for Object-level logging is set to Disabled, then object-level logging of write events for the selected s3 bucket is not set
   - Select **Edit** to enable the `Write` event.
   - You can choose to select `All current and future S3 buckets` or `Individual bucket`.
8. Repeat steps 2 to 7 to enable object-level logging of read events for other S3 buckets.

### From Command Line

### From Command Line

1. To enable object-level data events logging for S3 buckets within your AWS account, run put-event-selectors command using the name of the trail that you want to reconfigure as identifier:

```bash
aws cloudtrail put-event-selectors --region <region-name> --trail-name <trail-name> --event-selectors '[{ "ReadWriteType": "WriteOnly", "IncludeManagementEvents":true, "DataResources": [{ "Type": "AWS::S3::Object", "Values": ["arn:aws:s3:::<s3-bucket-name>/"] }] }]'
```

2. The command output will be object-level event trail configuration.
3. If you want to enable it for all buckets at once then change Values parameter to `["arn:aws:s3"]` in command given above.
4. Repeat step 1 for each s3 bucket to update `object-level` logging of write events.
5. Change the AWS region by updating the --region command parameter and perform the process for other regions.
