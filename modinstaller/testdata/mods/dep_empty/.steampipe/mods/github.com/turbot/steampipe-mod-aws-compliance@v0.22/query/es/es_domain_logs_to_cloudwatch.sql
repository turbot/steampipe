select
  -- Required Columns
  arn as resource,
  case
    when
      ( log_publishing_options -> 'ES_APPLICATION_LOGS' -> 'Enabled' = 'true'
        and log_publishing_options -> 'ES_APPLICATION_LOGS' -> 'CloudWatchLogsLogGroupArn' is not null
      )
      and
      ( log_publishing_options -> 'SEARCH_SLOW_LOGS' -> 'Enabled' = 'true'
        and log_publishing_options -> 'SEARCH_SLOW_LOGS' -> 'CloudWatchLogsLogGroupArn' is not null
      )
      and
      ( log_publishing_options -> 'INDEX_SLOW_LOGS' -> 'Enabled' = 'true'
        and log_publishing_options -> 'INDEX_SLOW_LOGS' -> 'CloudWatchLogsLogGroupArn' is not null
      )
       then 'ok'
    else 'alarm'
  end as status,
  case
    when
      ( log_publishing_options -> 'ES_APPLICATION_LOGS' -> 'Enabled' = 'true'
        and log_publishing_options -> 'ES_APPLICATION_LOGS' -> 'CloudWatchLogsLogGroupArn' is not null
      )
      and
      ( log_publishing_options -> 'SEARCH_SLOW_LOGS' -> 'Enabled' = 'true'
        and log_publishing_options -> 'SEARCH_SLOW_LOGS' -> 'CloudWatchLogsLogGroupArn' is not null
      )
      and
      ( log_publishing_options -> 'INDEX_SLOW_LOGS' -> 'Enabled' = 'true'
        and log_publishing_options -> 'INDEX_SLOW_LOGS' -> 'CloudWatchLogsLogGroupArn' is not null
      ) then title || ' logging enabled for search , index and error.'
    else title || ' logging not enabled for all search, index and error.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_elasticsearch_domain;