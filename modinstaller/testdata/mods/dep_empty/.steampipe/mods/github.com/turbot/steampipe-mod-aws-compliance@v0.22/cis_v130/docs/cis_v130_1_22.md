## Description

In multi-account environments, IAM user centralization facilitates greater user control. User access beyond the initial account is then provide via role assumption. Centralization of users can be accomplished through federation with an external identity provider or through the use of AWS Organizations.

Centralizing IAM user management to a single identity provider, reduces complexity and thus less access management errors.

## Remediation

### From Console

Perform the following action to check:

For multi-account AWS environments with an external identity provider

1. Determine the master account for identity federation or IAM user management.
2. Sign into the AWS console and open the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
3. In the left navigation pane, choose **Identity providers**.
4. Verify the configuration.

For all accounts that should not have local users present. For each account

1. Determine all accounts that should not have local users present
2. Sign into the AWS console and open the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
3. Switch role into each identified account.
4. Click **Users**.
5. Confirm that no IAM users representing individuals are present.

For multi-account AWS environments implementing AWS Organizations without an external identity provider

1. Determine all accounts that should not have local users present.
2. Sign into the AWS console and open the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
3. Switch role into each identified account.
4. Click **Users**.
5. Confirm that no IAM users representing individuals are present.

**Note**: The remediation procedure might vary based on the individual organization's implementation of identity federation and/or AWS Organizations with the acceptance criteria that no non-service IAM users, and non-root accounts, are present outside the account providing centralized IAM user management.