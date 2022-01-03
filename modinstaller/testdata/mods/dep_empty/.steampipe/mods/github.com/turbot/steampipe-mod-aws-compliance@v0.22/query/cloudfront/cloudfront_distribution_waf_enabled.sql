select
  -- Required Columns
  arn as resource,
  case
    when web_acl_id <> '' then 'ok'
    else 'alarm'
  end as status,
  case
    when web_acl_id <> '' then title || ' associated with WAF.'
    else title || ' not associated with WAF.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_cloudfront_distribution;