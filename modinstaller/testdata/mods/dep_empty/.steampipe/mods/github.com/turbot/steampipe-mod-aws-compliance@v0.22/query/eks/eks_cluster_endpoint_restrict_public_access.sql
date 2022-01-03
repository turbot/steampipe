select
  -- Required Columns
  arn as resource,
  case
    when resources_vpc_config ->> 'EndpointPublicAccess' = 'true' then 'alarm'
    else 'ok'
  end as status,
  case
    when resources_vpc_config ->> 'EndpointPublicAccess' = 'true' then title || ' endpoint publicly accessible.'
    else title || ' endpoint not publicly accessible.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_eks_cluster;