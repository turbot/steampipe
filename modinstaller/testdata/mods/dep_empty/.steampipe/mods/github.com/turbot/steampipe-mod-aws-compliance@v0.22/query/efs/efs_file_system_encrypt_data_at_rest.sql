select
  -- Required Columns
  arn as resource,
  case
    when encrypted then 'ok'
    else 'alarm'
  end as status,
  case
    when encrypted then title || ' encrypted at rest.'
    else title || ' not encrypted at rest.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_efs_file_system;
