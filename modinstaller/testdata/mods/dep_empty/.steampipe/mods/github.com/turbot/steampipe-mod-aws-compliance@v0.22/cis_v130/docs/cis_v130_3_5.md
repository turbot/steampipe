## Description

AWS Config is a web service that performs configuration management of supported AWS resources in your account and delivers log files to you. The recorded information includes the configuration item (AWS resource), relationships between configuration items (AWS resources), and any configuration changes between resources.

The AWS configuration item history captured by AWS Config enables security analysis, resource change tracking, and compliance auditing.

## Remediation

To implement AWS Config configuration:

### From Console

1. Open the AWS Config console at [Config](https://console.aws.amazon.com/config/).
2. Select the Region to configure AWS Config in.
3. On the Settings page, do the following:
    - Under Resource types to record, select Record all resources supported in this region and Include global resources (e.g., AWS IAM resources).
    - Under Amazon S3 bucket, specify the bucket to use or create a bucket and optionally include a prefix.
    - Under Amazon SNS topic, select an Amazon SNS topic from your account or create one.
    - Under AWS Config role, either choose Create AWS Config service-linked role or choose Choose a role from your account and then select the role to use.
4. Choose Next.
5. On the AWS Config rules page, choose Skip.
6. Choose **Confirm**.

### From Command Line

1. Ensure there is an appropriate S3 bucket, SNS topic, and IAM role per the AWS Config Service prerequisites.
2. Run this command to set up the configuration recorder

```bash
aws configservice subscribe --s3-bucket my-config-bucket --sns-topic arn:aws:sns:us-east-1:012345678912:my-config-notice --iam-role arn:aws:iam::012345678912:role/myConfigRole
```

3. Run this command to start the configuration recorder:

```bash
start-configuration-recorder --configuration-recorder-name <value>
```
