## Description

This control checks that the default security group of a VPC does not allow inbound or outbound traffic.

It does not check for access restrictions for other security groups that are not default, and other VPC configurations.

## Remediation

To remediate this issue, create new security groups and assign those security groups to your resources. To prevent the default security groups from being used, remove their inbound and outbound rules.

1. Open the [Amazon VPC console](https://console.aws.amazon.com/vpc/).
2. In the navigation pane, choose `Security groups`. View the `default security groups` details to see the resources that are assigned to them.
3. Select a default security group, and choose the **Inbound rules** tab. Choose **Edit** inbound rules. Then delete all of the inbound rules. Choose **Save rules**.
4. Repeat the previous step for each default security group.
5. Select a default security group and choose the **Outbound rules** tab. Choose **Edit** outbound rules. Then delete all of the outbound rules. Choose **Save rules**.
6. Repeat the previous step for each default security group.

Create a set of least-privilege security groups for the resources. For details on how to create security groups, see [Creating a security group](https://docs.aws.amazon.com/vpc/latest/userguide/VPC_SecurityGroups.html#CreatingSecurityGroups) in the Amazon VPC User Guide.