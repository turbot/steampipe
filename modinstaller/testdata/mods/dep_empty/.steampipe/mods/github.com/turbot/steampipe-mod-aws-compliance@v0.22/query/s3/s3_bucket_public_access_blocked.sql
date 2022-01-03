select
  -- Required Columns
  arn as resource,
  case
    when
      block_public_acls
      and block_public_policy
      and ignore_public_acls
      and restrict_public_buckets
    then
      'ok'
    else
      'alarm'
  end status,
  case
    when
      block_public_acls
      and block_public_policy
      and ignore_public_acls
      and restrict_public_buckets
    then name || ' blocks public access.'
    else name || ' does not block public access.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_s3_bucket