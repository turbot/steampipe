## Description

This control checks whether HTTP to HTTPS redirection is configured on all HTTP listeners of Application Load Balancers. The control fails if any of the HTTP listeners of Application Load Balancers do not have HTTP to HTTPS redirection configured.

Before you start to use your Application Load Balancer, you must add one or more listeners. A listener is a process that uses the configured protocol and port to check for connection requests. Listeners support both the HTTP and HTTPS protocols. You can use an HTTPS listener to offload the work of encryption and decryption to your load balancer. To enforce encryption in transit, you should use redirect actions with Application Load Balancers to redirect client HTTP requests to an HTTPS request on port 443.

## Remediation

To enable VPC flow logging

1. Open the [Amazon EC2 console](https://console.aws.amazon.com/ec2/).
2. In the navigation pane, under Load Balancing, choose **Load balancers**.
3. Choose an `Application Load Balancer`.
4. Choose **Listeners**.
5. Select the check box for an HTTP listener (port 80 TCP) and then choose **Edit**.
6. If there is an existing rule, you must delete it. Otherwise, choose **Add action** and then choose **Redirect to....**
7. Choose `HTTPS` and then enter `443`.
8. Choose the check mark in a circle symbol and then choose **Update**.