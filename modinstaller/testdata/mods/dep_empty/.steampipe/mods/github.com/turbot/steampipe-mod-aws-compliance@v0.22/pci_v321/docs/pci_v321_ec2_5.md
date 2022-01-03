## Description

This control checks whether security groups in use disallow unrestricted incoming SSH traffic.

It does not evaluate outbound traffic.

Note that security groups are stateful. If you send a request from your instance, the response traffic for that request is allowed to flow in regardless of inbound security group rules. Responses to allowed inbound traffic are allowed to flow out regardless of outbound rules.

## Remediation

Perform the following steps for each security group associated with a VPC.

1. Open the Amazon [VPC console](https://console.aws.amazon.com/vpc/).
2. In the navigation pane, under Security, choose **Security groups**.
3. Select a `security group`.
4. In the bottom section of the page, choose `Inbound rules`.
5. Choose **Edit** `inbound rules`.
6. Identify the rule that allows access through port 22 and then choose the `X` to remove it.
7. Choose **Save** rules.
