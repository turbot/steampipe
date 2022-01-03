select
  -- Required Columns
  'arn' || ':' || 'aws' || ':::' || account_id as resource,
  case
    when block_public_acls
      and block_public_policy
      and ignore_public_acls
      and restrict_public_buckets
      then 'ok'
    else 'alarm'
  end as status,
  case
    when block_public_acls
      and block_public_policy
      and ignore_public_acls
      and restrict_public_buckets
      then 'Account level public access blocks enabled.'
    else 'Account level public access not enabled for: ' ||
      concat_ws(', ',
        case when not (block_public_acls ) then 'block_public_acls' end,
        case when not (block_public_policy) then 'block_public_policy' end,
        case when not (ignore_public_acls ) then 'ignore_public_acls' end,
        case when not (restrict_public_buckets) then 'restrict_public_buckets' end
      ) || '.'
  end as reason,
  -- Additional Dimensions
  account_id
from
  aws_s3_account_settings;