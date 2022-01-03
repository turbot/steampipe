## Description

IAM policies are the means by which privileges are granted to users, groups, or roles. It is recommended and considered a standard security practice to grant least privilege that is, granting only the permissions required to perform a task. Determine what users need to do what and then accordingly create policies for them instead of allowing full administrative privileges.

It's more secure to start with a minimum set of permissions and grant additional permissions as necessary, rather than starting with permissions that are too lenient and then trying to tighten them later.

Providing full administrative privileges instead of restricting to the minimum set of permissions that the user is required to do exposes the resources to potentially unwanted actions.

IAM policies that have a statement with `"Effect"`: `"Allow"` with `"Action"`: `"*"` over `"Resource"`: `"*"` should be removed.

## Remediation

### From Console:

Perform the following action to detach the policy that has full administrative privileges:

1. Sign into the AWS console and open the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, click **Policies** and then search for the policy name having administrative privileges.
3. Select the policy that needs to be detached. Go to **Policy usage** tab.
4. Select all `Users`, `Groups`, `Roles` that have this policy attached.
5. Click **Detach**. It will ask for re-confirmation.
6. Click **Detach** again.

Repeat the above steps for all the policies having administrative privileges.

### From Command Line:

Perform the following action to detach the policy that has full administrative privileges:

1.  Lists all IAM users, groups, and roles that the specified managed policy is attached to.
```bash
aws iam list-entities-for-policy --policy-arn <policy_arn>
```
2. Detach the policy from all IAM Users:
```bash
aws iam detach-user-policy --user-name <iam_user> --policy-arn <policy_arn>
```
3. Detach the policy from all IAM Groups:
```bash
aws iam detach-group-policy --group-name <iam_group> --policy-arn <policy_arn>
```
4. Detach the policy from all IAM Roles:
```bash
aws iam detach-role-policy --role-name <iam_role> --policy-arn <policy_arn>
```