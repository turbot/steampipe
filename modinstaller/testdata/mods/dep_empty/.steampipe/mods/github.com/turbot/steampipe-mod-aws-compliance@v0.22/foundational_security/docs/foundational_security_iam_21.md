## Description

This control checks whether the IAM identity-based policies that you create have Allow statements that use the * wildcard to grant permissions for all actions on any service. The control fails if any policy statement includes "Effect": "Allow" with "Action": "Service:*".

This control only applies to customer managed IAM policies. It does not apply to IAM policies that are managed by AWS.

When you assign permissions to AWS services, it is important to scope the allowed IAM actions in your IAM policies. You should restrict IAM actions to only those actions that are needed. This helps you to provision least privilege permissions. Overly permissive policies might lead to privilege escalation if the policies are attached to an IAM principal that might not require the permission.

In some cases, you might want to allow IAM actions that have a similar prefix, such as DescribeFlowLogs and DescribeAvailabilityZones. In these authorized cases, you can add a suffixed wildcard to the common prefix. For example, ec2:Describe*.

## Remediation

### From Console:

Perform the following action to disable user console password:

1. Sign into the AWS console and navigate to the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, choose Users.
3. Select the **User name** whose `Console last sign-in` is greater than 90 days.
4. Click on **Security credentials** tab.
5. In section `Sign-in credentials`, `Console password` click **Manage**.
6. Select `Disable`, click **Apply**

Perform the following action to deactivate access keys:

1. Sign into the AWS console as an **Administrator** and navigate to the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, choose Users.
3. Click on the **User name** for which access key is over 90 days old.
4. Click on **Security credentials** tab.
5. Click on the **Make inactive** to `deactivate` the key that is over 90 days old and that have not been used.