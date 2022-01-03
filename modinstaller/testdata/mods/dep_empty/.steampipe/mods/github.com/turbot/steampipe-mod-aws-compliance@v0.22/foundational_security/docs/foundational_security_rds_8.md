## Description

This control checks whether your RDS DB instances that use one of the listed database engines have deletion protection enabled.

Enabling instance deletion protection is an additional layer of protection against accidental database deletion or deletion by an unauthorized entity.

While deletion protection is enabled, an RDS DB instance cannot be deleted. Before a deletion request can succeed, deletion protection must be disabled.

## Remediation

To remediate this issue, update your RDS DB instance to enable deletion protection.

**To enable deletion protection for an RDS DB instance**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. In the navigation pane, choose `Databases`, then choose the DB instance that you want to modify.
3. Choose `Modify`.
4. Under `Deletion protection`, choose `Enable deletion protection`.
5. Choose `Continue`.
6. Under `Scheduling of modifications`, choose when to apply modifications. The options are `Apply during the next scheduled maintenance window` or `Apply immediately`.
7. Choose `Modify DB Instance`.