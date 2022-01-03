## Description

This control checks whether high availability is enabled for your RDS DB instances.

RDS DB instances should be configured for multiple Availability Zones (AZs). This ensures the availability of the data stored. Multi-AZ deployments allow for automated failover if there is an issue with Availability Zone availability and during regular RDS maintenance.

## Remediation

To remediate this issue, update your DB instances to enable multiple Availability Zones.

**To enable multiple Availability Zones for a DB instance**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. In the navigation pane, choose `Databases`, and then choose the DB instance that you want to modify.
3. Choose `Modify`. The `Modify DB Instance` page appears.
4. Under `Instance Specifications`, set `Multi-AZ deployment` to `Yes`.
5. Choose `Continue` and then check the summary of modifications.
6. (Optional) Choose `Apply immediately` to apply the changes immediately. Choosing this option can cause an outage in some cases. For more information, see Using the [Apply Immediately setting](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html#USER_ModifyInstance.ApplyImmediately).
7. On the confirmation page, review your changes. If they are correct, choose `Modify DB Instance` to save your changes.