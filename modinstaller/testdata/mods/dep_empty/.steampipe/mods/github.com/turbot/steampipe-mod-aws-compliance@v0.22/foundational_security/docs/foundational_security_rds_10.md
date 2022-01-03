## Description

This control checks whether an RDS DB instance has IAM database authentication enabled.

IAM database authentication allows authentication to database instances with an authentication token instead of a password. Network traffic to and from the database is encrypted using SSL.

## Remediation

To remediate this issue, update your DB instance to enable IAM authentication.

**To enable IAM authentication for an existing DB instance**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. Choose `Databases`.
3. Select the DB instance to modify.
4. Choose `Modify`.
5. Under `Database options`, choose `Enable IAM DB authentication`.
6. Choose `Continue`.
7. Under `Scheduling of modifications`, choose when to apply modifications. The options are `Apply during the next scheduled maintenance window` or `Apply immediately`.
8. For clusters, choose `Modify DB Instance.`