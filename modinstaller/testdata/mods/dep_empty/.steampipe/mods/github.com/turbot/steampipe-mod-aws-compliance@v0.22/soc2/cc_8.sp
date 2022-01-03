locals {
  soc_2_cc_8_common_tags = merge(local.soc_2_common_tags, {
    soc_2_section_id = "cc8"
  })
}

benchmark "soc_2_cc_8" {
  title       = "CC8.0 - Change Management"
  description = "The criteria relevant to how an entity (i) identifies the need for changes, (ii) makes the changes using a controlled change management process, and (iii) prevents unauthorized changes from being made."

  children = [
    benchmark.soc_2_cc_8_1
  ]

  tags = local.soc_2_cc_8_common_tags
}

benchmark "soc_2_cc_8_1" {
  title         = "CC8.1 The entity authorizes, designs, develops or acquires, configures, documents, tests, approves, and implements changes to infrastructure, data, software, and procedures to meet its objectives"
  documentation = file("./soc2/docs/cc_8_1.md")

  children = [
    control.config_enabled_all_regions,
    control.codebuild_project_source_repo_oauth_configured,
    control.codebuild_project_plaintext_env_variables_no_sensitive_aws_values
  ]

  tags = merge(local.soc_2_cc_8_common_tags, {
    soc_2_item_id = "8.1"
    soc_2_type    = "automated"
  })
}