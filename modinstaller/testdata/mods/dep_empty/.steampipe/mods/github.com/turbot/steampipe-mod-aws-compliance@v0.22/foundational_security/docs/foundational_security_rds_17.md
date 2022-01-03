## Description

This control checks whether RDS DB instances are configured to copy all tags to snapshots when the snapshots are created.

Identification and inventory of your IT assets is a crucial aspect of governance and security. You need to have visibility of all your RDS DB instances so that you can assess their security posture and take action on potential areas of weakness. Snapshots should be tagged in the same way as their parent RDS database instances. Enabling this setting ensures that snapshots inherit the tags of their parent database instances.

## Remediation

**To enable automatic tag copying to snapshots for a DB instance**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. Choose `Databases`.
3. Select the DB instance to modify.
4. Choose `Modify`.
5. Under `Backup`, select `Copy tags to snapshots`.
6. Choose `Continue`.
7. Under `Scheduling of modifications`, choose when to apply modifications. You can choose either      `Apply during the next scheduled maintenance window` or `Apply immediately`.