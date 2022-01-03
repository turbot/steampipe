## Description

This control checks whether the assignment of public IPs in Amazon Virtual Private Cloud (Amazon VPC) subnets have MapPublicIpOnLaunch set to FALSE. The control passes if the flag is set to FALSE.

All subnets have an attribute that determines whether a network interface created in the subnet automatically receives a public IPv4 address. Instances that are launched into subnets that have this attribute enabled have a public IP address assigned to their primary network interface.

## Remediation

You can configure a subnet from the Amazon VPC console.

**To configure a subnet to not assign public IP addresses**

1. Open the [Amazon VPC console](https://console.aws.amazon.com/vpc/.)

2. In the navigation pane, choose `Subnets`.

3. Select your subnet and then choose `Subnet Actions`, `Modify auto-assign IP settings`.

4. Clear the `Enable auto-assign public IPv4 address` check box and then choose `Save`.