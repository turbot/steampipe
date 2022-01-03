## Description

This control checks whether Amazon RDS DB snapshots prohibit access by other accounts. You should also ensure that access to the snapshot and permission to change Amazon RDS configuration is restricted to authorized principals only.

**Note** that if the configuration is changed to allow public access, the AWS Config rule may not be able to detect the change for up to 12 hours. Until the AWS Config rule detects the change, the check passes even though the configuration violates the rule.

## Remediation

To remove public access for Amazon RDS Databases

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. Navigate to Snapshots and then select the public Snapshot you want to modify
3. From the **Actions** list, choose **Share Snapshots**
4. From **DB snapshot visibility**, choose **Private**
5. Under **DB snapshot visibility**, select **for all**
6. Choose **Save**
