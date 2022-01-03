locals {
  pci_v321_codebuild_common_tags = merge(local.pci_v321_common_tags, {
    service = "codebuild"
  })
}

benchmark "pci_v321_codebuild" {
  title         = "CodeBuild"
  documentation = file("./pci_v321/docs/pci_v321_codebuild.md")
  children = [
    control.pci_v321_codebuild_1,
    control.pci_v321_codebuild_2
  ]
  tags          = local.pci_v321_codebuild_common_tags
}

control "pci_v321_codebuild_1" {
  title         = "1 CodeBuild GitHub or Bitbucket source repository URLs should use OAuth"
  description   = "This control checks whether the GitHub or Bitbucket source repository URL contains either personal access tokens or a user name and password."
  severity      = "critical"
  sql           = query.codebuild_project_source_repo_oauth_configured.sql
  documentation = file("./pci_v321/docs/pci_v321_codebuild_1.md")

  tags = merge(local.pci_v321_codebuild_common_tags, {
    pci_item_id      = "codebuild_1"
    pci_requirements = "8.2.1"
  })
}

control "pci_v321_codebuild_2" {
  title         = "2 CodeBuild project environment variables should not contain clear text credentials"
  description   = "This control checks whether the project contains environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY."
  severity      = "critical"
  sql           = query.codebuild_project_plaintext_env_variables_no_sensitive_aws_values.sql
  documentation = file("./pci_v321/docs/pci_v321_codebuild_2.md")

  tags = merge(local.pci_v321_codebuild_common_tags, {
    pci_item_id      = "codebuild_2"
    pci_requirements = "8.2.1"
  })
}