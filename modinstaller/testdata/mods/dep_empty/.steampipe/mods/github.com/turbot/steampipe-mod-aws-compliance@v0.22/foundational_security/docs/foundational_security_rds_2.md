## Description

This control checks whether RDS instances are publicly accessible by evaluating the publiclyAccessible field in the instance configuration item. The value of publiclyAccessible indicates whether the DB instance is publicly accessible. When the DB instance is publicly accessible, it is an Internet-facing instance with a publicly resolvable DNS name, which resolves to a public IP address. When the DB instance isn't publicly accessible, it is an internal instance with a DNS name that resolves to a private IP address.

The control does not check VPC subnet routing settings or the Security Group rules. You should also ensure VPC subnet routing does not allow public access, and that the security group inbound rule associated with the RDS instance does not allow unrestricted access (0.0.0.0/0). You should also ensure that access to your RDS instance configuration is limited to only authorized users by restricting users' IAM permissions to modify RDS instances settings and resources.

## Remediation

To remove public access for Amazon RDS Databases

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. Navigate to Databases and then choose your public database.
3. Choose **Modify**.
4. Scroll to **Network & Security**.
5. For `Public accessibility`, choose **No**.
6. Scroll to the bottom and then choose **Continue**.
7. Under Scheduling of modifications, choose **Apply immediately**.
8. Choose Modify DB Instance.