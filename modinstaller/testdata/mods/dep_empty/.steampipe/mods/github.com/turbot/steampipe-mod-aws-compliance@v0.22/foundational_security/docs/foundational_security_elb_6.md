## Description

This control checks whether an Application Load Balancer has deletion protection enabled. The control fails if deletion protection is not configured.

Enable deletion protection to protect your Application Load Balancer from deletion.

## Remediation

To prevent your load balancer from being deleted accidentally, you can enable deletion protection. By default, deletion protection is disabled for your load balancer.

If you enable deletion protection for your load balancer, you must disable delete protection before you can delete the load balancer.

**To enable deletion protection from the console**

1. Open the [Amazon EC2 console](https://console.aws.amazon.com/ec2/).
2. On the navigation pane, under `LOAD BALANCING`, choose `Load Balancers`.
3. Choose the load balancer.
4. On the `Description` tab, choose `Edit attributes`.
5. On the `Edit load balancer attributes` page, select `Enable for Delete Protection`, and then choose `Save`.
6. Choose `Save`.