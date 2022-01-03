## Description

This control checks whether point-in-time recovery (PITR) is enabled for an Amazon DynamoDB table.

Backups help you to recover more quickly from a security incident. They also strengthen the resilience of your systems. DynamoDB point-in-time recovery automates backups for DynamoDB tables. It reduces the time to recover from accidental delete or write operations. DynamoDB tables that have PITR enabled can be restored to any point in time in the last 35 days.

## Remediation

To remediate this issue, add point-in-time recovery to your DynamoDB table.

**To enable DynamoDB point-in-time recovery for an existing table**

1. Open the [DynamoDB console](https://console.aws.amazon.com/dynamodb/).
2. Choose the table that you want to work with, and then choose `Backups`.
3. In the `Point-in-time Recovery` section, under `Status`, choose `Enable`.
4. Choose `Enable` again to confirm the change.