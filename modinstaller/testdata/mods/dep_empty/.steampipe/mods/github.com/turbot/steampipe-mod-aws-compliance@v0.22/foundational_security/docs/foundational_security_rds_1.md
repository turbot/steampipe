## Description

This control checks whether Amazon RDS snapshots are public.

This control is intended for RDS instances. It can also return findings for snapshots of Aurora DB instances, Neptune DB instances, and Amazon DocumentDB clusters, even though they are not evaluated for public accessibility. If these findings are not useful, you can suppress them.

RDS snapshots are used to back up the data on your RDS instances at a specific point in time. They can be used to restore previous states of RDS instances.

An RDS snapshot must not be public unless intended. If you share an unencrypted manual snapshot as public, this makes the snapshot available to all AWS accounts. This may result in unintended data exposure of your RDS instance.

Note that if the configuration is changed to allow public access, the AWS Config rule may not be able to detect the change for up to 12 hours. Until the AWS Config rule detects the change, the check passes even though the configuration violates the rule.

To learn more about sharing a DB snapshot, see [Sharing a DB snapshot](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ShareSnapshot.html).

## Remediation

To remediate this issue, update your RDS snapshots to remove public access.

**To remove public access for RDS snapshots**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/).
2. Navigate to `Snapshots` and then choose the public snapshot you want to modify.
3. From `Actions`, choose `Share Snapshots`.
4. From `DB snapshot visibility`, choose `Private`.
5. Under `DB snapshot visibility`, choose `all`.
6. Choose `Save`.