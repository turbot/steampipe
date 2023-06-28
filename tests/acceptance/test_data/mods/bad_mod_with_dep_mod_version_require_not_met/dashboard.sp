
dashboard "sample_dashboard" {
  title = "Sample dashboard"
  description = "Dashboard to test this mod(mod loading)"

  text {
    value = <<-EOT
    ## Note
    This report requires an [AWS Credential Report](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_getting-report.html) for each account.
    You can generate a credential report via the AWS CLI:
    EOT
  }

  text {
    width = 3
    value = <<-EOT
    ```bash
    aws iam generate-credential-report
    ```
    EOT
  }
}