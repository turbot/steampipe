## Description

This control checks whether automatic major version upgrades are enabled for the Amazon Redshift cluster.

Enabling automatic major version upgrades ensures that the latest major version updates to Amazon Redshift clusters are installed during the maintenance window. These updates might include security patches and bug fixes. Keeping up to date with patch installation is an important step in securing systems.

## Remediation

To remediate this issue from the AWS CLI, use the Amazon Redshift modify-cluster command to set the `--allow-version-upgrade attribute`.

From the AWS CLI, run

```bash
aws redshift modify-cluster --cluster-identifier clustername --allow-version-upgrade
```

Where `clustername` is the name of your Amazon Redshift cluster.