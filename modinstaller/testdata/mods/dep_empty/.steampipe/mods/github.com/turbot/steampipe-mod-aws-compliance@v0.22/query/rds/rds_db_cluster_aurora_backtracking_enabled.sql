select
  -- Required Columns
  arn as resource,
  case
    when engine not ilike '%aurora-mysql%' then 'skip'
    when backtrack_window is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when engine not ilike '%aurora-mysql%' then title || ' not Aurora MySQL-compatible edition.'
    when backtrack_window is not null then title || ' backtracking enabled.'
    else title || ' backtracking not enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from aws_rds_db_cluster;