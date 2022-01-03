select
  -- Required Columns
  arn as resource,
  case
    when web_acl_arn is not null then 'ok'
    else 'alarm'
  end as status,
  case
    when web_acl_arn is not null then title || ' associated with WAF web ACL.'
    else title || ' not associated with WAF web ACL.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_api_gateway_stage;