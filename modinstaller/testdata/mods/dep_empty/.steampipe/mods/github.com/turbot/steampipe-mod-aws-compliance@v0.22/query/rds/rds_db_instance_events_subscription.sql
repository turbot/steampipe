select
  -- Required Columns
  arn as resource,
  case
    when source_type <> 'db-instance' then 'skip'
    when source_type = 'db-instance' and enabled and event_categories_list @> '["failure", "maintenance", "configuration change"]' then 'ok'
    else 'alarm'
  end as status,
  case
    when source_type <> 'db-instance' then cust_subscription_id || ' event subscription of ' || source_type || ' type.'
    when source_type like 'db-instance' and enabled and event_categories_list @> '["failure", "maintenance", "configuration change"]' then cust_subscription_id || ' event subscription enabled for critical instance events.'
    else cust_subscription_id || ' event subscription missing critical instance events.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_event_subscription;