## Description

Macie along with other 3rd party tools can be used to discover, monitor, classify, and inventory S3 buckets.

Using a Cloud service or 3rd Party software to continuously monitor and automate the process of data discovery and classification for S3 buckets using machine learning and pattern matching is a strong defense in protecting that information.

## Remediation

### From Console

1. Enable Macie through the [Macie console](https://console.aws.amazon.com/macie/).
2. Create an S3 bucket to use as a repository for sensitive data discovery results.
3. Select the buckets you want Macie to analyze and then create a job.
4. After the job has run, review the findings by selecting **Findings** in the left pane.
