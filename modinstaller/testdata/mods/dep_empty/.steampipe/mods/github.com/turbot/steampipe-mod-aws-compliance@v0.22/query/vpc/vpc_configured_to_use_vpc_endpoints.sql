select
  -- Required Columns
  arn as resource,
  case
    when vpc_id not in (
      select
        vpc_id
      from
        aws_vpc_endpoint
      where
        service_name like 'com.amazonaws.' || region || '.ec2'
    ) then 'alarm'
    else 'ok'
  end as status,
  case
    when vpc_id not in (
      select
        vpc_id
      from
        aws_vpc_endpoint
      where
        service_name like 'com.amazonaws.' || region || '.ec2'
    ) then title || ' not configured to use VPC endpoints.'
    else title || ' configured to use VPC endpoints.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_vpc;