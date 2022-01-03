## Description

AWS S3 Bucket Access Logging generates a log that contains access records for each request made to your S3 bucket. An access log record contains details about the request, such as the request type, the resources specified in the request worked, and the time and date the request was processed. It is recommended that bucket access logging be enabled on the CloudTrail S3 bucket.

By enabling S3 bucket logging on target S3 buckets, it is possible to capture all events which may affect objects within any target buckets. Configuring logs to be placed in a separate bucket allows access to log information which can be useful in security and incident response workflows.

## Remediation

Perform the following to enable S3 bucket logging:

### From Console

1. Open the Amazon S3 console at [S3](https://console.aws.amazon.com/s3/).
2. Choose the bucket used for CloudTrail.
3. Choose **Properties**.
4. Choose **Server access logging**.
5. Choose **Edit**, then select `Enable`.
6. Select a bucket from the `Target bucket` list, and optionally enter a prefix.
7. Choose Save.
