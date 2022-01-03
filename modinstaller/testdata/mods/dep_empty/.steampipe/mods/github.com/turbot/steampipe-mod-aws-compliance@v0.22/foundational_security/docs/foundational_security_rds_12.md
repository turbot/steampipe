## Description

This control checks whether an RDS DB cluster has IAM database authentication enabled.

IAM database authentication allows for password-free authentication to database instances. The authentication uses an authentication token. Network traffic to and from the database is encrypted using SSL.

## Remediation

You can enable IAM authentication for a DB cluster from the Amazon RDS console.

**To enable IAM authentication for an existing DB cluster**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. Choose `Databases`.
3. Choose the DB cluster to modify.
4. Choose `Modify`.
5. Under `Database options`, select `Enable IAM DB authentication`.
6. Choose `Continue`.
7. Under `Scheduling of modifications`, choose when to apply modifications: `Apply during the next scheduled maintenance window` or `Apply immediately`.
8. Choose `Modify cluster`.