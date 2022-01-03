benchmark "nist_800_53_rev_4_ac" {
  title       = "Access Control (AC)"
  description = "The access control family consists of security requirements detailing system logging. This includes who has access to what assets and reporting capabilities like account management, system privileges, and remote access logging to determine when users have access to the system and their level of access."
  children = [
    benchmark.nist_800_53_rev_4_ac_2,
    benchmark.nist_800_53_rev_4_ac_3,
    benchmark.nist_800_53_rev_4_ac_4,
    benchmark.nist_800_53_rev_4_ac_5,
    benchmark.nist_800_53_rev_4_ac_6,
    benchmark.nist_800_53_rev_4_ac_17,
    benchmark.nist_800_53_rev_4_ac_21
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_2" {
  title       = "Account Management (AC-2)"
  description = "Manage system accounts, group memberships, privileges, workflow, notifications, deactivations, and authorizations."
  children = [
    benchmark.nist_800_53_rev_4_ac_2_1,
    benchmark.nist_800_53_rev_4_ac_2_3,
    benchmark.nist_800_53_rev_4_ac_2_4,
    benchmark.nist_800_53_rev_4_ac_2_12,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.cloudtrail_trail_integrated_with_logs,
    control.emr_cluster_kerberos_enabled,
    control.guardduty_enabled,
    control.iam_account_password_policy_strong_min_reuse_24,
    control.iam_group_not_empty,
    control.iam_policy_no_star_star,
    control.iam_root_user_mfa_enabled,
    control.iam_root_user_no_access_keys,
    control.iam_user_access_key_age_90,
    control.iam_user_in_group,
    control.iam_user_no_inline_attached_policies,
    control.iam_user_unused_credentials_90,
    control.rds_db_instance_logging_enabled,
    control.redshift_cluster_encryption_logging_enabled,
    control.s3_bucket_logging_enabled,
    control.secretsmanager_secret_rotated_as_scheduled,
    control.securityhub_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_2_1" {
  title       = "AC-2(1) Automated System Account Management"
  description = "The organization employs automated mechanisms to support the management of information system accounts."
  children = [
    control.guardduty_enabled,
    control.iam_account_password_policy_strong_min_reuse_24,
    control.iam_user_access_key_age_90,
    control.iam_user_in_group,
    control.iam_user_unused_credentials_90,
    control.secretsmanager_secret_rotated_as_scheduled,
    control.securityhub_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_2_3" {
  title       = "AC-2(3) Disable Inactive Accounts"
  description = "The information system automatically disables inactive accounts after 90 days for user accounts."
  children = [
    control.iam_user_unused_credentials_90
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_2_4" {
  title       = "AC-2(4) Automated Audit Actions"
  description = "The information system automatically audits account creation, modification, enabling, disabling, and removal actions, and notifies [Assignment: organization-defined personnel or roles]."
  children = [
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_trail_enabled,
    control.cloudtrail_trail_integrated_with_logs,
    control.cloudwatch_alarm_action_enabled,
    control.guardduty_enabled,
    control.rds_db_instance_logging_enabled,
    control.redshift_cluster_encryption_logging_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_2_12" {
  title       = "AC-2(12) Account Monitoring"
  description = "Monitors and reports atypical usage of information system accounts to organization-defined personnel or roles."
  children = [
    control.guardduty_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_3" {
  title       = "Access Enforcement (AC-3)"
  description = "The information system enforces approved authorizations for logical access to information and system resources in accordance with applicable access control policies."
  children = [
    control.dms_replication_instance_not_publicly_accessible,
    control.ebs_snapshot_not_publicly_restorable,
    control.emr_cluster_kerberos_enabled,
    control.iam_group_not_empty,
    control.iam_policy_no_star_star,
    control.iam_root_user_no_access_keys,
    control.iam_user_in_group,
    control.iam_user_no_inline_attached_policies,
    control.iam_user_unused_credentials_90,
    control.lambda_function_restrict_public_access,
    control.rds_db_snapshot_prohibit_public_access,
    control.redshift_cluster_prohibit_public_access,
    control.sagemaker_notebook_instance_direct_internet_access_disabled,
    control.s3_bucket_restrict_public_read_access,
    control.s3_bucket_restrict_public_write_access,
    control.s3_public_access_block_bucket_account
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_4" {
  title       = "Information Flow Enforcement (AC-4)"
  description = "The information system enforces approved authorizations for controlling the flow of information within the system and between interconnected systems based on organization-defined information flow control policies."
  children = [
    control.acm_certificate_expires_30_days,
    control.dms_replication_instance_not_publicly_accessible,
    control.ebs_snapshot_not_publicly_restorable,
    control.ec2_instance_in_vpc,
    control.ec2_instance_not_publicly_accessible,
    control.emr_cluster_master_nodes_no_public_ip,
    control.es_domain_in_vpc,
    control.lambda_function_in_vpc,
    control.lambda_function_restrict_public_access,
    control.rds_db_instance_prohibit_public_access,
    control.rds_db_snapshot_prohibit_public_access,
    control.redshift_cluster_prohibit_public_access,
    control.s3_bucket_restrict_public_read_access,
    control.s3_bucket_restrict_public_write_access,
    control.s3_public_access_block_bucket_account,
    control.sagemaker_notebook_instance_direct_internet_access_disabled,
    control.vpc_default_security_group_restricts_all_traffic,
    control.vpc_igw_attached_to_authorized_vpc,
    control.vpc_security_group_restrict_ingress_common_ports_all,
    control.vpc_security_group_restrict_ingress_ssh_all,
    control.vpc_security_group_restrict_ingress_tcp_udp_all
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_5" {
  title       = "Separation Of Duties (AC-5)"
  description = "Separate duties of individuals to prevent malevolent activity. automate separation of duties and access authorizations."
  children = [
    control.emr_cluster_kerberos_enabled,
    control.iam_group_not_empty,
    control.iam_policy_no_star_star,
    control.iam_user_no_inline_attached_policies
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_6" {
  title       = "Least Privilege (AC-6)"
  description = "The organization employs the principle of least privilege, allowing only authorized accesses for users (or processes acting on behalf of users) which are necessary to accomplish assigned tasks in accordance with organizational missions and business functions."
  children = [
    control.codebuild_project_plaintext_env_variables_no_sensitive_aws_values,
    control.dms_replication_instance_not_publicly_accessible,
    control.ebs_snapshot_not_publicly_restorable,
    control.ec2_instance_not_publicly_accessible,
    control.ec2_instance_uses_imdsv2,
    control.emr_cluster_kerberos_enabled,
    control.iam_group_not_empty,
    control.iam_group_user_role_no_inline_policies,
    control.iam_policy_no_star_star,
    control.iam_root_user_no_access_keys,
    control.iam_user_in_group,
    control.iam_user_no_inline_attached_policies,
    control.iam_user_unused_credentials_90,
    control.lambda_function_restrict_public_access,
    control.rds_db_instance_prohibit_public_access,
    control.rds_db_snapshot_prohibit_public_access,
    control.redshift_cluster_prohibit_public_access,
    control.s3_bucket_restrict_public_read_access,
    control.s3_bucket_restrict_public_write_access,
    control.s3_public_access_block_bucket_account,
    control.sagemaker_notebook_instance_direct_internet_access_disabled,
    benchmark.nist_800_53_rev_4_ac_6_10
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_6_10" {
  title       = "AC-6(10) Prohibit Non-Privileged Users From Executing Privileged Functions"
  description = "The information system prevents non-privileged users from executing privileged functions to include disabling, circumventing, or altering implemented security safeguards/countermeasures."
  children = [
    control.iam_root_user_no_access_keys
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_17" {
  title       = "Remote Access (AC-17)"
  description = "Authorize remote access systems prior to connection. Enforce remote connection requirements to information systems."
  children = [
    benchmark.nist_800_53_rev_4_ac_17_1,
    benchmark.nist_800_53_rev_4_ac_17_2,
    benchmark.nist_800_53_rev_4_ac_17_3
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_17_1" {
  title       = "AC-17(1) Automated Monitoring/Control"
  description = "The information system monitors and controls remote access methods."
  children = [
    control.guardduty_enabled,
    control.securityhub_enabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_17_2" {
  title       = "AC-17(2) Protection Of Confidentiality/Integrity Using Encryption"
  description = "The information system implements cryptographic mechanisms to protect the confidentiality and integrity of remote access sessions."
  children = [
    control.acm_certificate_expires_30_days,
    control.elb_application_lb_drop_http_headers,
    control.elb_application_lb_redirect_http_request_to_https,
    control.elb_classic_lb_use_ssl_certificate,
    control.elb_classic_lb_use_tls_https_listeners,
    control.redshift_cluster_encryption_in_transit_enabled,
    control.s3_bucket_enforces_ssl
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_17_3" {
  title       = "AC-17(3) Managed Access Control Points"
  description = "The information system routes all remote accesses through organization-defined managed network access control points."
  children = [
    control.vpc_igw_attached_to_authorized_vpc
  ]

  tags = local.nist_800_53_rev_4_common_tags
}

benchmark "nist_800_53_rev_4_ac_21" {
  title       = "Information Sharing (AC-21)"
  description = "Facilitate information sharing. Enable authorized users to grant access to partners."
  children = [
    control.dms_replication_instance_not_publicly_accessible,
    control.ebs_snapshot_not_publicly_restorable,
    control.ec2_instance_not_publicly_accessible,
    control.emr_cluster_master_nodes_no_public_ip,
    control.lambda_function_restrict_public_access,
    control.rds_db_instance_prohibit_public_access,
    control.rds_db_snapshot_prohibit_public_access,
    control.redshift_cluster_prohibit_public_access,
    control.s3_bucket_restrict_public_read_access,
    control.s3_bucket_restrict_public_write_access,
    control.s3_public_access_block_bucket_account,
    control.sagemaker_notebook_instance_direct_internet_access_disabled
  ]

  tags = local.nist_800_53_rev_4_common_tags
}
