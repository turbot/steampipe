locals {
  gdpr_article_32_common_tags = merge(local.gdpr_common_tags, {
    gdpr_article = "32"
  })
}

benchmark "article_32" {
  title       = "Article 32 Security of processing"
  documentation = file("./gdpr/docs/article_32.md")
  children = [
    control.acm_certificate_expires_30_days,
    control.apigateway_stage_cache_encryption_at_rest_enabled,
    control.cloudfront_distribution_encryption_in_transit_enabled,
    control.cloudtrail_trail_logs_encrypted_with_kms_cmk,
    control.cloudtrail_trail_validation_enabled,
    control.dax_cluster_encryption_at_rest_enabled,
    control.dynamodb_table_encrypted_with_kms_cmk,
    control.dynamodb_table_encryption_enabled,
    control.ebs_attached_volume_encryption_enabled,
    control.ebs_volume_encryption_at_rest_enabled,
    control.efs_file_system_encrypt_data_at_rest,
    control.elb_application_lb_drop_http_headers,
    control.elb_application_lb_redirect_http_request_to_https,
    control.elb_classic_lb_use_ssl_certificate,
    control.elb_classic_lb_use_tls_https_listeners,
    control.es_domain_encryption_at_rest_enabled,
    control.es_domain_node_to_node_encryption_enabled,
    control.log_group_encryption_at_rest_enabled,
    control.rds_db_instance_encryption_at_rest_enabled,
    control.rds_db_instance_in_backup_plan,
    control.rds_db_instance_logging_enabled,
    control.rds_db_snapshot_encrypted_at_rest,
    control.redshift_cluster_automatic_snapshots_min_7_days,
    control.redshift_cluster_encryption_in_transit_enabled,
    control.redshift_cluster_encryption_logging_enabled,
    control.s3_bucket_default_encryption_enabled,
    control.s3_bucket_default_encryption_enabled_kms,
    control.s3_bucket_enforces_ssl,
    control.sagemaker_endpoint_configuration_encryption_at_rest_enabled,
    control.sagemaker_notebook_instance_encryption_at_rest_enabled,
    control.sns_topic_encrypted_at_rest,
    control.wafv2_web_acl_logging_enabled
  ]

  tags = local.gdpr_article_32_common_tags
}
