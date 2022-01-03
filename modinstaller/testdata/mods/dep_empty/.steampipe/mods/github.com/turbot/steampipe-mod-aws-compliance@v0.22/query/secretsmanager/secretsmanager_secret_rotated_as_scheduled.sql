select
  -- Required Columns
  arn as resource,
  case
    when primary_region is not null and region != primary_region then 'skip' -- Replica secret
    when rotation_rules is null then 'alarm' -- Rotation not enabled
    when last_rotated_date is null
      and (date(current_date) - date(created_date)) <= (rotation_rules -> 'AutomaticallyAfterDays')::integer then 'ok' -- New secret not due for rotation yet
    when last_rotated_date is null
      and (date(current_date) - date(created_date)) > (rotation_rules -> 'AutomaticallyAfterDays')::integer then 'alarm' -- New secret overdue for rotation
    when last_rotated_date is not null
     and (date(current_date) - date(last_rotated_date)) > (rotation_rules -> 'AutomaticallyAfterDays')::integer then 'alarm' -- Secret has been rotated before but is overdue for another rotation
  end as status,
  case
    when primary_region is not null and region != primary_region then title || ' is a replica.'
    when rotation_rules is null then title || ' rotation not enabled.'
    when last_rotated_date is null
      and (date(current_date) - date(created_date)) <= (rotation_rules -> 'AutomaticallyAfterDays')::integer then title || ' scheduled for rotation.'
    when last_rotated_date is null
     and (date(current_date) - date(created_date)) > (rotation_rules -> 'AutomaticallyAfterDays')::integer then title || ' not rotated as per schedule.'
    when last_rotated_date is not null
      and (date(current_date) - date(last_rotated_date)) > (rotation_rules -> 'AutomaticallyAfterDays')::integer then title || ' not rotated as per schedule.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_secretsmanager_secret;
