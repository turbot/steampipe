## Description

This control checks whether a service endpoint for Amazon EC2 is created for each VPC. The control fails if a VPC does not have a VPC endpoint created for the Amazon EC2 service.

To improve the security posture of your VPC, you can configure Amazon EC2 to use an interface VPC endpoint. Interface endpoints are powered by AWS PrivateLink, a technology that enables you to access Amazon EC2 API operations privately. It restricts all network traffic between your VPC and Amazon EC2 to the Amazon network. Because endpoints are supported within the same Region only, you cannot create an endpoint between a VPC and a service in a different Region. This prevents unintended Amazon EC2 API calls to other Regions.

## Remediation

To remediate this issue, you can create an interface VPC endpoint to Amazon EC2.

**To create an interface endpoint to Amazon EC2 from the Amazon VPC console**

1. Open the [Amazon VPC console](https://console.aws.amazon.com/vpc/).
2. In the navigation pane, choose `Endpoints`.
3. Choose `Create Endpoint`.
4. For `Service category`, choose `AWS services`.
5. For `Service Name`, choose `com.amazonaws.`*region*.`ec2`.
6. For `Type`, choose `Interface`.
7. Complete the following information.
   - For `VPC`, select a VPC in which to create the endpoint.
   - For `Subnets`, select the subnets (Availability Zones) in which to create the endpoint network interfaces. Not all Availability Zones are supported for all AWS services.
   - To enable private DNS for the interface endpoint, select the check box for `Enable DNS Name`. This option is enabled by default.
   - To use the private DNS option, the following attributes of your VPC must be set to true:
     - `enableDnsHostnames`
     - `enableDnsSupport`
     - For more information, see [Viewing and updating DNS support for your VPC](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-dns.html#vpc-dns-updating) in the Amazon VPC User Guide.
   - For `Security group`, select the security groups to associate with the endpoint network interfaces.
   - (Optional) Add or remove a tag. To add a tag, choose `Add tag` and do the following:
     - For `Key`, enter the tag name.
     - For `Value`, enter the tag value.
   - To remove a tag, choose the delete button (x) to the right of the tag `Key` and `Value`.
8. Choose `Create endpoint`.

**To create an interface VPC endpoint policy**

You can attach a policy to your VPC endpoint to control access to the Amazon EC2 API. The policy specifies the following:

  - The principal that can perform actions
  - The actions that can be performed
  - The resource on which the actions can be performed