## Description

CloudTrail log file validation creates a digitally signed digest file containing a hash of each log that CloudTrail writes to S3. These digest files can be used to determine whether a log file was changed, deleted, or unchanged after CloudTrail delivered the log. It is recommended that file validation be enabled on all CloudTrails.

The AWS API call history produced by CloudTrail enables security analysis, resource change tracking, and compliance auditing. Additionally, enabling log file validation will provide additional integrity checking of CloudTrail logs.

## Remediation

Perform the following to enable global (Multi-region) CloudTrail logging:

### From Console

1. Sign in to the AWS Management Console and open the IAM console at [cloudtrail](https://console.aws.amazon.com/cloudtrail)
2. Click on `Trails` on the left navigation pane
3. Click on target trail
4. Within the `S3` section click on the edit icon (pencil)
5. Click `Advanced`
6. Click on the **Yes** radio button in section Enable `log file validation`
7. Click Save

### From Command Line

```bash
aws cloudtrail update-trail --name <trail_name> --enable-log-file-validation
```

**Note**:  that periodic validation of logs using these digests can be performed by running the
following command:

```bash
aws cloudtrail validate-logs --trail-arn <trail_arn> --start-time <start_time> --end-time <end_time>
```
