select
  -- Required Columns
  arn as resource,
  case
    when source_type <> 'db-parameter-group' then 'skip'
    when source_type = 'db-parameter-group' and enabled and event_categories_list @> '["maintenance", "failure"]' then 'ok'
    else 'alarm'
  end as status,
  case
    when source_type <> 'db-parameter-group' then cust_subscription_id || ' event subscription of ' || source_type || ' type.'
    when source_type = 'db-parameter-group' and enabled and event_categories_list @> '["configuration change"]' then cust_subscription_id || ' event subscription enabled for critical database parameter group events.'
    else cust_subscription_id || ' event subscription missing critical database parameter group events.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_event_subscription;