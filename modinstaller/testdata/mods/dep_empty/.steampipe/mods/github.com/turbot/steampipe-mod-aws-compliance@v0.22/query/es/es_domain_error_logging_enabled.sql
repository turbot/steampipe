select
  -- Required Columns
  arn as resource,
  case
    when
      log_publishing_options -> 'ES_APPLICATION_LOGS' -> 'Enabled' = 'true'
      and log_publishing_options -> 'ES_APPLICATION_LOGS' -> 'CloudWatchLogsLogGroupArn' is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when
      log_publishing_options -> 'ES_APPLICATION_LOGS' -> 'Enabled' = 'true'
      and log_publishing_options -> 'ES_APPLICATION_LOGS' -> 'CloudWatchLogsLogGroupArn' is not null then title || ' error logging enabled.'
    else title || ' error logging disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_elasticsearch_domain;