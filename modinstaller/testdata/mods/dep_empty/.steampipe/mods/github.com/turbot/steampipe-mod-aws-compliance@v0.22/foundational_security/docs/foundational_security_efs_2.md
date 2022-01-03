## Description

This control checks whether Amazon Elastic File System (Amazon EFS) file systems are added to the backup plans in AWS Backup. The control fails if Amazon EFS file systems are not included in the backup plans.

Including EFS file systems in the backup plans helps you to protect your data from deletion and data loss.

## Remediation

To remediate this issue, update your file system to enable automatic backups.

**To enable automatic backups for an existing file system**

1. Open the [Amazon Elastic File System console](https://console.aws.amazon.com/efs/).
2. On the `File systems` page, choose the file system for which to enable automatic backups. The File system details page is displayed.
3. Under `General`, choose `Edit`.
4. To enable automatic backups, select `Enable automatic backups`.
5. Choose `Save changes`.