(
select
  -- Required Columns
  arn as resource,
  case
    when enabled_cloudwatch_logs_exports is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when enabled_cloudwatch_logs_exports is not null then title || ' enhanced monitoring enabled.'
    else title || ' enhanced monitoring not enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_cluster
)
union
(
select
  -- Required Columns
  arn as resource,
  case
    when class = 'db.m1.small' then 'skip'
    when enhanced_monitoring_resource_arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when class = 'db.m1.small' then title || ' enhanced monitoring not supported.'
    when enhanced_monitoring_resource_arn is not null then title || ' enhanced monitoring enabled.'
    else title || ' enhanced monitoring not enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_instance
);