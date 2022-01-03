## Description

This control checks whether connections to Amazon Redshift clusters are required to use encryption in transit. The check fails if the Amazon Redshift cluster parameter `require_SSL` is not set to 1.

TLS can be used to help prevent potential attackers from using person-in-the-middle or similar attacks to eavesdrop on or manipulate network traffic. Only encrypted connections over TLS should be allowed. Encrypting data in transit can affect performance. You should test your application with this feature to understand the performance profile and the impact of TLS.

## Remediation

To remediate this issue, update the parameter group to require encryption.

**To modify a parameter group**

1. Open the [Amazon Redshift console](https://console.aws.amazon.com/redshift/).
2. In the navigation menu, choose `Config`, then choose `Workload management` to display the `Workload management` page.
3. Choose the parameter group that you want to modify.
4. Choose `Parameters`.
5. Choose `Edit parameters` then set require_ssl to 1.
6. Enter your changes and then choose `Save`.