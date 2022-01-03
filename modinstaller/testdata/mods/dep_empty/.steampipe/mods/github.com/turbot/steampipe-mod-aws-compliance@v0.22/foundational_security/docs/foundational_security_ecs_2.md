## Description

This control checks whether Amazon ECS services are configured to automatically assign public IP addresses. This control fails if AssignPublicIP is ENABLED. This control passes if AssignPublicIP is DISABLED.

A public IP address is an IP address that is reachable from the internet. If you launch your Amazon ECS instances with a public IP address, then your Amazon ECS instances are reachable from the internet. Amazon ECS services should not be publicly accessible, as this may allow unintended access to your container application servers.

## Remediation

To disable automatic public IP assignment, see [To configure VPC and security group settings for your service](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-configure-network.html).
