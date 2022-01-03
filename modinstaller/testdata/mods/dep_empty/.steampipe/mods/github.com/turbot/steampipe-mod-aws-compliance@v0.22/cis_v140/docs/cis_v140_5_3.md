## Description

A VPC comes with a default security group whose initial settings deny all inbound traffic, allow all outbound traffic, and allow all traffic between instances assigned to the security group. If you don't specify a security group when you launch an instance, the instance is automatically assigned to this default security group. Security groups provide stateful filtering of ingress/egress network traffic to AWS resources. It is recommended that the default security group restrict all traffic.

Configuring all VPC default security groups to restrict all traffic will encourage least privilege security group development and mindful placement of AWS resources into security groups which will in-turn reduce the exposure of those resources.

## Remediation

Security Group Members
Perform the following to implement the prescribed state:

1. Identify AWS resources that exist within the default security group.
2. Create a set of least privilege security groups for those resources.
3. Place the resources in those security groups.
4. Remove the resources noted in #1 from the default security group.

### From Console

1. Login to the AWS [VPC Console](https://console.aws.amazon.com/vpc/home)
2. Repeat the next steps for all VPCs - including the default VPC in each AWS region:
3. In the left pane, click **Security Groups**
4. For each default security group, perform the following:
5. Select the `default` security group
6. For each default security group, choose the `Inbound rules` tab and delete all inbound rules.
7. For each default security group, choose the `Outbound rules` tab and delete all outbound rules.
8. Create a set of least-privilege security groups for the resources. See [here](https://docs.aws.amazon.com/vpc/latest/userguide/VPC_SecurityGroups.html#WorkingWithSecurityGroups) for more details.
