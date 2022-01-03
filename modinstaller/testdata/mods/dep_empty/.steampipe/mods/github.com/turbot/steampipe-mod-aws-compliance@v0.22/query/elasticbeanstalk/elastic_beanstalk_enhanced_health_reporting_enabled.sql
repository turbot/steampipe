select
  -- Required Columns
  application_name as resource,
  case
    when health_status is not null and health is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when health_status is not null and health is not null then application_name || ' enhanced health check enabled.'
    else application_name || ' enhanced health check disabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_elastic_beanstalk_environment;
