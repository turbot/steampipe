locals {
  conformance_pack_cloudtrail_common_tags = {
    service = "cloudtrail"
  }
}

control "cloudtrail_trail_integrated_with_logs" {
  title       = "CloudTrail trails should be integrated with CloudWatch logs"
  description = "Use Amazon CloudWatch to centrally collect and manage log event activity. Inclusion of AWS CloudTrail data provides details of API call activity within your AWS account."
  sql         = query.cloudtrail_trail_integrated_with_logs.sql

  tags = merge(local.conformance_pack_cloudtrail_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "cloudtrail_s3_data_events_enabled" {
  title       = "All S3 buckets should log S3 data events in CloudTrail"
  description = "The collection of Simple Storage Service (Amazon S3) data events helps in detecting any anomalous activity. The details include AWS account information that accessed an Amazon S3 bucket, IP address, and time of event."
  sql         = query.cloudtrail_s3_data_events_enabled.sql

  tags = merge(local.conformance_pack_cloudtrail_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "cloudtrail_trail_logs_encrypted_with_kms_cmk" {
  title       = "CloudTrail trail logs should be encrypted with KMS CMK"
  description = "To help protect sensitive data at rest, ensure encryption is enabled for your Amazon CloudWatch Log Groups."
  sql         = query.cloudtrail_trail_logs_encrypted_with_kms_cmk.sql

  tags = merge(local.conformance_pack_cloudtrail_common_tags, {
    gdpr               = "true"
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
  })
}

control "cloudtrail_multi_region_trail_enabled" {
  title       = "At least one multi-region AWS CloudTrail should be present in an account"
  description = "AWS CloudTrail records AWS Management Console actions and API calls. You can identify which users and accounts called AWS, the source IP address from where the calls were made, and when the calls occurred. CloudTrail will deliver log files from all AWS Regions to your S3 bucket if MULTI_REGION_CLOUD_TRAIL_ENABLED is enabled."
  sql         = query.cloudtrail_multi_region_trail_enabled.sql

  tags = merge(local.conformance_pack_cloudtrail_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "cloudtrail_trail_validation_enabled" {
  title       = "CloudTrail trail log file validation should be enabled"
  description = "Utilize AWS CloudTrail log file validation to check the integrity of CloudTrail logs. Log file validation helps determine if a log file was modified or deleted or unchanged after CloudTrail delivered it. This feature is built using industry standard algorithms: SHA-256 for hashing and SHA-256 with RSA for digital signing. This makes it computationally infeasible to modify, delete or forge CloudTrail log files without detection."
  sql         = query.cloudtrail_trail_validation_enabled.sql

  tags = merge(local.conformance_pack_cloudtrail_common_tags, {
    gdpr              = "true"
    hipaa             = "true"
    nist_800_53_rev_4 = "true"
    soc_2             = "true"
  })
}

control "cloudtrail_trail_enabled" {
  title       = "At least one enabled trail should be present in a region"
  description = "AWS CloudTrail can help in non-repudiation by recording AWS Management Console actions and API calls. You can identify the users and AWS accounts that called an AWS service, the source IP address where the calls generated, and the timings of the calls. Details of captured data are seen within AWS CloudTrail Record Contents."
  sql         = query.cloudtrail_trail_enabled.sql

  tags = merge(local.conformance_pack_cloudtrail_common_tags, {
    hipaa              = "true"
    nist_800_53_rev_4  = "true"
    nist_csf           = "true"
    rbi_cyber_security = "true"
    soc_2              = "true"
  })
}

control "cloudtrail_security_trail_enabled" {
  title       = "At least one trail should be enabled with security best practices"
  description = "This rule helps ensure the use of AWS recommended security best practices for AWS CloudTrail, by checking for the enablement of multiple settings. These include the use of log encryption, log validation, and enabling AWS CloudTrail in multiple regions."
  sql         = query.cloudtrail_security_trail_enabled.sql

  tags = merge(local.conformance_pack_cloudtrail_common_tags, {
    gdpr              = "true"
    nist_800_53_rev_4 = "true"
    soc_2             = "true"
  })
}

control "cloudtrail_enabled_all_regions" {
  title       = "Ensure CloudTrail is enabled in all Regions"
  description = "CloudTrail is a service that records AWS API calls for your account and delivers log files to you. The recorded information includes the identity of the API caller, the time of the API call, the source IP address of the API caller, the request parameters, and the response elements returned by the AWS service."
  sql         = query.cloudtrail_enabled_all_regions.sql

  tags = merge(local.conformance_pack_cloudtrail_common_tags, {
    gdpr = "true"
  })
}

control "cloudtrail_s3_logging_enabled" {
  title       = "Ensure S3 bucket access logging is enabled on the CloudTrail S3 bucket"
  description = "S3 Bucket Access Logging generates a log that contains access records for each request made to your S3 bucket. An access log record contains details about the request, such as the request type, the resources specified in the request worked, and the time and date the request was processed. It is recommended that bucket access logging be enabled on the CloudTrail S3 bucket."
  sql         = query.cloudtrail_s3_logging_enabled.sql

  tags = merge(local.conformance_pack_cloudtrail_common_tags, {
    gdpr = "true"
  })
}

control "cloudtrail_bucket_not_public" {
  title       = "Ensure the S3 bucket CloudTrail logs to is not publicly accessible"
  description = "CloudTrail logs a record of every API call made in your account. These log files are stored in an S3 bucket. Security Hub recommends that the S3 bucket policy,or access control list (ACL), applied to the S3 bucket that CloudTrail logs to prevents public access to the CloudTrail logs.."
  sql         = query.cloudtrail_bucket_not_public.sql

  tags = merge(local.conformance_pack_cloudtrail_common_tags, {
    gdpr = "true"
  })
}