## Description

This control checks whether CloudTrail is enabled in your AWS account.

However, some AWS services do not enable logging of all APIs and events. You should implement any additional audit trails other than CloudTrail and review the documentation for each service in CloudTrail Supported Services and Integrations.

## Remediation

To create a new trail in CloudTrail

1. Sign in to the AWS Management Console using the IAM user you configured for CloudTrail administration.
1. Open the CloudTrail console at [CloudTrail](https://console.aws.amazon.com/cloudtrail/).
1. In the Region selector, choose the AWS Region where you want your trail to be created. This is the Home Region for the trail.
1. The Home Region is the only AWS Region where you can view and update the trail after it is created, even if the trail logs events in all AWS Regions.
1. In the navigation pane, choose **Trails**.
1. On the Trails page, choose **Get Started Now**. If you do not see that option, choose **Create Trail**.
1. In Trail name, give your trail a name, such as My-Management-Events-Trail.
1. As a best practice, use a name that quickly identifies the purpose of the trail. In this case, you're creating a trail that logs management events.
1. In Management Events, make sure Read/Write events is set to **All**.
1. In Data Events, do not make any changes. This trail will not log any data events.
1. Create a new S3 bucket for the logs:
    1. In Storage Location, in Create a new S3 bucket, choose **Yes**.
    1. In S3 bucket, give your bucket a name, such as my-bucket-for-storing-cloudtrail-logs.
    1. The name of your S3 bucket must be globally unique. For more information about S3 bucket naming requirements, see the [AWS CloudTrail User Guide](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-s3-bucket-naming-requirements.html).
1. Under Advanced, choose **Yes** for both Encrypt log files with SSE-KMS and Enable log file validation.
1. Choose **Create**.