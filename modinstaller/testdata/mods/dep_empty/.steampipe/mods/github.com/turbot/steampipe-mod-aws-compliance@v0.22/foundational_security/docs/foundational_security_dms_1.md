## Description

This control checks whether AWS DMS replication instances are public. To do this, it examines the value of the PubliclyAccessible field.

A private replication instance has a private IP address that you cannot access outside of the replication network. A replication instance should have a private IP address when the source and target databases are in the same network, and the network is connected to the replication instance's VPC using a VPN, AWS Direct Connect, or VPC peering.

You should also ensure that access to your AWS DMS instance configuration is limited to only authorized users. To do this, restrict users’ IAM permissions to modify AWS DMS settings and resources.

## Remediation

**Note** that you cannot change the public access setting once a replication instance is created. It must be deleted and recreated.

To configure the AWS DMS replication instances setting to be not publicly accessible

1. Open the [AWS Database Migration Service console](https://console.aws.amazon.com/dms/).
2. In the left navigation pane, under `Resource management`, navigate to `Replication instances`.
3. To delete the public instance, select the check box for the instance, choose **Actions**, then choose **delete**.
4. Choose Create replication instance. Provide the configuration details.
5. To disable public access, make sure that Publicly accessible is not selected.
6. Choose **Create**.

For more information, see the section on [Creating a replication instance](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_ReplicationInstance.html#CHAP_ReplicationInstance.Creating) in the AWS Database Migration Service User Guide.