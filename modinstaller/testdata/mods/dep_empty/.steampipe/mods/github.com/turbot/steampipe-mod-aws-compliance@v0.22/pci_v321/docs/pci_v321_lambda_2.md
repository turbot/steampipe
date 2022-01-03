## Description

This control checks whether a Lambda function is in a VPC.

It does not evaluate the VPC subnet routing configuration to determine public reachability.

Note that if Lambda@Edge is found in the account, then this control generates failed findings. To prevent these findings, you can disable this control.

## Remediation

To configure a function to connect to private subnets in a virtual private cloud (VPC) in your account

1. Open the [AWS Lambda console](https://console.aws.amazon.com/lambda/).
2. Navigate to `Functions` and then select your Lambda function.
3. Scroll to **Network** and then select a **VPC** with the connectivity requirements of the function.
4. To run your functions in high availability mode, Security Hub recommends that you choose at least two subnets.
5. Choose at least one security group that has the connectivity requirements of the function.
6. Choose **Save**.
