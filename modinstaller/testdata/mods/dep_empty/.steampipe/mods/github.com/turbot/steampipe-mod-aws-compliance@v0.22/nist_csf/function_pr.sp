benchmark "nist_csf_pr" {
  title       = "Protect (PR)"
  description = "Develop and implement appropriate safeguards to ensure delivery of critical services."

  children = [
    benchmark.nist_csf_pr_ac,
    benchmark.nist_csf_pr_ds,
    benchmark.nist_csf_pr_ip,
    benchmark.nist_csf_pr_ma,
    benchmark.nist_csf_pr_pt
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ac" {
  title       = "Identity Management and Access Control (PR.AC)"
  description = "Access to physical and logical assets and associated facilities is limited to authorized users, processes, and devices, and is managed consistent with the assessed risk of unauthorized access to authorized activities and transactions."

  children = [
    benchmark.nist_csf_pr_ac_1,
    benchmark.nist_csf_pr_ac_3,
    benchmark.nist_csf_pr_ac_4,
    benchmark.nist_csf_pr_ac_5,
    benchmark.nist_csf_pr_ac_6,
    benchmark.nist_csf_pr_ac_7
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ac_1" {
  title       = "PR.AC-1"
  description = "Identities and credentials are issued, managed, verified, revoked, and audited for authorized devices, users and processes."

  children = [
    control.iam_account_password_policy_strong_min_reuse_24,
    control.iam_group_not_empty,
    control.iam_policy_no_star_star,
    control.iam_root_user_no_access_keys,
    control.iam_user_access_key_age_90,
    control.iam_user_in_group,
    control.iam_user_no_inline_attached_policies,
    control.iam_user_unused_credentials_90,
    control.secretsmanager_secret_automatic_rotation_enabled,
    control.secretsmanager_secret_rotated_as_scheduled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ac_3" {
  title       = "PR.AC-3"
  description = "Remote access is managed."

  children = [
    control.dms_replication_instance_not_publicly_accessible,
    control.ebs_snapshot_not_publicly_restorable,
    control.ec2_instance_in_vpc,
    control.ec2_instance_not_publicly_accessible,
    control.emr_cluster_master_nodes_no_public_ip,
    control.es_domain_in_vpc,
    control.iam_root_user_hardware_mfa_enabled,
    control.iam_root_user_mfa_enabled,
    control.iam_user_console_access_mfa_enabled,
    control.iam_user_mfa_enabled,
    control.lambda_function_in_vpc,
    control.lambda_function_restrict_public_access,
    control.rds_db_instance_prohibit_public_access,
    control.rds_db_snapshot_prohibit_public_access,
    control.redshift_cluster_prohibit_public_access,
    control.s3_bucket_restrict_public_read_access,
    control.s3_bucket_restrict_public_write_access,
    control.s3_public_access_block_account,
    control.sagemaker_notebook_instance_direct_internet_access_disabled,
    control.vpc_default_security_group_restricts_all_traffic,
    control.vpc_igw_attached_to_authorized_vpc,
    control.vpc_security_group_restrict_ingress_common_ports_all,
    control.vpc_security_group_restrict_ingress_ssh_all,
    control.vpc_security_group_restrict_ingress_tcp_udp_all
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ac_4" {
  title       = "PR.AC-4"
  description = "Access permissions and authorizations are managed, incorporating the principles of least privilege and separation of duties."

  children = [
    control.emr_cluster_kerberos_enabled,
    control.iam_group_not_empty,
    control.iam_policy_no_star_star,
    control.iam_root_user_no_access_keys,
    control.iam_user_in_group,
    control.iam_user_no_inline_attached_policies,
    control.iam_user_unused_credentials_90
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ac_5" {
  title       = "PR.AC-5"
  description = "Network integrity is protected (e.g., network segregation, network segmentation)."

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
    control.s3_public_access_block_account,
    control.sagemaker_notebook_instance_direct_internet_access_disabled,
    control.vpc_default_security_group_restricts_all_traffic,
    control.vpc_security_group_restrict_ingress_common_ports_all,
    control.vpc_security_group_restrict_ingress_ssh_all,
    control.vpc_security_group_restrict_ingress_tcp_udp_all
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ac_6" {
  title       = "PR.AC-6"
  description = "Identities are proofed and bound to credentials and asserted in interactions."

  children = [
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_trail_enabled,
    control.emr_cluster_kerberos_enabled,
    control.redshift_cluster_encryption_logging_enabled,
    control.s3_bucket_logging_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ac_7" {
  title       = "PR.AC-7"
  description = "Users, devices, and other assets are authenticated (e.g., single-factor, multi-factor) commensurate with the risk of the transaction (e.g., individuals’ security and privacy risks and other organizational risks)."

  children = [
    control.iam_root_user_hardware_mfa_enabled,
    control.iam_root_user_mfa_enabled,
    control.iam_user_console_access_mfa_enabled,
    control.iam_user_mfa_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ds" {
  title       = "Data Security (PR.DS)"
  description = "Information and records (data) are managed consistent with the organization’s risk strategy to protect the confidentiality, integrity, and availability of information."

  children = [
    benchmark.nist_csf_pr_ds_1,
    benchmark.nist_csf_pr_ds_2,
    benchmark.nist_csf_pr_ds_3,
    benchmark.nist_csf_pr_ds_4,
    benchmark.nist_csf_pr_ds_5,
    benchmark.nist_csf_pr_ds_6,
    benchmark.nist_csf_pr_ds_7,
    benchmark.nist_csf_pr_ds_8
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ds_1" {
  title       = "PR.DS-1"
  description = "Data-at-rest is protected."

  children = [
    control.apigateway_stage_cache_encryption_at_rest_enabled,
    control.cloudtrail_trail_logs_encrypted_with_kms_cmk,
    control.ebs_attached_volume_encryption_enabled,
    control.efs_file_system_encrypt_data_at_rest,
    control.es_domain_encryption_at_rest_enabled,
    control.kms_key_not_pending_deletion,
    control.log_group_encryption_at_rest_enabled,
    control.rds_db_instance_encryption_at_rest_enabled,
    control.s3_bucket_default_encryption_enabled,
    control.s3_bucket_object_lock_enabled,
    control.sagemaker_endpoint_configuration_encryption_at_rest_enabled,
    control.sagemaker_notebook_instance_encryption_at_rest_enabled,
    control.sns_topic_encrypted_at_rest
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ds_2" {
  title       = "PR.DS-2"
  description = "Data-in-transit is protected."

  children = [
    control.acm_certificate_expires_30_days,
    control.elb_application_lb_drop_http_headers,
    control.elb_application_lb_redirect_http_request_to_https,
    control.elb_classic_lb_use_ssl_certificate,
    control.elb_classic_lb_use_tls_https_listeners,
    control.es_domain_node_to_node_encryption_enabled,
    control.redshift_cluster_encryption_in_transit_enabled,
    control.s3_bucket_enforces_ssl
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ds_3" {
  title       = "PR.DS-3"
  description = "Assets are formally managed throughout removal, transfers, and disposition."

  children = [
    control.ec2_instance_ssm_managed,
    control.ssm_managed_instance_compliance_association_compliant,
    control.vpc_eip_associated,
    control.vpc_security_group_associated_to_eni
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ds_4" {
  title       = "PR.DS-4"
  description = "Adequate capacity to ensure availability is maintained."

  children = [
    control.autoscaling_group_with_lb_use_health_check,
    control.elasticache_redis_cluster_automatic_backup_retention_15_days,
    control.elb_application_lb_deletion_protection_enabled,
    control.rds_db_instance_and_cluster_enhanced_monitoring_enabled,
    control.rds_db_instance_backup_enabled,
    control.rds_db_instance_multiple_az_enabled,
    control.s3_bucket_cross_region_replication_enabled,
    control.s3_bucket_versioning_enabled,
    control.vpc_vpn_tunnel_up
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ds_5" {
  title       = "PR.DS-5"
  description = "Protections against data leaks are implemented."

  children = [
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_s3_data_events_enabled,
    control.cloudtrail_trail_enabled,
    control.codebuild_project_plaintext_env_variables_no_sensitive_aws_values,
    control.dms_replication_instance_not_publicly_accessible,
    control.ebs_snapshot_not_publicly_restorable,
    control.elb_application_classic_lb_logging_enabled,
    control.guardduty_enabled,
    control.lambda_function_restrict_public_access,
    control.rds_db_instance_prohibit_public_access,
    control.rds_db_snapshot_prohibit_public_access,
    control.redshift_cluster_prohibit_public_access,
    control.s3_bucket_logging_enabled,
    control.s3_bucket_restrict_public_read_access,
    control.s3_bucket_restrict_public_write_access,
    control.s3_public_access_block_account,
    control.sagemaker_notebook_instance_direct_internet_access_disabled,
    control.securityhub_enabled,
    control.vpc_flow_logs_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ds_6" {
  title       = "PR.DS-6"
  description = "Integrity checking mechanisms are used to verify software, firmware, and information integrity."

  children = [
    control.cloudtrail_trail_validation_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ds_7" {
  title       = "PR.DS-7"
  description = "The development and testing environment(s) are separate from the production environment."

  children = [
    control.cloudtrail_trail_validation_enabled,
    control.ebs_attached_volume_delete_on_termination_enabled,
    control.ec2_instance_ssm_managed,
    control.ec2_stopped_instance_30_days,
    control.elb_application_lb_deletion_protection_enabled,
    control.ssm_managed_instance_compliance_association_compliant,
    control.vpc_security_group_restrict_ingress_ssh_all
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ds_8" {
  title       = "PR.DS-8"
  description = "Integrity checking mechanisms are used to verify hardware integrity."

  children = [
    control.ec2_instance_ssm_managed,
    control.securityhub_enabled
  ]

  tags = local.nist_csf_common_tags
}


benchmark "nist_csf_pr_ip" {
  title       = "Information Protection Processes and Procedures (PR.IP)"
  description = "Security policies (that address purpose, scope, roles, responsibilities, management commitment, and coordination among organizational entities), processes, and procedures are maintained and used to manage protection of information systems and assets."

  children = [
    benchmark.nist_csf_pr_ip_1,
    benchmark.nist_csf_pr_ip_2,
    benchmark.nist_csf_pr_ip_3,
    benchmark.nist_csf_pr_ip_4,
    benchmark.nist_csf_pr_ip_7,
    benchmark.nist_csf_pr_ip_8,
    benchmark.nist_csf_pr_ip_9,
    benchmark.nist_csf_pr_ip_12
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ip_1" {
  title       = "PR.IP-1"
  description = "A baseline configuration of information technology/industrial control systems is created and maintained incorporating security principles (e.g. concept of least functionality)."

  children = [
    control.ebs_attached_volume_delete_on_termination_enabled,
    control.ec2_instance_ssm_managed,
    control.ec2_stopped_instance_30_days,
    control.ssm_managed_instance_compliance_association_compliant
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ip_2" {
  title       = "PR.IP-2"
  description = "A System Development Life Cycle to manage systems is implemented."

  children = [
    control.codebuild_project_plaintext_env_variables_no_sensitive_aws_values,
    control.codebuild_project_source_repo_oauth_configured,
    control.ec2_instance_ssm_managed
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ip_3" {
  title       = "PR.IP-3"
  description = "Configuration change control processes are in place."

  children = [
    control.elb_application_lb_deletion_protection_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ip_4" {
  title       = "PR.IP-4"
  description = "Backups of information are conducted, maintained, and tested periodically."

  children = [
    control.dynamodb_table_point_in_time_recovery_enabled,
    control.elasticache_redis_cluster_automatic_backup_retention_15_days,
    control.rds_db_instance_backup_enabled,
    control.s3_bucket_cross_region_replication_enabled,
    control.s3_bucket_versioning_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ip_7" {
  title       = "PR.IP-7"
  description = "Protection processes are improved."

  children = [
    control.ec2_instance_ebs_optimized
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ip_8" {
  title       = "PR.IP-8"
  description = "Effectiveness of protection technologies is shared."

  children = [
    control.dms_replication_instance_not_publicly_accessible,
    control.ebs_snapshot_not_publicly_restorable,
    control.ec2_instance_not_publicly_accessible,
    control.eks_cluster_endpoint_restrict_public_access,
    control.emr_cluster_master_nodes_no_public_ip,
    control.lambda_function_restrict_public_access,
    control.rds_db_instance_prohibit_public_access,
    control.rds_db_snapshot_prohibit_public_access,
    control.redshift_cluster_prohibit_public_access,
    control.s3_bucket_restrict_public_read_access,
    control.s3_bucket_restrict_public_write_access,
    control.s3_public_access_block_account,
    control.s3_public_access_block_bucket_account,
    control.sagemaker_notebook_instance_direct_internet_access_disabled,
    control.vpc_subnet_auto_assign_public_ip_disabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ip_9" {
  title       = "PR.IP-9"
  description = "Response plans (Incident Response and Business Continuity) and recovery plans (Incident Recovery and Disaster Recovery) are in place and managed."

  children = [
    control.backup_plan_min_retention_35_days,
    control.backup_recovery_point_encryption_enabled,
    control.backup_recovery_point_manual_deletion_disabled,
    control.dynamodb_table_auto_scaling_enabled,
    control.dynamodb_table_in_backup_plan,
    control.dynamodb_table_point_in_time_recovery_enabled,
    control.dynamodb_table_protected_by_backup_plan,
    control.ebs_volume_in_backup_plan,
    control.ebs_volume_protected_by_backup_plan,
    control.ec2_instance_protected_by_backup_plan,
    control.efs_file_system_in_backup_plan,
    control.efs_file_system_protected_by_backup_plan,
    control.elasticache_redis_cluster_automatic_backup_retention_15_days,
    control.elb_application_lb_deletion_protection_enabled,
    control.elb_classic_lb_cross_zone_load_balancing_enabled,
    control.fsx_file_system_protected_by_backup_plan,
    control.rds_db_cluster_aurora_protected_by_backup_plan,
    control.rds_db_instance_in_backup_plan,
    control.rds_db_instance_multiple_az_enabled,
    control.rds_db_instance_protected_by_backup_plan,
    control.redshift_cluster_automatic_snapshots_min_7_days,
    control.s3_bucket_cross_region_replication_enabled,
    control.s3_bucket_versioning_enabled,
    control.vpc_vpn_tunnel_up
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ip_12" {
  title       = "PR.IP-12"
  description = "A vulnerability management plan is developed and implemented."

  children = [
    control.config_enabled_all_regions,
    control.ec2_instance_ssm_managed,
    control.ssm_managed_instance_compliance_association_compliant,
    control.ssm_managed_instance_compliance_patch_compliant
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ma" {
  title       = "Maintenance (PR.MA)"
  description = "Maintenance and repairs of industrial control and information system components are performed consistent with policies and procedures."

  children = [
    benchmark.nist_csf_pr_ma_2
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_ma_2" {
  title       = "PR.MA-2"
  description = "Remote maintenance of organizational assets is approved, logged, and performed in a manner that prevents unauthorized access."

  children = [
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_trail_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_pt" {
  title       = "Protective Technology (PR.PT)"
  description = "Maintenance and repairs of industrial control and information system components are performed consistent with policies and procedures."

  children = [
    benchmark.nist_csf_pr_pt_1,
    benchmark.nist_csf_pr_pt_3,
    benchmark.nist_csf_pr_pt_4,
    benchmark.nist_csf_pr_pt_5
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_pt_1" {
  title       = "PR.PT-1"
  description = "Audit/log records are determined, documented, implemented, and reviewed in accordance with policy."

  children = [
    control.apigateway_stage_logging_enabled,
    control.cloudtrail_multi_region_trail_enabled,
    control.cloudtrail_trail_enabled,
    control.cloudtrail_trail_integrated_with_logs,
    control.elb_application_classic_lb_logging_enabled,
    control.s3_bucket_logging_enabled,
    control.vpc_flow_logs_enabled
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_pt_3" {
  title       = "PR.PT-3"
  description = "The principle of least functionality is incorporated by configuring systems to provide only essential capabilities."

  children = [
    control.dms_replication_instance_not_publicly_accessible,
    control.ebs_snapshot_not_publicly_restorable,
    control.iam_policy_no_star_star,
    control.iam_root_user_no_access_keys,
    control.iam_user_no_inline_attached_policies,
    control.lambda_function_restrict_public_access,
    control.rds_db_snapshot_prohibit_public_access,
    control.redshift_cluster_prohibit_public_access,
    control.s3_bucket_restrict_public_read_access,
    control.s3_bucket_restrict_public_write_access,
    control.s3_public_access_block_account
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_pt_4" {
  title       = "PR.PT-4"
  description = "Communications and control networks are protected."

  children = [
    control.ec2_instance_in_vpc,
    control.es_domain_in_vpc,
    control.lambda_function_in_vpc,
    control.rds_db_instance_prohibit_public_access,
    control.redshift_cluster_prohibit_public_access,
    control.vpc_security_group_restrict_ingress_common_ports_all,
    control.vpc_security_group_restrict_ingress_ssh_all,
    control.vpc_security_group_restrict_ingress_tcp_udp_all
  ]

  tags = local.nist_csf_common_tags
}

benchmark "nist_csf_pr_pt_5" {
  title       = "PR.PT-5"
  description = "Mechanisms (e.g., failsafe, load balancing, hot swap) are implemented to achieve resilience requirements in normal and adverse situations."

  children = [
    control.autoscaling_group_with_lb_use_health_check,
    control.elasticache_redis_cluster_automatic_backup_retention_15_days,
    control.elb_application_lb_deletion_protection_enabled,
    control.rds_db_instance_backup_enabled,
    control.rds_db_instance_multiple_az_enabled,
    control.s3_bucket_cross_region_replication_enabled,
    control.s3_bucket_versioning_enabled,
    control.vpc_vpn_tunnel_up
  ]

  tags = local.nist_csf_common_tags
}
