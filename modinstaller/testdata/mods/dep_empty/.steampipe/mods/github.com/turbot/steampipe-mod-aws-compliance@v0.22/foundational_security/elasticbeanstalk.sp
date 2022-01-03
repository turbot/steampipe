locals {
  foundational_security_elasticbeanstalk_common_tags = merge(local.foundational_security_common_tags, {
    service = "elasticbeanstalk"
  })
}

benchmark "foundational_security_elasticbeanstalk" {
  title         = "Elastic Beanstalk"
  documentation = file("./foundational_security/docs/foundational_security_elasticbeanstalk.md")
  children = [
    control.foundational_security_elasticbeanstalk_1
  ]
  tags          = local.foundational_security_elasticbeanstalk_common_tags
}

control "foundational_security_elasticbeanstalk_1" {
  title         = "1 Elastic Beanstalk environments should have enhanced health reporting enabled"
  description   = "This control checks whether enhanced health reporting is enabled for your AWS Elastic Beanstalk environments.Elastic Beanstalk enhanced health reporting enables a more rapid response to changes in the health of the underlying infrastructure. These changes could result in a lack of availability of the application."
  severity      = "low"
  sql           = query.elastic_beanstalk_enhanced_health_reporting_enabled.sql
  documentation = file("./foundational_security/docs/foundational_security_elasticbeanstalk_1.md")

  tags = merge(local.foundational_security_elasticbeanstalk_common_tags, {
    foundational_security_item_id  = "elasticbeanstalk_1"
    foundational_security_category = "application_monitoring"
  })
}