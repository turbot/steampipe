## Description

This control checks whether your EC2 instance metadata version is configured with Instance Metadata Service Version 2 (IMDSv2). The control passes if `HttpTokens` is set to `required for IMDSv2`. The control fails if `HttpTokens` is set to `optional`.

You use instance metadata to configure or manage the running instance. The IMDS provides access to temporary, frequently rotated credentials. These credentials remove the need to hard code or distribute sensitive credentials to instances manually or programmatically. The IMDS is attached locally to every EC2 instance. It runs on a special "link local" IP address of 169.254.169.254. This IP address is only accessible by software that runs on the instance.

Version 2 of the IMDS adds new protections for the following types of vulnerabilities. These vulnerabilities could be used to try to access the IMDS.

- Open website application firewalls
- Open reverse proxies
- Server-side request forgery (SSRF) vulnerabilities
- Open Layer 3 firewalls and network address translation (NAT)

Security Hub recommends that you configure your EC2 instances with IMDSv2.

## Remediation

To remediate an EC2 instance that is not configured with IMDSv2, you can require the use of IMDSv2.

To require IMDSv2 on an existing instance, when you request instance metadata, modify the Amazon EC2 metadata options. Follow the instructions in [Configuring instance metadata options for existing instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html) in the Amazon EC2 User Guide for Linux Instances.

To require the use of IMDSv2 on a new instance when you launch it, follow the instructions in [Configuring instance metadata options for new instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-instance-metadata-service.html) in the Amazon EC2 User Guide for Linux Instances.

**To configure your new EC2 instance with IMDSv2 from the console**

1. Open the [Amazon EC2 console](https://console.aws.amazon.com/ec2/).
2. Choose `Launch instance` and then choose `Launch instance`.
3. In the `Configure Instance Details` step, under `Advanced Details`, for `Metadata version`, choose `V2 (token required)`.
4. Choose `Review and Launch`.