## Description

This control checks whether CloudTrail trails are configured to send logs to CloudWatch Logs.

It does not check for user permissions to alter logs or log groups. You should create specific CloudWatch rules to alert when CloudTrail logs are altered.

This control also does not check for any additional audit log sources other than CloudTrail being sent to a CloudWatch Logs group.

## Remediation

To enable CloudTrail log file validation

1. Open the CloudTrail console at [CloudTrail](https://console.aws.amazon.com/cloudtrail/).
1. In the navigation pane, choose **Trails**.
1. Choose a trail that there is no value for in the **CloudWatch Logs Log group** column.
1. Scroll down to the **CloudWatch Logs** section and then choose **Edit**.
1. For Log group field, do one of the following:
    - To use the default log group, keep the name as is.
    - To use an existing log group, choose **Existing** and then enter the name of the log group to use.
    - To create a new log group, choose **New** and then enter a name for the log group to create.
1. Choose **Continue**.
1. For IAM role, do one of the following:
    - To use an existing role, choose **Existing** and then choose the role from the drop-down list.
    - To create a new role, choose **New** and then enter a name for the role to create.
    - The new role is assigned a policy that grants the necessary permissions.
    To view the permissions granted to the role, expand the **Policy document**.
1. Choose **Save** changes.