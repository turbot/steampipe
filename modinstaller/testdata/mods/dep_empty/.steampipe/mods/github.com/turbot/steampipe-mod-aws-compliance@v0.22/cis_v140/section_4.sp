locals {
  cis_v140_4_common_tags = merge(local.cis_v140_common_tags, {
    cis_section_id = "4"
  })
}

benchmark "cis_v140_4" {
  title         = "4 Monitoring"
  documentation = file("./cis_v140/docs/cis_v140_4.md")
  tags          = local.cis_v140_4_common_tags
  children = [
    control.cis_v140_4_1,
    control.cis_v140_4_2,
    control.cis_v140_4_3,
    control.cis_v140_4_4,
    control.cis_v140_4_5,
    control.cis_v140_4_6,
    control.cis_v140_4_7,
    control.cis_v140_4_8,
    control.cis_v140_4_9,
    control.cis_v140_4_10,
    control.cis_v140_4_11,
    control.cis_v140_4_12,
    control.cis_v140_4_13,
    control.cis_v140_4_14,
    control.cis_v140_4_15
  ]
}

control "cis_v140_4_1" {
  title         = "4.1 Ensure a log metric filter and alarm exist for unauthorized API calls"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for unauthorized API calls."
  sql           = query.log_metric_filter_unauthorized_api.sql
  documentation = file("./cis_v140/docs/cis_v140_4_1.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.1"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_2" {
  title         = "4.2 Ensure a log metric filter and alarm exist for Management Console sign-in without MFA"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for console logins that are not protected by multi-factor authentication (MFA)."
  sql           = query.log_metric_filter_console_login_mfa.sql
  documentation = file("./cis_v140/docs/cis_v140_4_2.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.2"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_3" {
  title         = "4.3 Ensure a log metric filter and alarm exist for usage of 'root' account"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for root login attempts."
  sql           = query.log_metric_filter_root_login.sql
  documentation = file("./cis_v140/docs/cis_v140_4_3.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.3"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_4" {
  title         = "4.4 Ensure a log metric filter and alarm exist for IAM policy changes"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established changes made to Identity and Access Management (IAM) policies."
  sql           = query.log_metric_filter_iam_policy.sql
  documentation = file("./cis_v140/docs/cis_v140_4_4.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.4"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_5" {
  title         = "4.5 Ensure a log metric filter and alarm exist for CloudTrail configuration changes"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for detecting changes to CloudTrail's configurations."
  sql           = query.log_metric_filter_cloudtrail_configuration.sql
  documentation = file("./cis_v140/docs/cis_v140_4_5.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.5"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_6" {
  title         = "4.6 Ensure a log metric filter and alarm exist for AWS Management Console authentication failures"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for failed console authentication attempts."
  sql           = query.log_metric_filter_console_authentication_failure.sql
  documentation = file("./cis_v140/docs/cis_v140_4_6.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.6"
    cis_level   = "2"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_7" {
  title         = "4.7 Ensure a log metric filter and alarm exist for disabling or scheduled deletion of customer created CMKs"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for customer created CMKs which have changed state to disabled or scheduled deletion."
  sql           = query.log_metric_filter_disable_or_delete_cmk.sql
  documentation = file("./cis_v140/docs/cis_v140_4_7.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.7"
    cis_level   = "2"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_8" {
  title         = "4.8 Ensure a log metric filter and alarm exist for S3 bucket policy changes"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for changes to S3 bucket policies."
  sql           = query.log_metric_filter_bucket_policy.sql
  documentation = file("./cis_v140/docs/cis_v140_4_8.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.8"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_9" {
  title         = "4.9 Ensure a log metric filter and alarm exist for AWS Config configuration changes"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for detecting changes to CloudTrail's configurations."
  sql           = query.log_metric_filter_config_configuration.sql
  documentation = file("./cis_v140/docs/cis_v140_4_9.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.9"
    cis_level   = "2"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_10" {
  title         = "4.10 Ensure a log metric filter and alarm exist for security group changes"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. Security Groups are a stateful packet filter that controls ingress and egress traffic within a VPC. It is recommended that a metric filter and alarm be established for detecting changes to Security Groups."
  sql           = query.log_metric_filter_security_group.sql
  documentation = file("./cis_v140/docs/cis_v140_4_10.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.10"
    cis_level   = "2"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_11" {
  title         = "4.11 Ensure a log metric filter and alarm exist for changes to Network Access Control Lists (NACL)"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. NACLs are used as a stateless packet filter to control ingress and egress traffic for subnets within a VPC. It is recommended that a metric filter and alarm be established for changes made to NACLs."
  sql           = query.log_metric_filter_network_acl.sql
  documentation = file("./cis_v140/docs/cis_v140_4_11.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.11"
    cis_level   = "2"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_12" {
  title         = "4.12 Ensure a log metric filter and alarm exist for changes to network gateways"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. Network gateways are required to send/receive traffic to a destination outside of a VPC. It is recommended that a metric filter and alarm be established for changes to network gateways."
  sql           = query.log_metric_filter_network_gateway.sql
  documentation = file("./cis_v140/docs/cis_v140_4_12.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.12"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_13" {
  title         = "4.13 Ensure a log metric filter and alarm exist for route table changes"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. Routing tables are used to route network traffic between subnets and to network gateways. It is recommended that a metric filter and alarm be established for changes to route tables."
  sql           = query.log_metric_filter_route_table.sql
  documentation = file("./cis_v140/docs/cis_v140_4_13.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.13"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_14" {
  title         = "4.14 Ensure a log metric filter and alarm exist for VPC changes"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is possible to have more than 1 VPC within an account, in addition it is also possible to create a peer connection between 2 VPCs enabling network traffic to route between VPCs. It is recommended that a metric filter and alarm be established for changes made to VPCs."
  sql           = query.log_metric_filter_vpc.sql
  documentation = file("./cis_v140/docs/cis_v140_4_14.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.14"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}

control "cis_v140_4_15" {
  title         = "4.15 Ensure a log metric filter and alarm exists for AWS Organizations changes"
  description   = "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs and establishing corresponding metric filters and alarms. It is recommended that a metric filter and alarm be established for AWS Organizations changes made in the master AWS Account."
  sql           = query.log_metric_filter_organization.sql
  documentation = file("./cis_v140/docs/cis_v140_4_15.md")

  tags = merge(local.cis_v140_4_common_tags, {
    cis_item_id = "4.15"
    cis_level   = "1"
    cis_type    = "automated"
    service     = "cloudwatch"
  })
}
