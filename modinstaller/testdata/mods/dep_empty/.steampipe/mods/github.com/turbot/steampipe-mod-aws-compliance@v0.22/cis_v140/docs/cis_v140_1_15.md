## Description

IAM users are granted access to services, functions, and data through IAM policies. There are multiple ways to define policies for an user, such as:
  - Add the user to an IAM group that has an attached policy.
  - Attach an inline policy directly to an user.
  - Attach a managed policy directly to an user.

Only the first implementation is recommended.

Assigning IAM policy only through groups simplifies permissions management to a single, flexible layer consistent with organizational functional roles. By simplifying permissions management, the likelihood of excessive permissions is reduced.

## Remediation

### From Console

Perform the following to create an IAM group and assign a list of policies to it:

1. Sign into the AWS console and open the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, click **User groups** and then click **Create group**.
3. In the `User group name` box, type the name of the group.
4. In the list of policies, select the `check box` for each policy that you want to apply to all members of the group (You can attach up to 10 policies to this user group).
5. Click **Create group**. Group is created with the list of permissions.

Perform the following to add a user to a given group:

1. Sign into the AWS console and open the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, click **User groups**.
3. Select the `Group name` to add an user to.
4. Click `Add users` to group.
5. Select the users to be added to the group.
6. Click **Add users**. Users are added to the group.

Perform the following to remove a direct association between an user and the policy:

1. Sign into the AWS console and open the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, click on **Users**.
3. For each user:
    - Select the user, it will take you to `Permissions` tab.
    - Expand Permissions policies.
    - Click `X` for each policy and then click **Remove** (depending on policy type).
