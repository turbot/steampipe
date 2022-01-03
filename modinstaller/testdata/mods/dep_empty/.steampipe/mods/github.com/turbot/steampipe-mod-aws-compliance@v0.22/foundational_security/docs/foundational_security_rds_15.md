## Description

This control checks whether high availability is enabled for your RDS DB clusters.

RDS DB clusters should be configured for multiple Availability Zones to ensure availability of the data that is stored. Deployment to multiple Availability Zones allows for automated failover in the event of an Availability Zone availability issue and during regular RDS maintenance events.

## Remediation

To remediate this control, configure your DB cluster for multiple Availability Zones.

**To enable multi-AZ for a DB cluster**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. In the navigation pane, choose `Databases`, and then choose the DB instance to modify.
3. Choose `Modify`. The `Modify DB Instance` page appears.
4. Under `Instance Specifications`, set `Multi-AZ deployment` to `Yes`.
5. Choose Continue and check the summary of modifications.
6. Choose `Continue`.
7. (Optional) Choose `Apply immediately` to apply the changes immediately. Choosing this option can cause an outage in some cases. For more information, see [Using the Apply Immediately setting](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html#USER_ModifyInstance.ApplyImmediately)

On the confirmation page, review your changes. If they are correct, choose `Modify DB Instance`.