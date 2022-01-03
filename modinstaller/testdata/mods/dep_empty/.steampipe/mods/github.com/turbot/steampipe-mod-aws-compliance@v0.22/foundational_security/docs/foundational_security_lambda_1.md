## Description

This control checks whether the Lambda function resource-based policy prohibits public access.

It does not check for access to the Lambda function by internal principals, such as IAM roles. You should ensure that access to the Lambda function is restricted to authorized principals only by using least privilege Lambda resource-based policies.

For more information about using resource-based policies for AWS Lambda, see the [AWS Lambda Developer Guide](https://docs.aws.amazon.com/lambda/latest/dg/access-control-resource-based.html).

## Remediation

To remediate this issue, you update the resource-based policy to change the publicly accessible Lambda function to a private Lambda function.
You can only update resource-based policies for Lambda resources within the scope of the AddPermission and AddLayerVersionPermission API actions.
You cannot author policies for your Lambda resources in JSON, or use conditions that don't map to parameters for those actions using the CLI or the SDK.

**To use the AWS CLI to revoke function-use permission from an AWS service or another account**

1. To get the ID of the statement from the output of GetPolicy, from the AWS CLI, run the following:

```bash
aws lambda get-policy —function-name yourfunctionname
```
This command returns the Lambda resource-based policy string associated with the publicly accessible Lambda function.

2. From the policy statement returned by the get-policy command, copy the string value of the Sid field.

3. From the AWS CLI, run

```bash
aws lambda remove-permission --function-name yourfunctionname —statement-id youridvalue
```

To use the Lambda console to restrict access to the Lambda function

1. Open the [AWS Lambda console](https://console.aws.amazon.com/lambda/).
2. Navigate to Functions and then select your publicly accessible Lambda function.
3. Under **Designer**, choose the key icon at the top left. It has the tool-tip View permissions.
4. Under Function policy, if the policy allows actions for the principal element `“*”` or `{“AWS”: “*”}`, it is publicly accessible.
   - Consider adding the following IAM condition to scope access to your account only.

    ```json
        "Condition": {
      "StringEquals": {
        "AWS:SourceAccount": "<account_id>"
        }
      }
    }
    ```