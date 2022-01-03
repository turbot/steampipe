## Description

This control checks that none of your IAM users have policies attached. IAM users must inherit permissions from IAM groups or roles.

It does not check whether least privileged policies are applied to IAM roles and groups.

## Remediation

To resolve this issue, do the following:

1. Create an IAM group
2. Assign the policy to the group
3. Add the users to the group

The policy is applied to each user in the group.

**To create an IAM group**

1. Open the [IAM console](https://console.aws.amazon.com/iam/).
2. Choose **Groups** and then choose **Create New Group**.
3. Enter a name for the group to create and then choose **Next Step**.
4. Select each policy to assign to the group and then choose **Next Step**.
5. The policies that you choose should include any policies currently attached directly to a user account. The next step to resolve a failed check is to add users to a group and then assign the policies to that group.
6. Each user in the group gets assigned the policies assigned to the group.
7. Confirm the details on the **Review** page and then choose **Create Group**.

**To add users to an IAM group**

1. Open the [IAM console](https://console.aws.amazon.com/iam/).
2. Choose **Groups**.
3. Choose **Group Actions** and then choose **Add Users to Group**.
4. Choose the users to add to the group and then choose **Add Users**.

**To remove a policy attached directly to a user**

1. Open the [IAM console](https://console.aws.amazon.com/iam/).
2. Choose **Users**.
3. For the user to detach a policy from, in the User name column, choose the name.
4. For each policy listed under **Attached directly**, to remove the policy from the user, choose the X on the right side of the page and then choose **Remove**.
5. Confirm that the user can still use AWS services as expected.