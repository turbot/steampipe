## Description

This control checks whether your Auto Scaling groups that are associated with a load balancer are using Elastic Load Balancing health checks.

PCI DSS does not require load balancing or highly available configurations. However, this check aligns with AWS best practices.

## Remediation

To enable Elastic Load Balancing health checks

1. Open the Amazon [EC2 console](https://console.aws.amazon.com/ec2/)
2. On the navigation pane, under `Auto Scaling`, choose **Auto Scaling Groups**
3. To select the group from the list, choose the right box
4. Choose **Edit**
5. For `Health Check Type`, choose **ELB**
6. For `Health Check Grace Period`, enter `300`
7. Choose **Save**