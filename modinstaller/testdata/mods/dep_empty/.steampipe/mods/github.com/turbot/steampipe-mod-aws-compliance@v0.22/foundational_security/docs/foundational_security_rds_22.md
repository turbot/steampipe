## Description

This control checks whether an Amazon RDS event subscription exists with notifications enabled for the following source type, event category key-value pairs.

```
DBSecurityGroup: ["configuration change","failure"]
```

RDS event notifications use Amazon SNS to make you aware of changes in the availability or configuration of your RDS resources. These notifications allow for a rapid response. For additional information about RDS event notifications, see [Using Amazon RDS event notification](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Events.html) in the Amazon RDS User Guide.

## Remediation

**To subscribe to RDS database parameter group event notifications**

1. Open the [Amazon RDS console](https://console.aws.amazon.com/rds/.)

2. In the navigation pane, choose `Event subscriptions`.

3. Under `Event subscriptions`, choose `Create event subscription`.

4. In the `Create event subscription` dialog, do the following:

    a. For `Name`, enter a name for the event notification subscription.

    b. For `Send notifications to`, choose an existing Amazon SNS ARN for an SNS topic. To use a new topic, choose `create topic` to enter the name of a topic and a list of recipients.

    c. For `Source type`, choose `Security groups`.

    d. Under `Instances to include`, select `All security groups`.

    e. Under `Event categories to include`, select `Specific event categories`. The control also passes if you select `All event categories`.

    f. Select `configuration change` and `failure`.

    g. Choose `Create`.
    