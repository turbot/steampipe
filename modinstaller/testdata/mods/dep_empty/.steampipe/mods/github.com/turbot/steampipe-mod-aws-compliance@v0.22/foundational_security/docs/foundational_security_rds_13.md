## Description

This control checks whether automatic minor version upgrades are enabled for the RDS database instance.

Enabling automatic minor version upgrades ensures that the latest minor version updates to the relational database management system (RDBMS) are installed. These upgrades might include security patches and bug fixes. Keeping up to date with patch installation is an important step in securing systems.

## Remediation

You can enable minor version upgrades for a DB instance from the Amazon RDS console.

**To enable automatic minor version upgrades for an existing DB instance**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. Choose `Databases`.
3. Choose the DB instance to modify.
4. Choose `Modify`.
5. Under `Maintenance`, select `Yes` for `Auto minor version upgrade`.
6. Choose `Continue`.
7. Under `Scheduling of modifications`, choose when to apply modifications: `Apply during the next scheduled maintenance window` or `Apply immediately`.
8. Choose `Modify DB Instance`.