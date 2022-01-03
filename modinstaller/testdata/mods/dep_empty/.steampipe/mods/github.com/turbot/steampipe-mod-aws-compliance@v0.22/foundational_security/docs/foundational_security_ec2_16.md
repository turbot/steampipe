## Description

This control checks whether there are any unused network access control lists (ACLs).

The control checks the item configuration of the resource AWS::EC2::NetworkAcl and determines the relationships of the network ACL.

If the only relationship is the VPC of the network ACL, then the control fails.

## Remediation

For instructions on how to delete an unused network ACL, see [Delete a network ACL](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-network-acls.html#DeleteNetworkACL).