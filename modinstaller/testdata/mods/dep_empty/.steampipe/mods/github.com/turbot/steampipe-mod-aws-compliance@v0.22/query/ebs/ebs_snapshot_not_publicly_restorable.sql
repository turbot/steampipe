select
  -- Required Columns
  'arn:' || partition || ':ec2:' || region || ':' || account_id || ':snapshot/' || snapshot_id as resource,
  case
    when create_volume_permissions @> '[{"Group": "all", "UserId": null}]' then 'alarm'
    else 'ok'
  end status,
  case
    when create_volume_permissions @> '[{"Group": "all", "UserId": null}]' then title || ' is publicly restorable.'
    else title || ' is not publicly restorable.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_ebs_snapshot;