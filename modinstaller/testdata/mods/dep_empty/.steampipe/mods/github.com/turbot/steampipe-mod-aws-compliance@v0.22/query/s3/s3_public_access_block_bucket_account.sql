select
  -- Required Columns
  arn as resource,
  case
    when (bucket.block_public_acls or s3account.block_public_acls)
      and (bucket.block_public_policy or s3account.block_public_policy)
      and (bucket.ignore_public_acls or s3account.ignore_public_acls)
      and (bucket.restrict_public_buckets or s3account.restrict_public_buckets)
      then 'ok'
    else 'alarm'
  end as status,
  case
    when (bucket.block_public_acls or s3account.block_public_acls)
      and (bucket.block_public_policy or s3account.block_public_policy)
      and (bucket.ignore_public_acls or s3account.ignore_public_acls)
      and (bucket.restrict_public_buckets or s3account.restrict_public_buckets)
      then name || ' all public access blocks enabled.'
    else name || ' not enabled for: ' ||
      concat_ws(', ',
        case when not (bucket.block_public_acls or s3account.block_public_acls) then 'block_public_acls' end,
        case when not (bucket.block_public_policy or s3account.block_public_policy) then 'block_public_policy' end,
        case when not (bucket.ignore_public_acls or s3account.ignore_public_acls) then 'ignore_public_acls' end,
        case when not (bucket.restrict_public_buckets or s3account.restrict_public_buckets) then 'restrict_public_buckets' end
      ) || '.'
  end as reason,
  -- Additional Dimensions
  bucket.region,
  bucket.account_id
from
  aws_s3_bucket as bucket,
  aws_s3_account_settings as s3account
where
  s3account.account_id = bucket.account_id;
