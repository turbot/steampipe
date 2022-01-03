## Description

Access Analyzer generates a finding when a policy on a resource within your zone of trust allows access from outside your zone of trust.
Enable IAM Access analyzer for IAM policies about all resources. After the Analyzer is enabled in IAM, scan results are displayed on the console.

AWS IAM Access Analyzer helps you identify the resources in your organization and accounts, such as Amazon S3 buckets or IAM roles, that are shared with an external entity. This lets you identify unintended access to your resources and data. IAM Access Analyzer continuously monitors all policies for S3 bucket, IAM roles, KMS(Key Management Service) keys, AWS Lambda functions, and Amazon SQS(Simple Queue Service) queues.

## Remediation

### From Console

Perform the following to enable IAM Access analyzer for IAM policies:

1. Sign into the AWS console and open the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, choose **Access analyzer**.
3. Click **Create analyzer**.
4. On the `Create analyzer` page, confirm that the region displayed is the region where you want to enable Access Analyzer.
5. Enter a name for the analyzer or can keep the system generated.
6. Optional. add any tags that you want to apply to the analyzer.
7. Choose **Create analyzer**.

### From Command Line

Run the following command:

```bash
aws accessanalyzer create-analyzer --analyzer-name --type
```

**Note**: The type of analyzer to create. Only ACCOUNT and ORGANIZATION analyzers are supported. You can create only one analyzer per account per Region. You can create up to 5 analyzers per organization per Region.
