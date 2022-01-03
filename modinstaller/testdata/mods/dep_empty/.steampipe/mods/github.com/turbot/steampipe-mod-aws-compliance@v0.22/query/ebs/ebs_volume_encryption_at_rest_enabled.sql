select
  -- Required Columns
  arn as resource,
  case
    when encrypted then 'ok'
    else 'alarm'
  end status,
  case
    when encrypted then volume_id || ' encrypted.'
    else volume_id || ' not encrypted.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ebs_volume;