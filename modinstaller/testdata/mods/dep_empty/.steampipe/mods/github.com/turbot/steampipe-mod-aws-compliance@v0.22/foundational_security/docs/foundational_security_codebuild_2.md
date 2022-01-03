## Description

This control checks whether the project contains the environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

Authentication credentials `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` should never be stored in clear text, as this could lead to unintended data exposure and unauthorized access.

## Remediation

To remediate this issue, update your CodeBuild project to remove the environment variable.

**To remove environment variables from a CodeBuild project**

1. Open the [CodeBuild console](https://console.aws.amazon.com/codebuild/).
2. Expand `Build`.
3. Choose `Build project`, and then choose the build project that contains plaintext credentials.
4. From `Edit`, choose `Environment`.
5. Expand `Additional configuration`.
6. Choose `Remove` next to the environment variables.
7. Choose `Update environment`.

**To store sensitive values in the Amazon EC2 Systems Manager Parameter Store and then retrieve them from your build spec**

1. Open the [CodeBuild console](https://console.aws.amazon.com/codebuild/).
2. Expand `Build`.
3. Choose `Build project`, and then choose the build project that contains plaintext credentials.
4. From `Edit`, choose `Environment`.
5. Expand `Additional configuration` and scroll to `Environment variables`.
6. Follow this [tutorial](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-paramstore-console.html) to create a Systems Manager parameter that contains your sensitive data.
7. After you create the parameter, copy the parameter name.
8. Back in the CodeBuild console, choose `Create environmental variable`.
9. Enter the name of your variable as it appears in your build spec.
10. For `Value`, paste the name of your parameter.
11. For `Type`, choose `Parameter`.
12. To remove your noncompliant environmental variable that contains plaintext credentials, choose `Remove`.
13. Choose `Update environment`.