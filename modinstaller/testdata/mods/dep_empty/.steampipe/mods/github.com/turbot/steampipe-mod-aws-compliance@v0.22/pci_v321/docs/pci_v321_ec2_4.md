## Description

This control checks whether Elastic IP addresses that are allocated to a VPC are attached to Amazon EC2 instances or in-use elastic network interfaces (ENIs).

A failed finding indicates you may have unused Amazon EC2 EIPs.

This will help you maintain an accurate asset inventory of EIPs in your cardholder data environment (CDE). Unless there is a business need to retain them, you should remove unused resources to maintain an accurate inventory of system components.

## Remediation

To remediate this issue, create new security groups and assign those security groups to your resources. To prevent the default security groups from being used, remove their inbound and outbound rules.

1. Open the [Amazon EC2 console](https://console.aws.amazon.com/ec2/).
2. In the navigation pane, under Network & Security, choose Elastic IPs.
3. Choose the Elastic IP address, choose **Actions**, and then choose **Release Elastic IP address**.
4. When prompted, choose **Release**.
