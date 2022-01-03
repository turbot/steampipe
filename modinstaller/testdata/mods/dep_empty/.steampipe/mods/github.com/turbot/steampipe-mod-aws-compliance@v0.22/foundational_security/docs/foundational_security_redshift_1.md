## Description

This control checks whether Amazon Redshift clusters are publicly accessible by evaluating the publiclyAccessible field in the cluster configuration item.

## Remediation

1. Open the [Amazon Redshift console](https://console.aws.amazon.com/redshift/).
2. On the navigation pane, choose **Clusters** and then select your public Amazon Redshift cluster.
3. From the Cluster drop-down menu, choose **Modify cluster**.
4. In `Publicly accessible`, choose **No**.
5. Choose **Modify**.