## Description

This control checks whether an EC2 instance uses multiple Elastic Network Interfaces (ENIs) or Elastic Fabric Adapters (EFAs). This control passes if a single network adapter is used. The control includes an optional parameter list to identify the allowed ENIs.

Multiple ENIs can cause dual-homed instances, meaning instances that have multiple subnets. This can add network security complexity and introduce unintended network paths and access.

## Remediation

To remediate this issue, detach the additional ENIs.

**To detach a network interface**

1. Open the [Amazon EC2 console](https://console.aws.amazon.com/ec2/).
2. Under `Network & Security`, choose `Network Interfaces`.
3. Filter the list by the noncompliant instance IDs to see the associated ENIs.
4. Select the ENIs that you want to remove.
5. From the `Actions` menu, choose `Detach`.
6. If you see the prompt `Are you sure that you want to detach the following network interface?`,      choose `Detach`.