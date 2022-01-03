## Description

This control helps you maintain an accurate asset inventory of needed security groups in your cardholder data environment (CDE). It does so by checking that security groups are attached to Amazon EC2 instances or to an ENI. A failed finding indicates you may have unused Amazon EC2 security groups.

Unless there is a business need to retain them, you should remove unused resources to maintain an accurate inventory of system components.

## Remediation

You must perform the following steps for each security group not attached to an ENI.

1. Open the Amazon [VPC console](https://console.aws.amazon.com/vpc/)
2. In the navigation pane, under Security, choose **Security groups**.
3. Select the check box for the security group to **delete**.
4. From **Actions**, choose **Delete security group**.
5. Choose **Delete**.
