## Description

This control checks whether the RDS cluster or instance uses a port other than the default port of the database engine.

If you use a known port to deploy an RDS cluster or instance, an attacker can guess information about the cluster or instance. The attacker can use this information in conjunction with other information to connect to an RDS cluster or instance or gain additional information about your application.

When you change the port, you must also update the existing connection strings that were used to connect to the old port. You should also check the security group of the DB instance to ensure that it includes an ingress rule that allows connectivity on the new port.

## Remediation

**To modify the default port of an existing DB instance**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/.)

2. Choose `Databases`.

3. Select the DB instance to modify

4. Choose `Modify`.

5. Under `Database options`, change `Database port` to a non-default value.

6. Choose `Continue`.

7. Under `Scheduling of modifications`, choose when to apply modifications. You can choose either `Apply during the next scheduled maintenance window` or `Apply immediately`.

8. For clusters, choose `Modify cluster`. For instances, choose `Modify DB Instance`.
