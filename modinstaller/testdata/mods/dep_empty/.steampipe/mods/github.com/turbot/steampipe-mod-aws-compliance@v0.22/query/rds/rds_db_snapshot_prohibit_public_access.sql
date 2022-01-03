(
select
  -- Required Columns
  arn as resource,
  case
    when cluster_snapshot -> 'AttributeValues' = '["all"]' then 'alarm'
    else 'ok'
  end status,
  case
    when cluster_snapshot -> 'AttributeValues' = '["all"]' then title || ' publicly restorable.'
    else title || ' not publicly restorable.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_cluster_snapshot,
  jsonb_array_elements(db_cluster_snapshot_attributes) as cluster_snapshot
)
union
(
select
  -- Required Columns
  arn as resource,
  case
    when database_snapshot -> 'AttributeValues' = '["all"]' then 'alarm'
    else 'ok'
  end status,
  case
    when database_snapshot -> 'AttributeValues' = '["all"]' then title || ' publicly restorable.'
    else title || ' not publicly restorable.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_rds_db_snapshot,
  jsonb_array_elements(db_snapshot_attributes) as database_snapshot
);
