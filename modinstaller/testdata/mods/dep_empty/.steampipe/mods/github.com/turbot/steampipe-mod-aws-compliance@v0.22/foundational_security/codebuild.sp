locals {
  foundational_security_codebuild_common_tags = merge(local.foundational_security_common_tags, {
    service = "codebuild"
  })
}

benchmark "foundational_security_codebuild" {
  title         = "CodeBuild"
  documentation = file("./foundational_security/docs/foundational_security_codebuild.md")
  children = [
    control.foundational_security_codebuild_1,
    control.foundational_security_codebuild_2
  ]
  tags          = local.foundational_security_codebuild_common_tags
}

control "foundational_security_codebuild_1" {
  title         = "1 CodeBuild GitHub or Bitbucket source repository URLs should use OAuth"
  description   = "Authentication credentials should never be stored or transmitted in clear text or appear in the repository URL. Instead of personal access tokens or user name and password, you should use OAuth to grant authorization for accessing GitHub or Bitbucket repositories. Using personal access tokens or a user name and password could expose your credentials to unintended data exposure and unauthorized access."
  severity      = "critical"
  sql           = query.codebuild_project_source_repo_oauth_configured.sql
  documentation = file("./foundational_security/docs/foundational_security_codebuild_1.md")

  tags = merge(local.foundational_security_codebuild_common_tags, {
    foundational_security_item_id  = "codebuild_1"
    foundational_security_category = "secure_development"
  })
}

control "foundational_security_codebuild_2" {
  title         = "2 CodeBuild project environment variables should not contain clear text credentials"
  description   = "This control checks whether the project contains the environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY. Authentication credentials AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY should never be stored in clear text, as this could lead to unintended data exposure and unauthorized access."
  severity      = "critical"
  sql           = query.codebuild_project_plaintext_env_variables_no_sensitive_aws_values.sql
  documentation = file("./foundational_security/docs/foundational_security_codebuild_2.md")

  tags = merge(local.foundational_security_codebuild_common_tags, {
    foundational_security_item_id  = "codebuild_2"
    foundational_security_category = "secure_development"
  })
}