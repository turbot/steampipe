locals {
  cis_v130_3_common_tags = merge(local.cis_v130_common_tags, {
    cis_section_id = "3"
  })
}

benchmark "cis_v130_3" {
  title = "3 Logging"
  #documentation = file("docs/cis_v130_3.md")
  children = [
    control.cis_v130_3_1,
    control.cis_v130_3_2,
    control.cis_v130_3_3,
    control.cis_v130_3_4,
    control.cis_v130_3_5,
    control.cis_v130_3_6,
    control.cis_v130_3_7,
    control.cis_v130_3_8,
    control.cis_v130_3_9,
    control.cis_v130_3_10,
    control.cis_v130_3_11
  ]
  tags = local.cis_v130_3_common_tags
}

control "cis_v130_3_1" {
  title       = "3.1 Ensure CloudTrail is enabled in all regions"
  description = "AWS CloudTrail is a web service that records AWS API calls for your account and delivers log files to you. The recorded information includes the identity of the API caller, the time of the API call, the source IP address of the API caller, the request parameters, and the response elements returned by the AWS service. CloudTrail provides a history of AWS API calls for an account, including API calls made via the Management Console, SDKs, command line tools, and higher-level AWS services (such as CloudFormation)."
  sql         = query.ok.sql
  #documentation = file("docs/cis_v130_3_1.md")

  tags = merge(local.cis_v130_3_common_tags, {
    cis_item_id  = "3.1"
    cis_type     = "automated"
    cis_levels   = "1"
    cis_controls = "6.2"
  })
}

control "cis_v130_3_2" {
  title       = "3.2 Ensure CloudTrail log file validation is enabled."
  description = "CloudTrail log file validation creates a digitally signed digest file containing a hash of each log that CloudTrail writes to S3. These digest files can be used to determine whether a log file was changed, deleted, or unchanged after CloudTrail delivered the log. It is recommended that file validation be enabled on all CloudTrails."
  sql         = query.ok.sql

  tags = merge(local.cis_v130_3_common_tags, {
    cis_item_id  = "3.2"
    cis_type     = "automated"
    cis_levels   = "2"
    cis_controls = "6"
  })
}

control "cis_v130_3_3" {
  title       = "3.3 Ensure the S3 bucket used to store CloudTrail logs is not publicly accessible"
  description = "CloudTrail logs a record of every API call made in your AWS account. These logs file are stored in an S3 bucket. It is recommended that the bucket policy or access control list (ACL) applied to the S3 bucket that CloudTrail logs to prevent public access to the CloudTrail logs."
  sql         = query.ok.sql
  #documentation = file("docs/cis_v130_3_3.md")

  tags = merge(local.cis_v130_3_common_tags, {
    cis_item_id  = "3.3"
    cis_type     = "automated"
    cis_levels   = "1"
    cis_controls = "14.6"
  })
}

control "cis_v130_3_4" {
  title       = "3.4 Ensure CloudTrail trails are integrated with CloudWatch Logs"
  description = "AWS CloudTrail is a web service that records AWS API calls made in a given AWS account. The recorded information includes the identity of the API caller, the time of the API call, the source IP address of the API caller, the request parameters, and the response elements returned by the AWS service. CloudTrail uses Amazon S3 for log file storage and delivery, so log files are stored durably. In addition to capturing CloudTrail logs within a specified S3 bucket for long term analysis, realtime analysis can be performed by configuring CloudTrail to send logs to CloudWatch Logs. For a trail that is enabled in all regions in an account, CloudTrail sends log files from all those regions to a CloudWatch Logs log group. It is recommended that CloudTrail logs be sent to CloudWatch Logs."
  sql         = query.ok.sql
  #documentation = file("./cis_v130/docs/cis_v130_3_4.md")

  tags = merge(local.cis_v130_3_common_tags, {
    "cis_item_id" = "3.4"
    "cis_type"    = "automated"
    "cis_level"   = "1"
    "cis_control" = "6.2"
  })
}

control "cis_v130_3_5" {
  title       = "3.5 Ensure AWS Config is enabled in all regions"
  description = "AWS Config is a web service that performs configuration management of supported AWS resources within your account and delivers log files to you. The recorded information includes the configuration item (AWS resource), relationships between configuration items (AWS resources), any configuration changes between resources. It is recommended to enable AWS Config be enabled in all regions."
  sql         = query.ok.sql
  #documentation = file("./cis_v130/docs/cis_v130_3_5.md")

  tags = merge(local.cis_v130_3_common_tags, {
    "cis_item_id" = "3.5"
    "cis_type"    = "automated"
    "cis_level"   = "1"
    "cis_control" = "1.4,11.2,16.1"
  })
}

control "cis_v130_3_6" {
  title       = "3.6 Ensure S3 bucket access logging is enabled on the CloudTrail S3 bucket"
  description = "S3 Bucket Access Logging generates a log that contains access records for each request made to your S3 bucket. An access log record contains details about the request, such as the request type, the resources specified in the request worked, and the time and date the request was processed. It is recommended that bucket access logging be enabled on the CloudTrail S3 bucket."
  sql         = query.ok.sql
  #documentation = file("./cis_v130/docs/cis_v130_3_6.md")

  tags = merge(local.cis_v130_3_common_tags, {
    "cis_item_id" = "3.6"
    "cis_type"    = "automated"
    "cis_level"   = "1"
    "cis_control" = "6.2,14.9"
  })
}

control "cis_v130_3_7" {
  title       = "3.7 Ensure CloudTrail logs are encrypted at rest using KMS CMKs"
  description = "AWS CloudTrail is a web service that records AWS API calls for an account and makes those logs available to users and resources in accordance with IAM policies. AWS Key Management Service (KMS) is a managed service that helps create and control the encryption keys used to encrypt account data, and uses Hardware Security Modules (HSMs) to protect the security of encryption keys. CloudTrail logs can be configured to leverage server side encryption (SSE) and KMS customer created master keys (CMK) to further protect CloudTrail logs. It is recommended that CloudTrail be configured to use SSE-KMS."
  sql         = query.ok.sql
  #documentation = file("./cis_v130/docs/cis_v130_3_7.md")

  tags = merge(local.cis_v130_3_common_tags, {
    "cis_item_id" = "3.7"
    "cis_type"    = "automated"
    "cis_level"   = "2"
    "cis_control" = "6"
  })
}

control "cis_v130_3_8" {
  title       = "3.8 Ensure rotation for customer created CMKs is enabled"
  description = "AWS Key Management Service (KMS) allows customers to rotate the backing key which is key material stored within the KMS which is tied to the key ID of the Customer Created customer master key (CMK). It is the backing key that is used to perform cryptographic operations such as encryption and decryption. Automated key rotation currently retains all prior backing keys so that decryption of encrypted data can take place transparently. It is recommended that CMK key rotation be enabled."
  sql         = query.ok.sql
  #documentation = file("./cis_v130/docs/cis_v130_3_8.md")

  tags = merge(local.cis_v130_3_common_tags, {
    "cis_item_id" = "3.8"
    "cis_type"    = "automated"
    "cis_level"   = "2"
    "cis_control" = "6"
  })
}

control "cis_v130_3_9" {
  title       = "3.9 Ensure VPC flow logging is enabled in all VPCs"
  description = "VPC Flow Logs is a feature that enables you to capture information about the IP traffic going to and from network interfaces in your VPC. After you've created a flow log, you can view and retrieve its data in Amazon CloudWatch Logs. It is recommended that VPC Flow Logs be enabled for packet \"Rejects\" for VPCs."
  sql         = query.ok.sql
  #documentation = file("./cis_v130/docs/cis_v130_3_9.md")

  tags = merge(local.cis_v130_3_common_tags, {
    "cis_item_id" = "3.9"
    "cis_type"    = "automated"
    "cis_level"   = "2"
    "cis_control" = "6.2,12.5"
  })
}

control "cis_v130_3_10" {
  title         = "3.10 Ensure that Object-level logging for write events is enabled for S3 bucket"
  description   = "S3 object-level API operations such as GetObject, DeleteObject, and PutObject are called data events. By default, CloudTrail trails don't log data events and so it is recommended to enable Object-level logging for S3 buckets."
  sql           = query.ok.sql
  documentation = file("./cis_v130/docs/cis_v130_3_10.md")

  tags = merge(local.cis_v130_3_common_tags, {
    "cis_item_id" = "3.10"
    "cis_type"    = "automated"
    "cis_level"   = "2"
    "cis_control" = "6.2,6.3"
  })
}

control "cis_v130_3_11" {
  title       = "3.11 Ensure that Object-level logging for read events is enabled for S3 bucket"
  description = "S3 object-level API operations such as GetObject, DeleteObject, and PutObject are called data events. By default, CloudTrail trails don't log data events and so it is recommended to enable Object-level logging for S3 buckets."
  sql         = query.ok.sql
  # documentation = file("./cis_v130/docs/cis_v130_3_11.md")

  tags = merge(local.cis_v130_3_common_tags, {
    "cis_item_id" = "3.11"
    "cis_type"    = "automated"
    "cis_level"   = "2"
    "cis_control" = "6.2,6.3"
  })
}
