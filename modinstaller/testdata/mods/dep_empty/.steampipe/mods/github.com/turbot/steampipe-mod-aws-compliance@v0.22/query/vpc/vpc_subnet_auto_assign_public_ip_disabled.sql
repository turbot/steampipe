select
  -- Required Columns
  subnet_id as resource,
  case
    when map_public_ip_on_launch = 'false' then 'ok'
    else 'alarm'
  end as status,
  case
    when map_public_ip_on_launch = 'false' then title || ' auto assign public IP disabled.'
    else title || ' auto assign public IP enabled.'
  end as reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_vpc_subnet;