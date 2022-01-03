## Description

This control checks whether the GitHub or Bitbucket source repository URL contains either personal access tokens or a user name and password.

You can use CodeBuild in your PCI DSS environment to compile your source code, run unit tests, or produce artifacts that are ready to deploy. If you do, your authentication credentials should never be stored or transmitted in clear text or appear in the repository URL.

You should use OAuth instead of personal access tokens or a user name and password to grant authorization for accessing GitHub or Bitbucket repositories. This is a method to use strong cryptography to render authentication credentials unreadable.

## Remediation

To remove basic authentication / (GitHub) Personal Access Token from CodeBuild Project Source

1. Open the [CodeBuild console](<https://console.aws.amazon.com/codebuild/>)
2. Select your Build project that contains personal access tokens or a user name and password.
3. From **Edit**, choose Source.
4. Choose **Disconnect** from `GitHub / Bitbucket`.
5. Choose **Connect** using `OAuth` and then choose Connect to `GitHub / Bitbucket`.
6. In the message displayed by your `source provider`, authorize as appropriate.
7. Reconfigure your Repository URL and additional configuration settings, as needed.
8. Choose **Update** source.
