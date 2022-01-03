select
  -- Required Columns
  arn as resource,
  case
    when
      log_publishing_options -> 'AUDIT_LOGS' -> 'Enabled' = 'true'
      and log_publishing_options -> 'AUDIT_LOGS' -> 'CloudWatchLogsLogGroupArn' is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when
      log_publishing_options -> 'AUDIT_LOGS' -> 'Enabled' = 'true'
      and log_publishing_options -> 'AUDIT_LOGS' -> 'CloudWatchLogsLogGroupArn' is not null then title || ' audit logging enabled.'
    else title || ' audit logging disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_elasticsearch_domain;