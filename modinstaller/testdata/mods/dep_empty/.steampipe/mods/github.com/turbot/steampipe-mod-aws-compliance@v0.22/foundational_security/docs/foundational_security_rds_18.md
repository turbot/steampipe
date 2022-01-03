## Description

This control checks whether an RDS instance is deployed in a VPC (EC2-VPC).

VPCs provide a number of network controls to secure access to RDS resources. These controls include VPC Endpoints, network ACLs, and security groups. To take advantage of these controls, we recommend that you move EC2-Classic RDS instances to EC2-VPC.

## Remediation

For detailed instructions on how to move RDS instances to VPC, see [Updating the VPC for a DB instance](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_VPC.html#USER_VPC.VPC2VPC) in the Amazon RDS User Guide.