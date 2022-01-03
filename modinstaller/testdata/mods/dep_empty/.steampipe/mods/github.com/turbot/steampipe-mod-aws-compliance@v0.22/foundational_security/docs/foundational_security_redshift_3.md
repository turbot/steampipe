## Description

This control checks whether Amazon Redshift clusters have automated snapshots enabled. It also checks whether the snapshot retention period is greater than or equal to seven.

Backups help you to recover more quickly from a security incident. They strengthen the resilience of your systems. Amazon Redshift takes periodic snapshots by default. This control checks whether automatic snapshots are enabled and retained for at least seven days. 

## Remediation

To remediate this issue, update the snapshot retention period to at least 7.

**To modify the snapshot retention period**

1. Open the [Amazon Redshift console](https://console.aws.amazon.com/redshift/).
2. In the navigation menu, choose `Clusters`, then choose the name of the cluster to modify.
3. Choose `Edit`.
4. Under `Backup`, set `Snapshot retention` to a value of 7 or greater.
5. Choose `Modify Cluster`.