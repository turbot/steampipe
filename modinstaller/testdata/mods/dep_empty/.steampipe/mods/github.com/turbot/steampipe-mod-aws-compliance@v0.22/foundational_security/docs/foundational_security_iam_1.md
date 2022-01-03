## Description

This control checks whether the default version of AWS Identity and Access Management policies (also known as customer managed policies) do not have administrator access with a statement that has `"Effect"`: "Allow" with "Action": "*" over "Resource": "*".

It only checks for the customer managed policies that you created, but does not check for full access to individual services, such as "S3:*".

It does not check for inline and AWS managed policies.

## Remediation

1. Open the [IAM console](https://console.aws.amazon.com/iam/).
2. Choose **Policies**.
3. Choose the radio button next to the policy to remove.
4. From **Policy actions**, choose **Detach**.
5. On the **Detach policy** page, choose the radio button next to each user to detach the policy from and then choose **Detach policy**.
6. Confirm that the user that you detached the policy from can still access AWS services and resources as expected.