## Description

AWS provides a support center that can be used for incident notification and response, as well as technical support and customer services. Create an IAM Role to allow authorized users to manage incidents with AWS Support.

By implementing least privilege for access control, an IAM Role will require an appropriate IAM Policy to allow Support Center Access in order to manage Incidents with AWS Support.

All AWS Support plans include an unlimited number of account and billing support cases, with no long-term contracts. Support billing calculations are performed on a per-account basis for all plans. Enterprise Support plan customers have the option to include multiple enabled accounts in an aggregated monthly billing calculation. Monthly charges for the Business and Enterprise support plans are based on each month's AWS usage charges, subject to a monthly minimum, billed in advance.

## Remediation

### From Console

Perform the following action to attach 'AWSSupportAccess' managed policy to the created IAM role :

1. Sign into the AWS console and open the [IAM Dashboard](https://console.aws.amazon.com/iam/home#/home).
2. In the left navigation pane, click **Roles** and then choose **Create Role**.
3. For Role type, choose the Another AWS account.
4. For Account ID, enter the AWS account ID of the AWS account to which you want to grant access to your resources.
5. Choose **Next: Permissions**.
6. Search for the managed policy `AWSSupportAccess`.
7. Select the check box for the `AWSSupportAccess` managed policy.
8. Choose **Next: Tags**.
9. Choose **Next: Review**.
10. For `Role name`, enter a name for your role. Then click **Create role**.

You can attach the above role to any user you want that is needed.

### From Command Line

1. Create a IAM policy for managing incidents with AWS.
    - Create a trust relationship policy document that allows <iam_user> to manage AWS incidents, and save it locally as /tmp/TrustPolicy.json.
      ```json
      {
        "Version":"2012-10-17",
        "Statement":[
          {
            "Effect":"Allow",
            "Principal":{
              "AWS":"<iam_user>"
            },
            "Action":"sts:AssumeRole"
          }
        ]
      }
      ```
2. Create the IAM role using the above trust policy.
```bash
aws iam create-role --role-name <aws_support_iam_role> --assume-role-policy- document file:///tmp/TrustPolicy.json
```
3. Attach 'AWSSupportAccess' managed policy to the created IAM role.
```bash
aws iam attach-role-policy --policy-arn arn:aws:iam::aws:policy/AWSSupportAccess --role-name <aws_support_iam_role>
```