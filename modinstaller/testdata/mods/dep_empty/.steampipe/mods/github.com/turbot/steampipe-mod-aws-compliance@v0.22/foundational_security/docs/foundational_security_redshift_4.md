## Description

This control checks whether an Amazon Redshift cluster has audit logging enabled.

Amazon Redshift audit logging provides additional information about connections and user activities in your cluster. This data can be stored and secured in Amazon S3 and can be helpful in security audits and investigations.

## Remediation

To enable cluster audit logging.

**To modify the snapshot retention period**

1. Open the [Amazon Redshift console](https://console.aws.amazon.com/redshift/).
2. In the navigation menu, choose `Clusters`, then choose the name of the cluster to modify.
3. Choose `Maintenance and monitoring`.
4. Under `Audit logging`, choose `Edit`.
5. Set `Enable audit logging` to `yes`, then enter the log destination bucket details.
6. Choose `Confirm`.