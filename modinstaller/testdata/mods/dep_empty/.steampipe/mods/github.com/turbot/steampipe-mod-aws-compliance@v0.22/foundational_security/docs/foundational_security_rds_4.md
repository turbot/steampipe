## Description

This control checks whether RDS DB snapshots are encrypted.

This control is intended for RDS DB instances. However, it can also generate findings for snapshots of Aurora DB instances, Neptune DB instances, and Amazon DocumentDB clusters. If these findings are not useful, then you can suppress them.

Encrypting data at rest reduces the risk that an unauthenticated user gets access to data that is stored on disk. Data in RDS snapshots should be encrypted at rest for an added layer of security.

## Remediation

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. In the navigation pane, choose `Snapshots`.
3. Find the snapshot to encrypt under `Manual` or `System`.
4. Select the check box next to the snapshot to encrypt.
5. Choose `Actions`, then choose `Copy Snapshot`.
6. Under `New DB Snapshot Identifier`, type a name for the new snapshot.
7. Under `Encryption`, select `Enable Encryption`.
8. Choose the KMS key to use to encrypt the snapshot.
9. Choose `Copy Snapshot`.
10. After the new snapshot is created, delete the original snapshot.
11. For `Backup Retention Period`, choose a positive nonzero value. For example, 30 days.