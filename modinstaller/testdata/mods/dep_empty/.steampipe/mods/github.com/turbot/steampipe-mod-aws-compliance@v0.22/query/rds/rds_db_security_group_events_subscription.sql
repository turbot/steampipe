select
  -- Required Columns
  arn as resource,
  case
    when source_type <> 'db-security-group' then 'skip'
    when source_type = 'db-security-group' and enabled and event_categories_list @> '["failure", "configuration change"]' then 'ok'
    else 'alarm'
  end as status,
  case
    when source_type <> 'db-security-group' then cust_subscription_id || ' event subscription of ' || source_type || ' type.'
    when source_type = 'db-security-group' and enabled and event_categories_list @> '["failure", "configuration change"]' then cust_subscription_id || ' event subscription enabled for critical database security group events.'
    else cust_subscription_id || ' event subscription missing critical database security group events.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_event_subscription;