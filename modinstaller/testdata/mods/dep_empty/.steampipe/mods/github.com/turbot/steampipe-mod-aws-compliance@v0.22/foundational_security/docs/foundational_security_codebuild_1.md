## Description

This control checks whether the GitHub or Bitbucket source repository URL contains either personal access tokens or a user name and password.

Authentication credentials should never be stored or transmitted in clear text or appear in the repository URL. Instead of personal access tokens or user name and password, you should use OAuth to grant authorization for accessing GitHub or Bitbucket repositories. Using personal access tokens or a user name and password could expose your credentials to unintended data exposure and unauthorized access.

## Remediation

You can update your CodeBuild project to use OAuth.

**To remove basic authentication / (GitHub) Personal Access Token from CodeBuild project source**

1. Open the [CodeBuild console](https://console.aws.amazon.com/codebuild/).
2. Choose the build project that contains personal access tokens or a user name and password.
3. From `Edit`, choose `Source`.
4. Choose `Disconnect from GitHub / Bitbucket`.
5. Choose `Connect using OAuth`, then choose `Connect to GitHub / Bitbucket`.
6. When prompted, choose `authorize as appropriate`.
7. Reconfigure your repository URL and additional configuration settings, as needed.
8. Choose `Update source`.