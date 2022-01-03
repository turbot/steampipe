## Description

This control checks whether direct internet access is disabled for an SageMaker notebook instance. To do this, it checks whether the DirectInternetAccess field is disabled for the notebook instance.

If you configure your SageMaker instance without a VPC, then by default direct internet access is enabled on your instance. You should configure your instance with a VPC and change the default setting to `Disable — Access the internet through a VPC`.

To train or host models from a notebook, you need internet access. To enable internet access, make sure that your VPC has a NAT gateway and your security group allows outbound connections. To learn more about how to connect a notebook instance to resources in a VPC, see [Connect a notebook instance to resources in a VPC](https://docs.aws.amazon.com/sagemaker/latest/dg/appendix-notebook-and-internet-access.html) in the Amazon SageMaker Developer Guide.

You should also ensure that access to your SageMaker configuration is limited to only authorized users. Restrict users' IAM permissions to modify SageMaker settings and resources.

## Remediation

Note that you cannot change the internet access setting after a notebook instance is created. It must be stopped, deleted, and recreated.

To configure an SageMaker notebook instance to deny direct internet access

1. Open the [SageMaker console](https://console.aws.amazon.com/sagemaker/)
2. Navigate to **Notebook instances**.
3. Delete the instance that has direct internet access enabled. Choose the instance, choose Actions, then choose stop.
4. After the instance is stopped, choose **Actions**, then choose **delete**.
5. Choose Create notebook instance. Provide the configuration details.
6. Expand the **Network** section. Then choose a VPC, subnet, and security group. Under **Direct internet access**, choose **Disable — Access the internet through a VPC**.
7. Choose **Create notebook instance**.