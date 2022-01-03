## Description

This control checks whether RDS clusters have deletion protection enabled.

This control is intended for RDS DB instances. However, it can also generate findings for Aurora DB instances, Neptune DB instances, and Amazon DocumentDB clusters. If these findings are not useful, then you can suppress them.

Enabling cluster deletion protection is an additional layer of protection against accidental database deletion or deletion by an unauthorized entity.

When deletion protection is enabled, an RDS cluster cannot be deleted. Before a deletion request can succeed, deletion protection must be disabled.

## Remediation

To remediate this issue, update your RDS DB cluster to enable delete protection.

**To enable deletion protection for an RDS DB cluster**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. In the navigation pane, choose `Databases`, then choose the DB cluster that you want to modify.
3. Choose `Modify`.
4. Under `Deletion protection`, choose `Enable deletion protection`.
5. Choose `Continue`.
6. Under `Scheduling of modifications`, choose when to apply modifications. The options are `Apply during the next scheduled maintenance window` or `Apply immediately`.
7. Choose `Modify Cluster`.