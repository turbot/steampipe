select
  -- Required Columns
  arn as resource,
  case
    when source_type <> 'db-cluster' then 'skip'
    when source_type = 'db-cluster' and enabled and event_categories_list @> '["failure", "maintenance"]' then 'ok'
    else 'alarm'
  end as status,
  case
    when source_type <> 'db-cluster' then cust_subscription_id || ' event subscription of ' || source_type || ' type.'
    when source_type = 'db-cluster' and enabled and event_categories_list @> '["failure", "maintenance"]' then cust_subscription_id || ' event subscription enabled for critical db cluster events.'
    else cust_subscription_id || ' event subscription missing critical db cluster events.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_event_subscription;