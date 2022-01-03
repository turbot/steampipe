select
  -- Required Columns
  arn resource,
  case
    when ip_permissions is null and ip_permissions_egress is null then 'ok'
    else 'alarm'
  end status,
  case
    when ip_permissions is not null and ip_permissions_egress is not null
      then 'Default security group ' || group_id || ' has inbound and outbound rules.'
    when ip_permissions is not null and ip_permissions_egress is null
      then 'Default security group ' || group_id || ' has inbound rules.'
    when ip_permissions is null and ip_permissions_egress is not null
      then 'Default security group ' || group_id || ' has outbound rules.'
    else 'Default security group ' || group_id || ' has no inbound or outbound rules.'
  end reason,
  -- Additional Dimensions
  region,
  account_id
from
  aws_vpc_security_group
where
  group_name = 'default';
