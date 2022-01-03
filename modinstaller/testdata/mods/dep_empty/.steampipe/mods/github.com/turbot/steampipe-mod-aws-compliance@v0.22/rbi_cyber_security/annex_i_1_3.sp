benchmark "rbi_cyber_security_annex_i_1_3" {
  title       = "Annex I (1.3)"
  description = "Appropriately manage and provide protection within and outside UCB/network, keeping in mind how the data/information is stored, transmitted, processed, accessed and put to use within/outside the UCB’s network, and level of risk they are exposed to depending on the sensitivity of the data/information."

  children = [
    control.acm_certificate_expires_30_days,
    control.apigateway_rest_api_stage_use_ssl_certificate,
    control.apigateway_stage_cache_encryption_at_rest_enabled,
    control.autoscaling_launch_config_public_ip_disabled,
    control.cloudtrail_trail_logs_encrypted_with_kms_cmk,
    control.dms_replication_instance_not_publicly_accessible,
    control.dynamodb_table_encrypted_with_kms_cmk,
    control.ebs_attached_volume_encryption_enabled,
    control.ebs_snapshot_not_publicly_restorable,
    control.ebs_volume_encryption_at_rest_enabled,
    control.ec2_instance_in_vpc,
    control.ec2_instance_not_publicly_accessible,
    control.efs_file_system_encrypt_data_at_rest,
    control.elb_application_lb_drop_http_headers,
    control.elb_application_lb_redirect_http_request_to_https,
    control.elb_application_network_lb_use_ssl_certificate,
    control.elb_classic_lb_use_ssl_certificate,
    control.elb_classic_lb_use_tls_https_listeners,
    control.emr_cluster_master_nodes_no_public_ip,
    control.es_domain_encryption_at_rest_enabled,
    control.es_domain_in_vpc,
    control.es_domain_node_to_node_encryption_enabled,
    control.kms_cmk_rotation_enabled,
    control.kms_key_not_pending_deletion,
    control.lambda_function_in_vpc,
    control.lambda_function_restrict_public_access,
    control.log_group_encryption_at_rest_enabled,
    control.rds_db_instance_encryption_at_rest_enabled,
    control.rds_db_instance_prohibit_public_access,
    control.rds_db_snapshot_encrypted_at_rest,
    control.rds_db_snapshot_prohibit_public_access,
    control.redshift_cluster_encryption_in_transit_enabled,
    control.redshift_cluster_encryption_logging_enabled,
    control.redshift_cluster_kms_enabled,
    control.redshift_cluster_prohibit_public_access,
    control.s3_bucket_default_encryption_enabled_kms,
    control.s3_bucket_default_encryption_enabled,
    control.s3_bucket_enforces_ssl,
    control.s3_bucket_restrict_public_read_access,
    control.s3_bucket_restrict_public_write_access,
    control.s3_public_access_block_bucket_account,
    control.sagemaker_endpoint_configuration_encryption_at_rest_enabled,
    control.sagemaker_notebook_instance_direct_internet_access_disabled,
    control.sagemaker_notebook_instance_encryption_at_rest_enabled,
    control.sns_topic_encrypted_at_rest,
    control.vpc_igw_attached_to_authorized_vpc,
    control.vpc_route_table_restrict_public_access_to_igw,
    control.vpc_security_group_restrict_ingress_common_ports_all,
    control.vpc_subnet_auto_assign_public_ip_disabled
  ]

  tags = merge(local.rbi_cyber_security_common_tags, {
    rbi_cyber_security_item_id = "annex_i_1_3"
  })
}
